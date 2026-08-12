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
	msgText := nowPlayingText(chatID, t)

	opts := &td.SendTextMessageOpts{
		ParseMode:             "HTML",
		ReplyMarkup:           core.GetPlayMarkup(chatID, r, false),
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
					&td.EditMessageMediaOpts{ReplyMarkup: opts.ReplyMarkup},
				)
				if err == nil {
					return m
				}
				logger.Errorf("EditMessageMedia (now playing) failed: %v", err)
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
				ReplyMarkup: opts.ReplyMarkup,
			},
		)
		if err != nil {
			logger.Errorf("SendPhoto (now playing) failed: %v", err)
			m, _ := c.SendTextMessage(chatID, msgText, opts)
			return m
		}
		return photo
	}

	if statusMsg != nil {
		m, _ := utils.EOR(c, statusMsg, msgText, &td.EditTextMessageOpts{
			ParseMode:             "HTML",
			ReplyMarkup:           opts.ReplyMarkup,
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
		if blockedChat, _ := database.IsBlacklistedChat(cb.ChatId); blockedChat {
			if isOwnerOrSudo(cb.SenderUserId) {
				return handler(c, cb)
			}
			return nil
		}
		return handler(c, cb)
	}
}

func WithBlacklistMessage(
	handler func(*td.Client, *td.Message) error,
) func(*td.Client, *td.Message) error {
	return func(c *td.Client, m *td.Message) error {
		if blockedChat, _ := database.IsBlacklistedChat(m.ChatID()); blockedChat {
			if isOwnerOrSudo(m.SenderID()) {
				return handler(c, m)
			}
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

			mentioned := false
			if fields := strings.Fields(m.Text()); len(fields) > 0 {
				if _, mention, ok := strings.Cut(fields[0], "@"); ok && mention != "" {
					if c.Me == nil {
						if me, err := c.GetMe(); err == nil && me != nil {
							c.Me = me
						}
					}
					username := ""
					if c.Me != nil && c.Me.Usernames != nil && len(c.Me.Usernames.ActiveUsernames) > 0 {
						username = strings.ToLower(c.Me.Usernames.ActiveUsernames[0])
					}
					mentioned = strings.EqualFold(mention, username)
				}
			}
			if m.IsPrivate() || mentioned {
				if _, rerr := m.ReplyText(c, F(m.ChatID(), "blacklist_user_blocked"), nil); rerr != nil {
					logger.Error(rerr)
				}
			}
			return nil
		}
		return handler(c, m)
	}
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
