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
	"slices"
	"time"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/core"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func init() {
	helpTexts["/reload"] = `<i>Reload admin cache and refresh voice chat state.</i>

<u>Usage:</u>
<b>/reload</b> — Refresh all cached data

<b>🔄 What Gets Reloaded:</b>
• Chat admin list
• Voice chat status
• Assistant presence status
• Assistant ban status

<b>🔒 Flood Protection:</b>
• Regular users: 5 minute cooldown
• Admins: 2 minute cooldown

<b>💡 When to Use:</b>
• After promoting/demoting admins
• Voice chat issues
• Permission problems
• Bot behaving incorrectly

<b>⚠️ Notes:</b>
• May reset room state if admin permissions required
`
}

func reloadHandler(c *td.Client, m *td.Message) error {
	return handleReload(c, m, false)
}

func creloadHandler(c *td.Client, m *td.Message) error {
	return handleReload(c, m, true)
}

func handleReload(c *td.Client, m *td.Message, cplay bool) error {
	r, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		_, _ = m.ReplyText(c, err.Error(), nil)
		return nil
	}

	chatID := m.ChatID()
	roomID := r.ID

	if handled, err := checkReloadFlood(c, m, chatID, roomID); handled {
		return err
	}

	statusMsg, err := m.ReplyText(c, F(chatID, "reload_start"), nil)
	if err != nil {
		return err
	}

	summary := ""

	admins, adminSummary := reloadAdminCache(c, chatID, roomID)
	summary += adminSummary

	isAdmin := slices.Contains(admins, m.SenderID())
	floodDuration := utils.IfElse(isAdmin, 2*time.Minute, 5*time.Minute)
	utils.SetFlood(fmt.Sprintf("reload:%d%d", roomID, m.SenderID()), floodDuration)

	cs, err := core.GetChatState(roomID)
	if err != nil {
		summary += F(chatID, "reload_assistant_fail", locales.Arg{
			"error": err.Error(),
		}) + "\n"
		_, _ = utils.EOR(c, statusMsg, F(chatID, "reload_done", locales.Arg{
			"summary": summary,
		}), &td.EditTextMessageOpts{ParseMode: td.ParseModeHTML})
		return nil
	}

	snapshot, snapErr := cs.Snapshot(true)
	summary += reloadVoiceChatStatus(chatID, snapshot, snapErr)
	summary += reloadAssistantStatus(chatID, snapshot, snapErr)

	if isAdmin {
		if core.DeleteRoom(roomID) {
			summary += F(chatID, "reload_room_reset") + "\n"
		}
	}

	_, _ = utils.EOR(c, statusMsg, F(chatID, "reload_done", locales.Arg{
		"summary": summary,
	}), &td.EditTextMessageOpts{ParseMode: td.ParseModeHTML})

	return nil
}

func checkReloadFlood(c *td.Client, m *td.Message, chatID, roomID int64) (bool, error) {
	floodKey := fmt.Sprintf("reload:%d%d", roomID, m.SenderID())
	if remaining := utils.GetFlood(floodKey); remaining > 0 {
		_, err := m.ReplyText(c, F(
			chatID,
			"flood_minutes",
			locales.Arg{
				"duration": utils.FormatDuration(int(remaining.Seconds())),
			},
		), nil)
		return true, err
	}
	return false, nil
}

func reloadAdminCache(c *td.Client, chatID, roomID int64) ([]int64, string) {
	admins, err := utils.RefreshChatAdmin(c, roomID)
	if err != nil {
		return nil, F(chatID, "reload_admin_cache_fail", locales.Arg{
			"error": err.Error(),
		}) + "\n"
	}
	return admins, F(chatID, "reload_admin_cache_ok") + "\n"
}

func reloadVoiceChatStatus(chatID int64, snap core.StateSnapshot, err error) string {
	if err != nil {
		if errors.Is(err, core.ErrAdminPermissionRequired) {
			return F(chatID, "reload_voice_admin_required") + "\n"
		}
		return F(chatID, "reload_voice_fail", locales.Arg{
			"error": err.Error(),
		}) + "\n"
	}

	if snap.VoiceChatActive {
		return F(chatID, "reload_voice_active") + "\n"
	}
	return F(chatID, "reload_voice_inactive") + "\n"
}

func reloadAssistantStatus(chatID int64, snap core.StateSnapshot, err error) string {
	if err != nil {
		if errors.Is(err, core.ErrAdminPermissionRequired) {
			return F(chatID, "reload_assistant_admin_required") + "\n"
		}
		return F(chatID, "reload_assistant_fail", locales.Arg{
			"error": err.Error(),
		}) + "\n"
	}

	if snap.AssistantBanned {
		return F(chatID, "reload_assistant_banned") + "\n"
	}

	if snap.AssistantPresent {
		return F(chatID, "reload_assistant_present") + "\n"
	}
	return F(chatID, "reload_assistant_not_present") + "\n"
}
