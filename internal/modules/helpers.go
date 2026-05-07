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
	"runtime/debug"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
	state "main/internal/core/models"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

var downloadCancels = make(map[int64]func())

func getEffectiveRoom(m *tg.NewMessage, cplay bool) (*core.RoomState, error) {
	chatID := m.ChannelID()

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
		r.ChatID = m.ChannelID()
	}
	return r, nil
}

func canBypassMaintenence(userID int64) bool {
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

func F(chatID int64, key string, values ...locales.Arg) string {
	lang, err := database.Language(chatID)
	if err != nil {
		gologging.ErrorF("failed to get language for %d: %v", chatID, err)
		lang = config.DefaultLang
	}
	return FWithLang(lang, key, values...)
}

func FWithLang(lang, key string, values ...locales.Arg) string {
	var val locales.Arg
	if len(values) > 0 {
		val = values[0]
	}
	return locales.Get(lang, key, val)
}

func isLoggerEnabled() bool {
	l, err := database.IsLoggerEnabled()
	if err != nil {
		gologging.ErrorF("failed to check if logger is enabled: %v", err)
		return false
	}
	return l
}

func sendPlayLogs(m *tg.NewMessage, track *state.Track, queued bool) {
	if config.LoggerID == 0 || config.LoggerID == m.ChatID() ||
		config.LoggerID == m.ChannelID() || m.SenderID() == config.OwnerID ||
		!isLoggerEnabled() {
		return
	}

	header := F(m.ChannelID(), "logger_playback_started")
	if queued {
		header = F(m.ChannelID(), "logger_playback_queued")
	}

	var sb strings.Builder
	sb.WriteString("🎵 ")
	if m.Channel.Username != "" {
		fmt.Fprintf(&sb, "<b><a href=\"%s\">%s</a></b>\n\n", m.Link(), header)
	} else {
		fmt.Fprintf(&sb, "<b><u>%s</u></b>\n\n", header)
	}

	groupName := m.Channel.Title
	if m.Channel.Username != "" {
		groupName = "@" + m.Channel.Username
	}

	requestedBy := utils.MentionHTML(m.Sender)
	if m.Sender.Username != "" {
		requestedBy = "@" + m.Sender.Username
	}

	if track.Artwork != "" {
		sb.WriteString("<blockquote>")
	}

	sb.WriteString(F(m.ChannelID(), "logger_playback_template", locales.Arg{
		"track_url":       track.URL,
		"track":           utils.EscapeHTML(utils.ShortTitle(track.Title)),
		"source":          string(track.Source),
		"group":           groupName,
		"group_id":        m.ChannelID(),
		"requested_by":    requestedBy,
		"requested_by_id": m.SenderID(),
	}))

	if track.Artwork != "" {
		sb.WriteString("\n</blockquote>")
	}

	text := sb.String()
	opts := &tg.SendOptions{ParseMode: "HTML"}
	if shouldShowThumb(config.LoggerID) && track.Artwork != "" {
		opts.Media = utils.CleanURL(track.Artwork)
	}

	_, err := core.Bot.SendMessage(config.LoggerID, text, opts)
	if err != nil {
		gologging.Error("failed to send logger msg: " + err.Error())
	}
}

func SafeCallbackHandler(
	handler func(*tg.CallbackQuery) error,
) func(*tg.CallbackQuery) error {
	return func(cb *tg.CallbackQuery) (err error) {
		if !canBypassMaintenence(cb.SenderID) {
			cb.Answer(
				F(cb.ChannelID(), "maint", locales.Arg{"reason": ""}),
				&tg.CallbackOptions{Alert: true},
			)
			return tg.ErrEndGroup
		}

		defer func() {
			if r := recover(); r != nil {
				handlePanic(r, cb, true)
				err = tg.ErrEndGroup
			}
		}()

		err = handler(cb)
		if err != nil && !errors.Is(err, tg.ErrEndGroup) {
			handlePanic(err, cb, false)
		}
		return err
	}
}

func SafeMessageHandler(
	handler func(*tg.NewMessage) error,
) func(*tg.NewMessage) error {
	return func(m *tg.NewMessage) (err error) {
		gologging.InfoF(
			"Handling message from %d in chat %d",
			m.SenderID(),
			m.ChannelID(),
		)

		if !canBypassMaintenence(m.SenderID()) {
			gologging.Debug("Maintenance mode active")
			if m.ChatType() == tg.EntityUser ||
				strings.HasSuffix(m.GetCommand(), m.Client.Me().Username) {
				reason, _ := database.MaintenanceReason()
				msg := F(m.ChannelID(), "maint", locales.Arg{
					"reason": F(
						m.ChannelID(),
						"maint_reason",
						locales.Arg{"reason": reason},
					),
				})
				m.Reply(msg)
				gologging.InfoF(
					"Sent maintenance notice to %d",
					m.SenderID(),
				)
			}
			return tg.ErrEndGroup
		}

		defer func() {
			if r := recover(); r != nil {
				gologging.ErrorF("recovered from panic: %v", r)
				handlePanic(r, m, true)
				err = fmt.Errorf("internal panic occurred")
			}
		}()

		if m.IsCommand() {
			isEnabled, _ := database.CommandDelete(m.ChannelID())
			if isEnabled {
				_, _ = m.Delete()
			} else {
				cleanMode, _ := database.CleanMode(m.ChannelID())
				if cleanMode {
					cleanScheduler.schedule(m.ChannelID(), m.ID)
				}
			}
		}

		cmd := getCommand(m)
		gologging.DebugF("Executing handler for command %s", cmd)
		err = handler(m)

		if err != nil {
			if errors.Is(err, tg.ErrEndGroup) {
				gologging.Debug("handler exited early (ErrEndGroup)")
				return err
			}
			gologging.ErrorF("handler error: %v", err)
			handlePanic(err, m, false)
		} else {
			gologging.InfoF("handler completed successfully for command %s", cmd)
		}

		return err
	}
}

func blacklistMessageMiddleware(next tg.MessageHandler) tg.MessageHandler {
	return func(m *tg.NewMessage) error {
		if blockedChat, _ := database.IsBlacklistedChat(m.ChannelID()); blockedChat {
			if isOwnerOrSudo(m.SenderID()) {
				return next(m)
			}
			m.Reply(F(m.ChannelID(), "blacklist_chat_blocked"))
			leaveChat(m.Client, m.ChannelID())
			return tg.ErrEndGroup
		}
		if blocked, _ := database.IsBlacklistedUser(m.SenderID()); blocked {
			if m.IsChannel() {
				chatOwnerID, err := utils.GetChatOwner(m.Client, m.ChannelID())
				if err == nil && chatOwnerID == m.SenderID() {
					m.Reply(F(m.ChannelID(), "blacklist_owner_blocked_leave"))
					leaveChat(m.Client, m.ChannelID())
					return tg.ErrEndGroup
				}
			}
			if m.IsPrivate() || strings.HasSuffix(m.GetCommand(), m.Client.Me().Username) {
				m.Reply(F(m.ChannelID(), "blacklist_user_blocked"))
			}
			return tg.ErrEndGroup
		}
		return next(m)
	}
}

func WithBlacklistCallback(
	handler func(*tg.CallbackQuery) error,
) func(*tg.CallbackQuery) error {
	return func(cb *tg.CallbackQuery) error {
		if blocked, _ := database.IsBlacklistedUser(cb.SenderID); blocked {
			return tg.ErrEndGroup
		}
		if blockedChat, _ := database.IsBlacklistedChat(cb.ChannelID()); blockedChat {
			if isOwnerOrSudo(cb.SenderID) {
				return handler(cb)
			}
			return tg.ErrEndGroup
		}
		return handler(cb)
	}
}

type panicInfo struct {
	userMention  string
	handlerType  string
	chatInfo     string
	messageInfo  string
	errorMessage string
	client       *tg.Client
}

func getPanicInfo(ctx, r any) panicInfo {
	var info panicInfo
	info.errorMessage = html.EscapeString(fmt.Sprint(r))

	switch c := ctx.(type) {
	case *tg.NewMessage:
		info.userMention = utils.MentionHTML(c.Sender)
		info.handlerType = "message"
		info.chatInfo = "ChatID: " + utils.IntToStr(c.ChannelID())
		info.messageInfo = "Message: " + html.EscapeString(c.Text()) + "\nLink: " + c.Link()
		info.client = c.Client

	case *tg.CallbackQuery:
		info.userMention = utils.MentionHTML(c.Sender)
		info.handlerType = "callback"
		info.chatInfo = "ChatID: " + utils.IntToStr(c.ChatID)
		info.messageInfo = "Data: " + html.EscapeString(c.DataString())
		info.client = c.Client
	}
	return info
}

func handlePanic(r, ctx any, isPanic bool) {
	info := getPanicInfo(ctx, r)
	stackRaw := debug.Stack()
	stackEsc := html.EscapeString(string(stackRaw))

	logPrefix := "🚨 Error"
	shortPrefix := "<b>Error</b>"
	if isPanic {
		logPrefix = "⚠️ Panic recovered"
		shortPrefix = "<b>⚠️ Panic</b>"
	}

	logMsg := fmt.Sprintf("%s in %s handler:\nFrom: %s\n%s\n%s\nError: `%v`",
		logPrefix, info.handlerType, info.userMention, info.chatInfo, info.messageInfo, r)
	shortMsg := fmt.Sprintf("%s in %s handler\n<b>From:</b> %s\n%s\n%s\n<b>Error:</b>\n<code>%s</code>",
		shortPrefix, info.handlerType, info.userMention, info.chatInfo, info.messageInfo, info.errorMessage)

	if isPanic {
		logMsg += "\nStack:\n" + string(stackRaw)
		shortMsg += "\n<pre>" + stackEsc + "</pre>"
	}

	gologging.Error(logMsg)

	if config.LoggerID != 0 && info.client != nil {
		if _, sendErr := info.client.SendMessage(config.LoggerID, shortMsg, &tg.SendOptions{ParseMode: "HTML"}); sendErr != nil {
			gologging.ErrorF(
				"failed to send panic message to log chat: %v",
				sendErr,
			)
		}
	}
}

func warnAndLeave(client *tg.Client, chatID int64) {
	text := F(chatID, "supergroup_needed", locales.Arg{"chat_id": chatID, "support_group": config.SupportChat})
	_, err := client.SendMessage(chatID, text, &tg.SendOptions{
		ReplyMarkup: core.AddMeMarkup(chatID),
		LinkPreview: false,
	})
	if err != nil {
		gologging.ErrorF("failed to send supergroup conversion message to chat %d: %v", chatID, err)
		return
	}

	go func() {
		leaveChat(client, chatID)
	}()
}

func leaveChat(client *tg.Client, chatID int64) {
	go func() {
		time.Sleep(1 * time.Second)
		if err := client.LeaveChannel(chatID); err != nil {
			gologging.ErrorF("failed to leave blacklisted chatID=%d: %v", chatID, err)
		}
		core.Assistants.WithAssistant(
			chatID,
			func(ass *core.Assistant) { ass.Client.LeaveChannel(chatID) },
		)
	}()
}

func getCommand(m *tg.NewMessage) string {
	cmd, _, _ := strings.Cut(m.GetCommand(), "@")
	return cmd
}
