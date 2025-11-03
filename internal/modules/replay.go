/*
 * This file is part of YukkiMusic.
 *
 * YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
 * Copyright (C) 2025 TheTeamVivek
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 */
package modules

import (
	"fmt"
	"html"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/utils"
)

func replayHandler(m *telegram.NewMessage) error {
	return handleReplay(m, false)
}

func creplayHandler(m *telegram.NewMessage) error {
	return handleReplay(m, true)
}

func handleReplay(m *telegram.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.EndGroup
	}

	if !r.IsActiveChat() {
		m.Reply("⚠️ <b>No active playback.</b>\nNothing is playing right now.")
		return telegram.EndGroup
	}

	if err := r.Replay(); err != nil {
		m.Reply(fmt.Sprintf("❌ <b>Replay Failed</b>\nError: <code>%v</code>", err))
	} else {
		trackTitle := html.EscapeString(utils.ShortTitle(r.Track.Title, 25))
		totalDuration := formatDuration(r.Track.Duration)
		m.Reply(fmt.Sprintf("🔁 Now replaying:\n\n<b>Title: </b>%s\n🎵 Duration: <code>%s</code>\n⏱️ Speed: %.2fx", trackTitle, totalDuration, r.Speed))
	}

	return telegram.EndGroup
}
