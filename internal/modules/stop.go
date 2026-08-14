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
	"time"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/core"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

const stopConfirmSuggestionCooldown = 4 * time.Second

func init() {
	helpTexts["/end"] = `<i>Stop playback and leave the voice chat.</i>

<u>Usage:</u>
<b>/stop</b> or <b>/end</b> — Stop playback

<b>⚙️ Behavior:</b>
• Stops current track
• Clears queue
• Assistant leaves voice chat
•
<b>🔒 Restrictions:</b>
• Only <b>chat admins</b> or <b>authorized users</b> can use this

<b>⚠️ Note:</b>
This action cannot be undone. Use <code>/pause</code> for temporary stops.`
	helpTexts["/stop"] = helpTexts["/end"]
}

func stopHandler(c *td.Client, m *td.Message) error {
	if !isSuperGroup(c, m) || !filterAuthUsers(c, m) {
		return nil
	}
	return handleStop(c, m, false)
}

func cstopHandler(c *td.Client, m *td.Message) error {
	if !isSuperGroup(c, m) || !filterAuthUsers(c, m) {
		return nil
	}
	return handleStop(c, m, true)
}

func handleStop(c *td.Client, m *td.Message, cplay bool) error {
	r, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		_, err := m.ReplyText(c, err.Error(), nil)
		return err
	}
	if !r.Active() {
		_, err := m.ReplyText(c, F(m.ChatID(), "room_no_active"), nil)
		return err
	}

	isPaused := r.Paused()
	isMuted := r.Muted()

	if isPaused || isMuted {
		stopSuggestFloodKey := fmt.Sprintf(
			"stop_suggest:%d",
			r.ID,
		)
		if utils.GetFlood(stopSuggestFloodKey) <= 0 {
			utils.SetFlood(stopSuggestFloodKey, stopConfirmSuggestionCooldown)
			msgKey := "stop_confirm_paused"
			if isMuted {
				msgKey = "stop_confirm_muted"
			}
			_, err := m.ReplyText(c, F(m.ChatID(), msgKey), &td.SendTextMessageOpts{
				ReplyMarkup: core.GetStopConfirmMarkup(m.ChatID(), r, isPaused),
			})
			return err
		}
	}

	scheduleOldPlayingMessage(r)
	core.DropRoom(r.ID)
	_, rerr := m.ReplyText(
		c,
		F(
			m.ChatID(),
			"stopped",
			locales.Arg{"user": mentionOf(nil, m.SenderID())},
		),
		nil,
	)
	return rerr
}
