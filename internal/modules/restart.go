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
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	td "github.com/AshokShau/gotdbot"
	"yukkimusic/internal/logger"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func init() {
	helpTexts["/restart"] = `<i>Restart the bot process.</i>

<u>Usage:</u>
<b>/restart</b> — Restart bot

<b>⚙️ Behavior:</b>
• Stops all active rooms
• Notifies all active chats
• Restarts bot process
• Clears download cache

<b>🔒 Restrictions:</b>
• <b>Owner only</b> command

<b>⚠️ Warning:</b>
All playback will be interrupted. Bot will be offline for a few seconds.`
}

func handleRestart(c *td.Client, m *td.Message) error {
	chatID := m.ChatID()
	r, ok := getActiveRoomForChat(chatID)
	if ok && r.Track() != nil {
		_, _ = m.ReplyText(c, F(chatID, "restart_confirm_running"), &td.SendTextMessageOpts{
			ReplyMarkup: core.GetRestartConfirmMarkup(chatID),
		})
		return nil
	}
	return performRestart(c, m, chatID)
}

func performRestart(c *td.Client, m *td.Message, chatID int64) error {
	statusMsg, err := m.ReplyText(c, F(chatID, "restart"), nil)
	if err != nil {
		logger.Error("Failed to send restart message: " + err.Error())
	}
	return executeRestart(c, chatID, statusMsg)
}

func getActiveRoomForChat(chatID int64) (*core.RoomState, bool) {
	r, ok := core.GetRoom(chatID, nil, false)
	if !ok || !r.IsActiveChat() || r.Track() == nil {
		return nil, false
	}
	return r, true
}

func restartConfirmHandler(c *td.Client, u *td.UpdateNewCallbackQuery) error {
	chatID := u.ChatId

	if u.SenderUserId != config.OwnerID {
		_ = u.Answer(c, 0, true, F(chatID, "only_owner"), "")
		return nil
	}

	action := strings.TrimPrefix(u.DataString(), "restart:")
	switch action {
	case "bot":
		_ = u.Answer(c, 0, false, F(chatID, "restart_confirm_bot"), "")
		statusMsg, _ := u.EditMessageText(c, F(chatID, "restart"), nil)
		return executeRestart(c, chatID, statusMsg)
	case "replay":
		r, ok := getActiveRoomForChat(chatID)
		if !ok {
			_ = u.Answer(c, 0, true, F(chatID, "room_no_active"), "")
			return nil
		}
		if err := r.Replay(); err != nil {
			_ = u.Answer(c, 0, true, F(chatID, "replay_failed", locales.Arg{"error": err}), "")
			return nil
		}
		_, _ = u.EditMessageText(c, F(chatID, "restart_confirm_replay_done"), nil)
		_ = u.Answer(c, 0, true, F(chatID, "restart_confirm_replay_done"), "")
	}

	return nil
}

func executeRestart(
	c *td.Client,
	chatID int64,
	statusMsg *td.Message,
) error {
	exePath, err := os.Executable()
	if err != nil {
		utils.EOR(statusMsg, F(chatID, "restart_exepath_fail", locales.Arg{
			"error": err.Error(),
		}), nil)
		return nil
	}

	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		utils.EOR(statusMsg, F(chatID, "restart_symlink_fail", locales.Arg{
			"error": err.Error(),
		}), nil)
		return nil
	}

	for roomChatID := range core.GetAllRooms() {
		core.DeleteRoom(roomChatID)
		_, _ = c.SendTextMessage(roomChatID, F(roomChatID, "restart_service", locales.Arg{
			"bot": mentionOf(c.Me, c.Me.Id),
		}), nil)
		time.Sleep(time.Second)
	}

	utils.EOR(statusMsg, F(chatID, "restart_initiated"), nil)

	_ = os.RemoveAll("downloads")
	_ = os.RemoveAll("cache")

	if err := syscall.Exec(exePath, os.Args, os.Environ()); err != nil {
		utils.EOR(statusMsg, F(chatID, "restart_fail", locales.Arg{
			"error": err.Error(),
		}), nil)
	}

	return nil
}
