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
	"strings"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/core"
	"yukkimusic/internal/database"
	"yukkimusic/internal/utils"
)

func init() {
	helpTexts["/maintenance"] = `<i>Toggle maintenance mode.</i>

<u>Usage:</u>
<b>/maintenance</b> — Open the maintenance panel
<b>/maintenance enable</b> — Enable maintenance
<b>/maintenance disable</b> — Disable maintenance

<b>⚙️ Behavior When Active:</b>
• Stops all active rooms
• Blocks non-owner/sudo commands
• Shows maintenance message to users

<b>🔒 Restrictions:</b>
• <b>Owner only</b> command`
	helpTexts["/maint"] = helpTexts["/maintenance"]

	dashboards["maintenance"] = dashboardDef{
		title:       func(chatID int64) string { return F(chatID, "maint_title") },
		description: func(chatID int64) string { return F(chatID, "maint_desc") },
		canUse: func(c *td.Client, u *td.UpdateNewCallbackQuery) bool {
			if isOwner(u.SenderUserId) {
				return true
			}
			u.Answer(c, 0, true, F(u.ChatId, "only_owner_cb"), "")
			return false
		},
		get: func(int64) (bool, error) { return database.GetMaintenance() },
		set: func(_ int64, on bool) error { return applyMaintenance(on) },
	}
}

func handleMaintenance(c *td.Client, m *td.Message) error {
	if !checkOwner(c, m) {
		return nil
	}

	if args := strings.Fields(m.Text()); len(args) > 1 {
		if enabled, err := utils.ParseBool(args[1]); err == nil {
			if current, cerr := database.GetMaintenance(); cerr == nil && enabled != current {
				if err := applyMaintenance(enabled); err != nil {
					return err
				}
			}
		}
	}

	return sendDashboard(c, m, "maintenance")
}

func applyMaintenance(on bool) error {
	if err := database.SetMaintenance(on); err != nil {
		return err
	}
	if on {
		go stopAllRooms()
	}
	return nil
}

// stopAllRooms destroys every active room.
func stopAllRooms() {
	for chatID := range core.Rooms.All() {
		core.Rooms.Delete(chatID)
	}
}
