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
	helpTexts["/mute"] = `<i>Mute the audio output in voice chat.</i>

<u>Usage:</u>
<b>/mute</b> — Mute indefinitely

<b>⚙️ Features:</b>
• Audio continues playing (progress tracked)

<b>💡 Examples:</b>
<code>/mute</code> — Mute until manual unmute

<b>⚠️ Notes:</b>
• Track continues playing in background
• Use <code>/unmute</code> to restore audio`
}

func muteHandler(c *td.Client, m *td.Message) error {
	return handleMute(c, m, false)
}

func cmuteHandler(c *td.Client, m *td.Message) error {
	return handleMute(c, m, true)
}

func handleMute(c *td.Client, m *td.Message, cplay bool) error {
	if m.Args() != "" || m.ReplyToMessageID() != 0 || !isSuperGroup(c, m) || !filterAuthUsers(c, m) {
		return nil
	}

	chatID := m.ChatID()
	r, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		_, err := m.ReplyText(c, err.Error(), nil)
		return err
	}

	if !r.Active() {
		_, err := m.ReplyText(c, F(chatID, "room_no_active"), nil)
		return err
	}

	if r.Muted() {
		_, err := m.ReplyRichMessageHTML(c, F(chatID, "mute_already_muted"), &td.SendRichTextOpts{
			ReplyMarkup: core.GetMuteToggle(chatID, r),
		})
		return err
	}

	sender, _ := m.GetUser(c)
	mention := mention(sender, m.SenderID())

	muteErr := r.Mute()

	if muteErr != nil {
		_, err := m.ReplyRichMessageHTML(c, F(chatID, "mute_failed", locales.Arg{
			"error": utils.EscapeHTML(muteErr.Error()),
		}), nil)
		return err
	}
	speedLine := ""
	if sp := r.Speed(); sp != 1.0 {
		speedLine = `<tr><td colspan="2">` + F(chatID, "speed_line", locales.Arg{
			"speed": fmt.Sprintf("%.2f", sp),
		}) + `</td></tr>`
	}

	_, rerr := m.ReplyRichMessageHTML(c, F(chatID, "mute_success", locales.Arg{
		"title":      utils.EscapeHTML(utils.ShortTitle(r.Track().Title, 25)),
		"url":        utils.EscapeHTML(r.Track().URL),
		"user":       mention,
		"speed_line": speedLine,
	}), &td.SendRichTextOpts{
		ReplyMarkup: core.GetMuteToggle(chatID, r),
	})
	if rerr != nil {
		return rerr
	}

	updatePlaybackMessageFor(c, r.StatusMsg(), m.SenderID(), r, "muted")
	return nil
}
