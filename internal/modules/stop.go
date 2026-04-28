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

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	"main/internal/locales"
	"main/internal/utils"
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

func stopHandler(m *telegram.NewMessage) error {
	return handleStop(m, false)
}

func cstopHandler(m *telegram.NewMessage) error {
	return handleStop(m, true)
}

func handleStop(m *telegram.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.ErrEndGroup
	}
	if !r.IsActiveChat() {
		m.Reply(F(m.ChannelID(), "room_no_active"))
		return telegram.ErrEndGroup
	}

	isPaused := r.IsPaused()
	isMuted := r.IsMuted()

	if isPaused || isMuted {
		stopSuggestFloodKey := fmt.Sprintf(
			"stop_suggest:%d",
			r.ID(),
		)
		if utils.GetFlood(stopSuggestFloodKey) <= 0 {
			utils.SetFlood(stopSuggestFloodKey, stopConfirmSuggestionCooldown)
			msgKey := "stop_confirm_paused"
			if isMuted {
				msgKey = "stop_confirm_muted"
			}
			m.Reply(F(m.ChannelID(), msgKey), &telegram.SendOptions{
				ReplyMarkup: core.GetStopConfirmMarkup(m.ChannelID(), r, isPaused),
			})
			return telegram.ErrEndGroup
		}
	}

	core.DeleteRoom(r.ID())
	m.Reply(
		F(
			m.ChannelID(),
			"stopped",
			locales.Arg{"user": utils.MentionHTML(m.Sender)},
		),
	)
	return telegram.ErrEndGroup
}
