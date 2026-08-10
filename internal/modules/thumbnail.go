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
	"yukkimusic/internal/locales"
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
}

func nothumbHandler(c *td.Client, m *td.Message) error {
	if !isSuperGroup(c, m) || !filterAuthUsers(c, m) {
		return nil
	}
	chatID := m.ChatID()
	args := strings.Fields(m.Text())

	current, err := database.ThumbnailsDisabled(chatID)
	if err != nil {
		_, err := m.ReplyText(c, F(chatID, "nothumb_fetch_fail"), nil)
		return err
	}

	if len(args) < 2 {
		action := utils.IfElse(!current, "enabled", "disabled")
		_, err := m.ReplyText(c, F(chatID, "nothumb_status", locales.Arg{
			"cmd":    getCommand(m),
			"action": action,
		}), nil)
		return err
	}

	value, err := utils.ParseBool(args[1])
	if err != nil {
		_, err := m.ReplyText(c, F(chatID, "invalid_bool"), nil)
		return err
	}

	if current == value {
		action := utils.IfElse(!value, "enabled", "disabled")
		_, err := m.ReplyText(c, F(chatID, "nothumb_already", locales.Arg{
			"action": action,
		}), nil)
		return err
	}

	if err := database.SetThumbnailsDisabled(chatID, value); err != nil {
		_, err := m.ReplyText(c, F(chatID, "nothumb_update_fail"), nil)
		return err
	}

	action := utils.IfElse(!value, "enabled", "disabled")
	_, rerr := m.ReplyText(c, F(chatID, "nothumb_updated", locales.Arg{
		"action": action,
	}), nil)
	return rerr
}
