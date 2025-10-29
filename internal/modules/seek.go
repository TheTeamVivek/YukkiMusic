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
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func seekHandler(m *telegram.NewMessage) error {
	return handleSeek(m, false, false)
}

func cseekHandler(m *telegram.NewMessage) error {
	return handleSeek(m, true, false)
}

func seekbackHandler(m *telegram.NewMessage) error {
	return handleSeek(m, false, true)
}

func cseekbackHandler(m *telegram.NewMessage) error {
	return handleSeek(m, true, true)
}

func handleSeek(m *telegram.NewMessage, cplay, isBack bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.EndGroup
	}
	if !r.IsActiveChat() {
		m.Reply("⚠️ <b>No track is currently playing.</b>")
		return telegram.EndGroup
	}
	args := strings.Fields(m.Text())
	if len(args) < 2 {
		m.Reply(fmt.Sprintf("⚠️ Please provide seconds. Example: <code>%s 40</code>", getCommand(m)))
		return telegram.EndGroup
	}
	seconds, err := strconv.Atoi(args[1])
	if err != nil {
		m.Reply(fmt.Sprintf("⚠️ Invalid seconds value. Example: <code>%s 40</code>", getCommand(m)))
		return telegram.EndGroup
	}

	var direction, emoji string
	var seekErr error

	if isBack {
		if (r.Position - seconds) <= 10 {
			m.Reply(fmt.Sprintf(
				"⚠️ Cannot seek backward %d seconds — that would be too close to the beginning of the track.",
				seconds,
			))
			return telegram.EndGroup
		}
		seekErr = r.Seek(-seconds)
		direction = "backward"
		emoji = "⏪"
	} else {
		if (r.Track.Duration - (r.Position + seconds)) <= 10 {
			m.Reply(fmt.Sprintf(
				"⚠️ Cannot seek forward %d seconds — that would be too close to the end of the track.",
				seconds,
			))
			return telegram.EndGroup
		}
		seekErr = r.Seek(seconds)
		direction = "forward"
		emoji = "⏩"
	}

	if seekErr != nil {
		m.Reply(fmt.Sprintf("❌ Failed to seek %s %d seconds.\nError: %v", direction, seconds, seekErr))
	} else {
		m.Reply(fmt.Sprintf(
			"%s Jumped %s to <u>%s</u> of <u>%s</u>.",
			emoji,
			direction,
			formatDuration(r.Position),
			formatDuration(r.Track.Duration),
		))
	}
	return telegram.EndGroup
}

func jumpHandler(m *telegram.NewMessage) error {
	return handleJump(m, false)
}

func cjumpHandler(m *telegram.NewMessage) error {
	return handleJump(m, true)
}

func handleJump(m *telegram.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.EndGroup
	}
	if !r.IsActiveChat() || r.Track == nil {
		m.Reply("⚠️ <b>No active track to jump within.</b>")
		return telegram.EndGroup
	}
	args := strings.Fields(m.Text())
	if len(args) < 2 {
		m.Reply(fmt.Sprintf("⚠️ Please provide seconds. Example: <code>%s 120</code>", getCommand(m)))
		return telegram.EndGroup
	}
	seconds, err := strconv.Atoi(args[1])
	if err != nil || seconds < 0 {
		m.Reply(fmt.Sprintf("⚠️ Invalid position. Example: <code>%s 120</code>", getCommand(m)))
		return telegram.EndGroup
	}
	if r.Track.Duration-seconds <= 10 {
		m.Reply(fmt.Sprintf(
			"⚠️ Cannot jump to %s — that’s too close to the end of the track.",
			formatDuration(seconds),
		))
		return telegram.EndGroup
	}
	if err := r.Seek(seconds - r.Position); err != nil {
		m.Reply(fmt.Sprintf("❌ Failed to jump to %d sec.\nError: %v", seconds, err))
	} else {
		m.Reply(fmt.Sprintf(
			"⏩ Jumped to <u>%s</u> of <u>%s</u>.",
			formatDuration(seconds),
			formatDuration(r.Track.Duration),
		))
	}
	return telegram.EndGroup
}
