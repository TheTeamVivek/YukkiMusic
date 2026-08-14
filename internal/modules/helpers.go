/*
 * ● YukkiMusic
 * ○ A high-performance engine for streaming music in Telegram voicechats.
 *
 * Copyright (C) 2026 TheTeamVivek
 *
 * This program is free software: you can redistribute it and/or modify it under the
 * terms of the GNU General Public License as published by the Free Software Foundation,
 * either version 3 of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
 * PARTICULAR PURPOSE. See the GNU General Public License for more details.
 *
 * Repository: https://github.com/TheTeamVivek/YukkiMusic
 */

package modules

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"yukkimusic/internal/logger"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	state "yukkimusic/internal/core/models"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func getEffectiveRoom(chatID int64, cplay bool) (*core.RoomState, error) {
	origChatID := chatID

	if cplay {
		cplayID, err := database.LinkedChannel(chatID)
		if err != nil || cplayID == 0 {
			return nil, errors.New(F(chatID, "cplay_id_not_set"))
		}
		chatID = cplayID
	}
	ass, err := core.Assistants.ForChat(chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assistant for your chat: %w", err)
	}
	r, _ := core.GetRoom(chatID, ass, true)

	if cplay {
		r.ChatID = origChatID
	}
	return r, nil
}

func shouldShowThumb(chatID int64) bool {
	noThumb, err := database.ThumbnailsDisabled(chatID)
	if err != nil {
		// On error, default to showing thumbnails
		return true
	}
	// ThumbnailsDisabled = true means DON'T show thumb
	// So we return the inverse
	return !noThumb
}

// sendNowPlaying sends or edits the "now playing" playback message for the
// given track, honoring the thumbnail/artwork setting, and returns the
// resulting message.
// nowPlayingText builds the "now playing" status text for a track.
func nowPlayingText(chatID int64, t *state.Track) string {
	return F(chatID, "stream_now_playing", locales.Arg{
		"url":      t.URL,
		"title":    utils.EscapeHTML(utils.ShortTitle(t.Title, 25)),
		"duration": utils.FormatDuration(t.Duration),
		"by":       t.Requester,
	})
}

func sendNowPlaying(
	c *td.Client,
	statusMsg *td.Message,
	chatID int64,
	r *core.RoomState,
	t *state.Track,
) *td.Message {
	return showTrackMessage(c, statusMsg, chatID, r, t, nowPlayingText(chatID, t), false)
}

// showTrackMessage sends or edits a track message with optional artwork,
// honoring the thumbnail/artwork setting for the chat, and returns the
// resulting message.
func showTrackMessage(
	c *td.Client,
	statusMsg *td.Message,
	chatID int64,
	r *core.RoomState,
	t *state.Track,
	msgText string,
	queued bool,
) *td.Message {
	markup := core.GetPlayMarkup(chatID, r, queued)

	opts := &td.SendTextMessageOpts{
		ParseMode:             "HTML",
		ReplyMarkup:           markup,
		DisableWebPagePreview: true,
	}

	if t.Artwork != "" && shouldShowThumb(chatID) {
		content := &td.InputMessagePhoto{
			Photo: &td.InputPhoto{
				Photo: td.InputFileRemote{Id: utils.CleanURL(t.Artwork)},
			},
		}
		if statusMsg != nil {
			if caption, err := c.GetFormattedText(msgText, nil, "HTML"); err == nil {
				content.Caption = caption
				m, err := c.EditMessageMedia(
					chatID,
					content,
					statusMsg.Id,
					&td.EditMessageMediaOpts{ReplyMarkup: markup},
				)
				if err == nil {
					return m
				}
				logger.Errorf("EditMessageMedia (track message) failed: %v", err)
			}
		}
		if statusMsg != nil {
			_ = statusMsg.Delete(c, true)
		}
		photo, err := c.SendPhoto(
			chatID,
			td.InputFileRemote{Id: utils.CleanURL(t.Artwork)},
			&td.SendPhotoOpts{
				Caption:     msgText,
				ParseMode:   "HTML",
				ReplyMarkup: markup,
			},
		)
		if err != nil {
			logger.Errorf("SendPhoto (track message) failed: %v", err)
			m, _ := c.SendTextMessage(chatID, msgText, opts)
			return m
		}
		return photo
	}

	if statusMsg != nil {
		m, _ := utils.EOR(c, statusMsg, msgText, &td.EditTextMessageOpts{
			ParseMode:             "HTML",
			ReplyMarkup:           markup,
			DisableWebPagePreview: true,
		})
		return m
	}
	m, _ := c.SendTextMessage(chatID, msgText, opts)
	return m
}

func F(chatID int64, key string, values ...locales.Arg) string {
	lang, err := database.Language(chatID)
	if err != nil {
		logger.Errorf("failed to get language for %d: %v", chatID, err)
		lang = config.DefaultLang
	}
	var val locales.Arg
	if len(values) > 0 {
		val = values[0]
	}
	return locales.Get(lang, key, val)
}

func isLoggerEnabled() bool {
	l, err := database.IsLoggerEnabled()
	if err != nil {
		logger.Errorf("failed to check if logger is enabled: %v", err)
		return false
	}
	return l
}

func sendPlayLogs(c *td.Client, m *td.Message, track *state.Track, queued bool) {
	if config.LoggerID == 0 || config.LoggerID == m.ChatID() ||
		m.SenderID() == config.OwnerID || !isLoggerEnabled() {
		return
	}

	header := F(m.ChatID(), "logger_playback_started")
	if queued {
		header = F(m.ChatID(), "logger_playback_queued")
	}

	chat, _ := m.GetChat(c)
	groupName := "N/A"
	groupLink := ""
	if chat != nil {
		groupName = chat.Title
		if l, err := m.GetLink(c); err == nil && l != nil && l.IsPublic {
			groupLink = l.Link
		}
	}

	var sb strings.Builder
	sb.WriteString("🎵 ")
	if groupLink != "" {
		fmt.Fprintf(&sb, "<b><a href=\"%s\">%s</a></b>\n\n", groupLink, header)
	} else {
		fmt.Fprintf(&sb, "<b><u>%s</u></b>\n\n", header)
	}

	sender, _ := m.GetUser(c)
	requestedBy := mentionOf(sender, m.SenderID())

	sb.WriteString(F(m.ChatID(), "logger_playback_template", locales.Arg{
		"track_url":       track.URL,
		"track":           utils.EscapeHTML(utils.ShortTitle(track.Title)),
		"source":          string(track.Source),
		"group":           groupName,
		"group_id":        m.ChatID(),
		"requested_by":    requestedBy,
		"requested_by_id": m.SenderID(),
	}))

	_, err := core.Bot.SendTextMessage(
		config.LoggerID,
		sb.String(),
		&td.SendTextMessageOpts{ParseMode: "HTML", DisableWebPagePreview: true},
	)
	if err != nil {
		logger.Error("failed to send logger msg: " + err.Error())
	}
}

func WithBlacklistCallback(
	handler func(*td.Client, *td.UpdateNewCallbackQuery) error,
) func(*td.Client, *td.UpdateNewCallbackQuery) error {
	return func(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
		if blocked, _ := database.IsBlacklistedUser(cb.SenderUserId); blocked {
			return nil
		}
		if blockedChat, _ := database.IsBlacklistedChat(cb.ChatId); blockedChat &&
			!isOwnerOrSudo(cb.SenderUserId) {
			return nil
		}

		if !canBypassMaintenance(cb.SenderUserId) {
			cb.Answer(c, 0, true, F(cb.ChatId, "maint", locales.Arg{"reason": ""}), "")
			return td.EndGroups
		}

		return handler(c, cb)
	}
}

func WithBlacklistMessage(
	handler func(*td.Client, *td.Message) error,
) func(*td.Client, *td.Message) error {
	return func(c *td.Client, m *td.Message) error {
		if blockedChat, _ := database.IsBlacklistedChat(m.ChatID()); blockedChat &&
			!isOwnerOrSudo(m.SenderID()) {
			if _, err := m.ReplyText(c, F(m.ChatID(), "blacklist_chat_blocked"), nil); err != nil {
				logger.Error(err)
			}
			leaveChat(c, m.ChatID())
			return nil
		}
		if blocked, _ := database.IsBlacklistedUser(m.SenderID()); blocked {
			if chat, err := m.GetChat(c); err == nil {
				if sg, ok := chat.Type.(*td.ChatTypeSupergroup); ok && sg.IsChannel {
					if chatOwnerID, err := utils.GetChatOwner(c, m.ChatID()); err == nil && chatOwnerID == m.SenderID() {
						if _, rerr := m.ReplyText(c, F(m.ChatID(), "blacklist_owner_blocked_leave"), nil); rerr != nil {
							logger.Error(rerr)
						}
						leaveChat(c, m.ChatID())
						return nil
					}
				}
			}

			if m.IsPrivate() || messageMentionsBot(c, m) {
				if _, rerr := m.ReplyText(c, F(m.ChatID(), "blacklist_user_blocked"), nil); rerr != nil {
					logger.Error(rerr)
				}
			}
			return nil
		}

		if !canBypassMaintenance(m.SenderID()) {
			if m.IsPrivate() || messageMentionsBot(c, m) {
				reason, _ := database.MaintenanceReason()
				msg := F(m.ChatID(), "maint", locales.Arg{
					"reason": F(
						m.ChatID(),
						"maint_reason",
						locales.Arg{"reason": reason},
					),
				})
				if _, rerr := m.ReplyText(c, msg, nil); rerr != nil {
					logger.Error(rerr)
				}
			}
			return td.EndGroups
		}

		err := handler(c, m)

		if m.IsCommand() {
			if isEnabled, _ := database.CommandDelete(m.ChatID()); isEnabled {
				if err := m.Delete(c, true); err != nil {
					logger.Debugf("failed to delete command message: %v", err)
				}
			} else {
				cleanMode, _ := database.CleanMode(m.ChatID())
				if cleanMode {
					cleanScheduler.schedule(m.ChatID(), m.Id)
				}
			}
		}

		return err
	}
}

// canBypassMaintenance reports whether the user is allowed to use the bot
// while maintenance mode is active (owner or sudo users).
func canBypassMaintenance(userID int64) bool {
	isMaint, _ := database.IsMaintenanceEnabled()
	if !isMaint {
		return true
	}
	if userID == config.OwnerID {
		return true
	}
	ok, _ := database.IsSudo(userID)
	return ok
}

// messageMentionsBot reports whether the first word of the message mentions
// the bot by username (e.g. "/cmd@BotUsername").
func messageMentionsBot(c *td.Client, m *td.Message) bool {
	fields := strings.Fields(m.Text())
	if len(fields) == 0 {
		return false
	}
	_, mention, ok := strings.Cut(fields[0], "@")
	if !ok || mention == "" {
		return false
	}
	username := ""
	if c.Me != nil && c.Me.Usernames != nil && len(c.Me.Usernames.ActiveUsernames) > 0 {
		username = strings.ToLower(c.Me.Usernames.ActiveUsernames[0])
	}
	return username != "" && strings.EqualFold(mention, username)
}

func leaveChat(c *td.Client, chatID int64) {
	go func() {
		time.Sleep(1 * time.Second)
		if err := c.LeaveChat(chatID); err != nil {
			logger.Errorf("failed to leave blacklisted chatID=%d: %v", chatID, err)
		}
		core.Assistants.WithAssistant(
			chatID,
			func(ass *core.Assistant) { ass.Client.LeaveChannel(chatID) },
		)
	}()
}
