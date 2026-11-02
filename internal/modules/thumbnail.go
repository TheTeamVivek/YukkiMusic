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

	"yukkimusic/internal/database"
	"yukkimusic/internal/utils"
)

func init() {
	helpTexts["/nothumb"] = `<i>Toggle thumbnail/artwork display in playback messages.</i>

<u>Usage:</u>
<b>/nothumb</b> — Show current status
<b>/nothumb [enable|disable]</b> — Change setting

<b>⚙️ Behavior:</b>
• <b>Disabled (default):</b> Shows track artwork/thumbnail
• <b>Enabled:</b> Hides artwork, text-only messages

<b>💡 Examples:</b>
<code>/nothumb enable</code> — Disable thumbnails
<code>/nothumb disable</code> — Enable thumbnails

<b>⚠️ Note:</b>
This setting affects all future playback messages in this chat.`
	dashboards["nothumb"] = dashboardDef{
		title:       func(chatID int64) string { return F(chatID, "nothumb_title") },
		description: func(chatID int64) string { return F(chatID, "nothumb_desc") },
		canUse:      canUseAdminCallback,
		get:         func(chatID int64) (bool, error) { return database.ThumbnailsDisabled(chatID) },
		set:         func(chatID int64, on bool) error { return database.SetThumbnailsDisabled(chatID, on) },
	}
}

func nothumbHandler(c *td.Client, m *td.Message) error {
	if !isSuperGroup(c, m) || !filterAuthUsers(c, m) {
		return nil
	}
	chatID := m.ChatID()

	if args := strings.Fields(m.Text()); len(args) > 1 {
		if enabled, err := utils.ParseBool(args[1]); err == nil {
			if current, cerr := database.ThumbnailsDisabled(chatID); cerr == nil && enabled != current {
				if err := database.SetThumbnailsDisabled(chatID, enabled); err != nil {
					return err
				}
			}
		}
	}

	return sendDashboard(c, m, "nothumb")
}
