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
	"fmt"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/core"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func init() {
	helpTexts["/unmute"] = `<i>Unmute the audio output in voice chat.</i>

<u>Usage:</u>
<b>/unmute</b> — Restore audio

<b>⚙️ Behavior:</b>
• Restores audio immediately
• Shows current playback info`
}

func unmuteHandler(c *td.Client, m *td.Message) error {
	return handleUnmute(c, m, false)
}

func cunmuteHandler(c *td.Client, m *td.Message) error {
	return handleUnmute(c, m, true)
}

func handleUnmute(c *td.Client, m *td.Message, cplay bool) error {
	if m.Args() != "" || m.ReplyToMessageID() != 0 || !isSuperGroup(c, m) || !filterAuthUsers(c, m) {
		return nil
	}

	r, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		_, err := m.ReplyText(c, err.Error(), nil)
		return err
	}

	chatID := m.ChatID()

	if !r.Active() {
		_, err := m.ReplyText(c, F(chatID, "room_no_active"), nil)
		return err
	}

	if !r.Muted() {
		_, err := m.ReplyRichMessageHTML(c, F(chatID, "unmute_already"), &td.SendRichTextOpts{
			ReplyMarkup: core.GetMuteToggle(chatID, r),
		})
		return err
	}

	title := utils.EscapeHTML(utils.ShortTitle(r.Track().Title, 25))
	sender, _ := m.GetUser(c)
	mention := mention(sender, m.SenderID())

	if err := r.Unmute(); err != nil {
		_, rerr := m.ReplyRichMessageHTML(c, F(chatID, "unmute_failed", locales.Arg{
			"error": utils.EscapeHTML(err.Error()),
		}), nil)
		return rerr
	}

	speedLine := ""
	if sp := r.Speed(); sp != 1.0 {
		speedLine = `<tr><td colspan="2">` + F(chatID, "speed_line", locales.Arg{
			"speed": fmt.Sprintf("%.2f", sp),
		}) + `</td></tr>`
	}

	msg := F(chatID, "unmute_success", locales.Arg{
		"title":      title,
		"url":        utils.EscapeHTML(r.Track().URL),
		"user":       mention,
		"speed_line": speedLine,
	})
	_, rerr := m.ReplyRichMessageHTML(c, msg, &td.SendRichTextOpts{
		ReplyMarkup: core.GetMuteToggle(chatID, r),
	})
	if rerr != nil {
		return rerr
	}

	updatePlaybackMessageFor(c, r.StatusMsg(), m.SenderID(), r, "unmuted")
	return nil
}
