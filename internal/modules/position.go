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
	helpTexts["/position"] = `<i>Show current playback position and track info.</i>

<u>Usage:</u>
<b>/position</b> — Show position
<b>/pos</b> — Alias for /position

<b>📊 Information Displayed:</b>
• Current track title & link
• Current playback position & duration
• Playback speed
• Requester`
}

func positionHandler(c *td.Client, m *td.Message) error {
	return handlePosition(c, m, false)
}

func cpositionHandler(c *td.Client, m *td.Message) error {
	return handlePosition(c, m, true)
}

func handlePosition(c *td.Client, m *td.Message, cplay bool) error {
	if !isSuperGroup(c, m) {
		return nil
	}

	if cplay && !filterAuthUsers(c, m) {
		return nil
	}

	chatID := m.ChatID()

	r, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		_, err := m.ReplyText(c, err.Error(), nil)
		return err
	}

	if !r.Active() || r.Track() == nil || r.Track().ID == "" {
		_, err := m.ReplyText(c, F(chatID, "room_no_active"), nil)
		return err
	}

	track := r.Track()
	dur := track.Duration
	pos := r.Position()
	if pos < 0 {
		pos = 0
	}

	posStr, durStr, remStr := "N/A", "N/A", "N/A"
	if dur > 0 {
		if pos > dur {
			pos = dur
		}
		posStr = utils.FormatTime(pos)
		durStr = utils.FormatTime(dur)
		remSec := max(dur-pos, 0)
		remStr = "-" + utils.FormatTime(remSec)
	}

	safeTitle := utils.EscapeHTML(utils.ShortTitle(track.Title, 40))
	var trackLink string
	if track.URL != "" {
		trackLink = `<b><a href="` + utils.EscapeHTML(track.URL) + `">` + safeTitle + `</a></b>`
	} else {
		trackLink = `<b>` + safeTitle + `</b>`
	}

	requester := track.Requester
	if requester == "" {
		requester = "Unknown"
	}

	speed := r.Speed()
	var speedStr string
	if speed == float64(int(speed)) {
		speedStr = fmt.Sprintf("%.1f", speed)
	} else {
		speedStr = fmt.Sprintf("%.2f", speed)
	}

	msg := F(chatID, "position_now", locales.Arg{
		"track":     trackLink,
		"requester": requester,
		"speed":     speedStr,
		"position":  posStr,
		"duration":  durStr,
		"remaining": remStr,
	})

	_, rerr := m.ReplyRichMessageHTML(c, msg, &td.SendRichTextOpts{
		ReplyMarkup: core.SuppMarkup(chatID),
	})
	return rerr
}
