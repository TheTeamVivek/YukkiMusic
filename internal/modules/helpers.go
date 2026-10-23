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
	"html"
	"strings"
	"time"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	state "yukkimusic/internal/core/models"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func getCommand(m *td.Message) string {
	parts := strings.Fields(m.Text())
	if len(parts) == 0 {
		return ""
	}
	cmd, _, _ := strings.Cut(parts[0], "@")
	return cmd
}

func checkSudo(c *td.Client, m *td.Message) bool {
	return allowed(c, m, isOwnerOrSudo(m.SenderID()), "only_sudo")
}

func checkOwner(c *td.Client, m *td.Message) bool {
	return allowed(c, m, isOwner(m.SenderID()), "only_owner")
}

// allowed returns ok and, when it is false, tells the sender why with the
// given locale key.
func allowed(c *td.Client, m *td.Message, ok bool, key string) bool {
	if !ok {
		replyKey(c, m, key)
	}
	return ok
}

// replyKey replies with a localized text and logs a failure to send it.
func replyKey(c *td.Client, m *td.Message, key string) {
	if _, err := m.ReplyText(c, F(m.ChatID(), key), nil); err != nil {
		utils.Log.Error(err.Error())
	}
}

func isOwner(userID int64) bool {
	return config.OwnerID != 0 && userID == config.OwnerID
}

func isSuperGroup(c *td.Client, m *td.Message) bool {
	if m.IsPrivate() {
		replyKey(c, m, "only_supergroup")
		database.AddServedUser(m.ChatID())
		return false
	}

	chat, err := m.GetChat(c)
	if err != nil {
		return false
	}

	sg, ok := chat.Type.(*td.ChatTypeSupergroup)
	if !ok || sg.IsChannel {
		return false
	}

	database.AddServedChat(m.ChatID())
	return true
}

func filterAuthUsers(c *td.Client, m *td.Message) bool {
	if isAdminOrAuth(c, m.ChatID(), m.SenderID()) {
		return true
	}
	if _, err := m.ReplyRichMessageHTML(c, F(m.ChatID(), adminDeniedKey(m.ChatID())), nil); err != nil {
		utils.Log.Error(err.Error())
	}
	return false
}

// adminDeniedKey picks the refusal text: "admins only" when the chat is in
// that mode, otherwise "admins or authorized users".
func adminDeniedKey(chatID int64) string {
	if mode, err := database.GetAdminMode(chatID); err == nil && mode == database.AdminModeAdminsOnly {
		return "only_admin"
	}
	return "only_admin_or_auth"
}

func isAdminOrAuth(c *td.Client, chatID, userID int64) bool {
	if isOwnerOrSudo(userID) {
		return true
	}

	mode, err := database.GetAdminMode(chatID)
	if err != nil {
		mode = database.AdminModeAdminAuth
	}

	if mode == database.AdminModeEveryone {
		return true
	}

	isAdmin, err := utils.IsChatAdmin(c, chatID, userID)
	if err == nil && isAdmin {
		return true
	}

	if mode == database.AdminModeAdminsOnly {
		return false
	}

	isAuth, err := database.IsAuthorized(chatID, userID)
	return err == nil && isAuth
}

// canUseAdminCallback is the callback-query counterpart of filterAuthUsers: it
// checks admin command access and, when denied, shows the proper toast (plain
// text, since callback answers do not render HTML).
func canUseAdminCallback(c *td.Client, u *td.UpdateNewCallbackQuery) bool {
	if isAdminOrAuth(c, u.ChatId, u.SenderUserId) {
		return true
	}
	u.Answer(c, F(u.ChatId, adminDeniedKey(u.ChatId)+"_cb"), &td.CallbackAnswerOpts{ShowAlert: true})
	return false
}

func canManageAuth(c *td.Client, chatID, userID int64) bool {
	if isOwnerOrSudo(userID) {
		return true
	}

	isAdmin, err := utils.IsChatAdmin(c, chatID, userID)
	return err == nil && isAdmin
}

func isOwnerOrSudo(userID int64) bool {
	if isOwner(userID) {
		return true
	}
	isSudo, err := database.IsSudo(userID)
	return err == nil && isSudo
}

func getEffectiveRoom(chatID int64, cplay bool) (*core.Room, error) {
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
	r := core.Rooms.GetOrCreate(chatID, ass)

	if cplay {
		r.ControlChatID = origChatID
	}
	return r, nil
}

func activeRoomFor(c *td.Client, m *td.Message, cplay bool) (*core.Room, error) {
	if !isSuperGroup(c, m) {
		return nil, nil
	}

	if !filterAuthUsers(c, m) {
		return nil, nil
	}

	chatID := m.ChatID()

	r, err := getEffectiveRoom(chatID, cplay)
	if err != nil {
		_, rerr := m.ReplyText(c, err.Error(), nil)
		return nil, rerr
	}

	if !r.Active() {
		_, rerr := m.ReplyText(c, F(chatID, "room_no_active"), nil)
		return nil, rerr
	}

	return r, nil
}

// nowPlayingText builds the "now playing" status text for a track.
func nowPlayingText(chatID int64, t *state.Track) string {
	return F(chatID, "stream_now_playing", locales.Arg{
		"url":      t.URL,
		"title":    utils.EscapeHTML(utils.ShortTitle(t.Title, 25)),
		"duration": utils.FormatDuration(t.Duration),
		"by":       t.Requester,
	})
}

// sendNowPlaying sends or edits the "now playing" message for a track and
// returns the resulting message.
func sendNowPlaying(
	c *td.Client,
	statusMsg *td.Message,
	chatID int64,
	r *core.Room,
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
	r *core.Room,
	t *state.Track,
	msgText string,
	queued bool,
) *td.Message {
	markup := core.GetPlayMarkup(chatID, r, queued)

	opts := &td.SendMessageOpts{
		ParseMode:             "HTML",
		ReplyMarkup:           markup,
		DisableWebPagePreview: true,
	}

	if noThumb, _ := database.ThumbnailsDisabled(chatID); t.Artwork != "" && !noThumb {
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
				utils.Log.Errorf("EditMessageMedia (track message) failed: %v", err)
			}
		}
		if statusMsg != nil {
			_ = statusMsg.Delete(c, &td.DeleteMessagesOpts{Revoke: true})
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
			utils.Log.Errorf("SendPhoto (track message) failed: %v", err)
			m, _ := c.SendMessage(chatID, msgText, opts)
			return m
		}
		return photo
	}

	if statusMsg != nil {
		m, _ := utils.EOR(c, statusMsg, msgText, &td.EditMessageTextOpts{
			ParseMode:             "HTML",
			ReplyMarkup:           markup,
			DisableWebPagePreview: true,
		})
		return m
	}
	m, _ := c.SendMessage(chatID, msgText, opts)
	return m
}

func F(chatID int64, key string, values ...locales.Arg) string {
	lang, err := database.Language(chatID)
	if err != nil {
		utils.Log.Errorf("failed to get language for %d: %v", chatID, err)
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
		utils.Log.Errorf("failed to check if logger is enabled: %v", err)
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
	requestedBy := mention(sender, m.SenderID())

	sb.WriteString(F(m.ChatID(), "logger_playback_template", locales.Arg{
		"track_url":       track.URL,
		"track":           utils.EscapeHTML(utils.ShortTitle(track.Title)),
		"source":          string(track.Source),
		"group":           groupName,
		"group_id":        m.ChatID(),
		"requested_by":    requestedBy,
		"requested_by_id": m.SenderID(),
	}))

	_, err := core.Bot.SendMessage(
		config.LoggerID,
		sb.String(),
		&td.SendMessageOpts{ParseMode: "HTML", DisableWebPagePreview: true},
	)
	if err != nil {
		utils.Log.Error("failed to send logger msg: " + err.Error())
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
			cb.Answer(c, F(cb.ChatId, "maint"), &td.CallbackAnswerOpts{ShowAlert: true})
			return td.EndGroups
		}

		return handler(c, cb)
	}
}

func WithBlacklistMessage(
	handler func(*td.Client, *td.Message) error,
) func(*td.Client, *td.Message) error {
	return func(c *td.Client, m *td.Message) error {
		chatID, userID := m.ChatID(), m.SenderID()

		if blocked, _ := database.IsBlacklistedChat(chatID); blocked && !isOwnerOrSudo(userID) {
			replyKey(c, m, "blacklist_chat_blocked")
			leaveChat(c, chatID)
			return nil
		}

		if blocked, _ := database.IsBlacklistedUser(userID); blocked {
			if isChannel(c, m) {
				if owner, err := utils.GetChatOwner(c, chatID); err == nil && owner == userID {
					replyKey(c, m, "blacklist_owner_blocked_leave")
					leaveChat(c, chatID)
					return nil
				}
			}
			if isAddressed(c, m) {
				replyKey(c, m, "blacklist_user_blocked")
			}
			return nil
		}

		if !canBypassMaintenance(userID) {
			if isAddressed(c, m) {
				replyKey(c, m, "maint")
			}
			return td.EndGroups
		}

		err := handler(c, m)

		if m.IsCommand() {
			if del, _ := database.CommandDelete(chatID); del {
				if err := m.Delete(c, &td.DeleteMessagesOpts{Revoke: true}); err != nil {
					utils.Log.Debugf("failed to delete command message: %v", err)
				}
			} else {
				cleaner.Add(chatID, m.Id) // a no-op unless clean mode is on
			}
		}

		return err
	}
}

// canBypassMaintenance reports whether the user may use the bot while
// maintenance mode is active (owner or sudo users).
func canBypassMaintenance(userID int64) bool {
	maintenance, _ := database.GetMaintenance()
	return !maintenance || isOwnerOrSudo(userID)
}

// isAddressed reports whether a message is meant for this bot: it is sent in
// private, or its first word mentions the bot (e.g. "/cmd@BotUsername").
func isAddressed(c *td.Client, m *td.Message) bool {
	if m.IsPrivate() {
		return true
	}
	fields := strings.Fields(m.Text())
	me := c.Me()
	if len(fields) == 0 || me == nil || me.Usernames == nil ||
		len(me.Usernames.ActiveUsernames) == 0 {
		return false
	}
	_, name, ok := strings.Cut(fields[0], "@")
	return ok && name != "" && strings.EqualFold(name, me.Usernames.ActiveUsernames[0])
}

func leaveChat(c *td.Client, chatID int64) {
	go func() {
		time.Sleep(1 * time.Second)
		if err := c.LeaveChat(chatID); err != nil {
			utils.Log.Errorf("failed to leave blacklisted chatID=%d: %v", chatID, err)
		}
		core.Assistants.WithAssistant(
			chatID,
			func(ass *core.Assistant) { ass.Client.LeaveChannel(chatID) },
		)
	}()
}

// isChannel reports whether m came from a broadcast channel.
func isChannel(c *td.Client, m *td.Message) bool {
	chat, err := m.GetChat(c)
	if err != nil {
		return false
	}
	sg, ok := chat.Type.(*td.ChatTypeSupergroup)
	return ok && sg.IsChannel
}

// mention builds an HTML mention for a gotdbot user, falling back to a plain
// "user" mention when the user is nil or has no display name.
func mention(user *td.User, fallbackId int64) string {
	if user == nil {
		return mentionHTML("user", fallbackId)
	}
	if strings.TrimSpace(user.FirstName+" "+user.LastName) == "" {
		return mentionHTML("user", user.Id)
	}
	return user.Mention(td.ParseModeHTML)
}

// mentionHTML builds an HTML user mention from a display name and id. It is
// used for senders that are not represented by a gotdbot *User.
func mentionHTML(name string, id int64) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "user"
	}
	return fmt.Sprintf(`<a href="tg://user?id=%d">%s</a>`, id, html.EscapeString(name))
}
