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
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	td "github.com/AshokShau/gotdbot"
	"yukkimusic/internal/logger"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	state "yukkimusic/internal/core/models"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

var downloads = &downloadManager{
	cancels: make(map[int64]context.CancelFunc),
}

type downloadManager struct {
	mu      sync.RWMutex
	cancels map[int64]context.CancelFunc
}

func (dm *downloadManager) Add(chatID int64, cancel context.CancelFunc) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.cancels[chatID] = cancel
}

func (dm *downloadManager) Remove(chatID int64) bool {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if cancel, ok := dm.cancels[chatID]; ok {
		cancel()
		delete(dm.cancels, chatID)
		return true
	}
	return false
}

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
		logger.Errorf("failed to get language for %d: %v", chatID, err)
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

	_, err := core.TDBot.SendTextMessage(
		config.LoggerID,
		sb.String(),
		&td.SendTextMessageOpts{ParseMode: "HTML"},
	)
	if err != nil {
		logger.Error("failed to send logger msg: " + err.Error())
	}
}

func WithBlacklistCallback(
	handler func(*td.UpdateNewCallbackQuery) error,
) func(*td.UpdateNewCallbackQuery) error {
	return func(cb *td.UpdateNewCallbackQuery) error {
		if blocked, _ := database.IsBlacklistedUser(cb.SenderUserId); blocked {
			return nil
		}
		if blockedChat, _ := database.IsBlacklistedChat(cb.ChatId); blockedChat {
			if isOwnerOrSudo(cb.SenderUserId) {
				return handler(cb)
			}
			return nil
		}
		return handler(cb)
	}
}

func warnAndLeave(c *td.Client, chatID int64) {
	text := F(chatID, "supergroup_needed", locales.Arg{
		"chat_id":       chatID,
		"support_group": config.SupportChat,
	})

	if _, err := c.SendTextMessage(chatID, text, nil); err != nil {
		logger.Errorf("failed to send supergroup conversion message to chat %d: %v", chatID, err)
		return
	}

	go func() {
		leaveChat(c, chatID)
	}()
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
