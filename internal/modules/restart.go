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

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
	"main/internal/locales"
	"main/internal/utils"
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

func handleRestart(m *tg.NewMessage) error {
	chatID := m.ChannelID()
	r, ok := getActiveRoomForChat(chatID)
	if ok && r.Track() != nil {
		_, _ = m.Reply(F(chatID, "restart_confirm_running"), &tg.SendOptions{
			ReplyMarkup: core.GetRestartConfirmMarkup(chatID),
		})
		return tg.ErrEndGroup
	}
	return performRestart(m, chatID)
}

func performRestart(m *tg.NewMessage, chatID int64) error {
	statusMsg, err := m.Reply(F(chatID, "restart"))
	if err != nil {
		gologging.Error("Failed to send restart message: " + err.Error())
	}
	return executeRestart(m.Client, chatID, statusMsg)
}

func getActiveRoomForChat(chatID int64) (*core.RoomState, bool) {
	r, ok := core.GetRoom(chatID, nil, false)
	if !ok || !r.IsActiveChat() || r.Track() == nil {
		return nil, false
	}
	return r, true
}

func restartConfirmHandler(cb *tg.CallbackQuery) error {
	chatID := cb.ChannelID()
	opt := &tg.CallbackOptions{Alert: true}

	if cb.SenderID != config.OwnerID {
		cb.Answer(F(chatID, "only_owner"), opt)
		return tg.ErrEndGroup
	}

	action := strings.TrimPrefix(cb.DataString(), "restart:")
	switch action {
	case "bot":
		cb.Answer(F(chatID, "restart_confirm_bot"))
		statusMsg, _ := cb.Edit(F(chatID, "restart"))
		return executeRestart(cb.Client, chatID, statusMsg)
	case "replay":
		r, ok := getActiveRoomForChat(chatID)
		if !ok {
			cb.Answer(F(chatID, "room_no_active"), opt)
			return tg.ErrEndGroup
		}
		if err := r.Replay(); err != nil {
			cb.Answer(F(chatID, "replay_failed", locales.Arg{"error": err}), opt)
			return tg.ErrEndGroup
		}
		_, _ = cb.Edit(F(chatID, "restart_confirm_replay_done"))
		cb.Answer(F(chatID, "restart_confirm_replay_done"))
	}

	return tg.ErrEndGroup
}

func executeRestart(
	client *tg.Client,
	chatID int64,
	statusMsg *tg.NewMessage,
) error {
	exePath, err := os.Executable()
	if err != nil {
		utils.EOR(statusMsg, F(chatID, "restart_exepath_fail", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		utils.EOR(statusMsg, F(chatID, "restart_symlink_fail", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	for roomChatID := range core.GetAllRooms() {
		core.DeleteRoom(roomChatID)
		client.SendMessage(roomChatID, F(roomChatID, "restart_service", locales.Arg{
			"bot": utils.MentionHTML(client.Me()),
		}))
		time.Sleep(time.Second)
	}

	utils.EOR(statusMsg, F(chatID, "restart_initiated"))

	_ = os.RemoveAll("downloads")
	_ = os.RemoveAll("cache")

	if err := syscall.Exec(exePath, os.Args, os.Environ()); err != nil {
		utils.EOR(statusMsg, F(chatID, "restart_fail", locales.Arg{
			"error": err.Error(),
		}))
	}

	return tg.ErrEndGroup
}
