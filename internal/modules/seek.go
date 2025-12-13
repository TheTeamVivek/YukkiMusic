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
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/locales"
)

func init() {
	helpTexts["/seek"] = `<i>Seek forward in the currently playing track.</i>

<u>Usage:</u>
<b>/seek [seconds]</b> — Skip forward by specified seconds

<b>⚙️ Features:</b>
• Jump ahead in current track
• Position tracking updated
• Cannot seek past track end (10s buffer)

<b>🔒 Restrictions:</b>
• Only <b>chat admins</b> or <b>authorized users</b> can use this

<b>💡 Examples:</b>
<code>/seek 30</code> — Skip forward 30 seconds
<code>/seek 120</code> — Skip forward 2 minutes

<b>⚠️ Notes:</b>
• Minimum: any positive value
• Maximum: track_duration - current_position - 10 seconds`

	helpTexts["/seekback"] = `<i>Seek backward in the currently playing track.</i>

<u>Usage:</u>
<b>/seekback [seconds]</b> — Go back by specified seconds

<b>🔒 Restrictions:</b>
• Only <b>chat admins</b> or <b>authorized users</b> can use this

<b>💡 Examples:</b>
<code>/seekback 15</code> — Go back 15 seconds
<code>/seekback 60</code> — Go back 1 minute
`

	helpTexts["/jump"] = `<i>Jump to a specific position in the track.</i>

<u>Usage:</u>
<b>/jump [seconds]</b> — Jump to exact position

<b>⚙️ Features:</b>
• Absolute position seeking
• Precise time control
• 10-second buffer from end

<b>🔒 Restrictions:</b>
• Only <b>chat admins</b> or <b>authorized users</b> can use this

<b>💡 Examples:</b>
<code>/jump 90</code> — Jump to 1:30
<code>/jump 0</code> — Jump to start (same as /replay)

<b>⚠️ Notes:</b>
• Position must be within track duration - 10 seconds
• More precise than <code>/seek</code> and <code>/seekback</code>`
}

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

func jumpHandler(m *telegram.NewMessage) error {
	return handleJump(m, false)
}

func cjumpHandler(m *telegram.NewMessage) error {
	return handleJump(m, true)
}

func handleSeek(m *telegram.NewMessage, cplay, isBack bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.ErrEndGroup
	}
	chatID := m.ChannelID()
	t := r.Track()
	if !r.IsActiveChat() {
		m.Reply(F(chatID, "seek_no_active"))
		return telegram.ErrEndGroup
	}

	args := strings.Fields(m.Text())
	if len(args) < 2 {
		m.Reply(F(chatID, "seek_usage", locales.Arg{
			"cmd": getCommand(m),
		}))
		return telegram.ErrEndGroup
	}

	seconds, err := strconv.Atoi(args[1])
	if err != nil {
		m.Reply(F(chatID, "seek_invalid_seconds", locales.Arg{
			"cmd": getCommand(m),
		}))
		return telegram.ErrEndGroup
	}

	var direction, emoji string
	var seekErr error

	if isBack {
		if (r.Position() - seconds) <= 10 {
			m.Reply(F(chatID, "seek_too_close_start", locales.Arg{
				"seconds": seconds,
			}))
			return telegram.ErrEndGroup
		}
		seekErr = r.Seek(-seconds)
		direction = "backward"
		emoji = "⏪"
	} else {
		if (t.Duration - (r.Position() + seconds)) <= 10 {
			m.Reply(F(chatID, "seek_too_close_end", locales.Arg{
				"seconds": seconds,
			}))
			return telegram.ErrEndGroup
		}
		seekErr = r.Seek(seconds)
		direction = "forward"
		emoji = "⏩"
	}

	if seekErr != nil {
		m.Reply(F(chatID, "seek_failed", locales.Arg{
			"direction": direction,
			"seconds":   seconds,
			"error":     seekErr,
		}))
	} else {
		m.Reply(F(chatID, "seek_success", locales.Arg{
			"emoji":     emoji,
			"direction": direction,
			"position":  formatDuration(r.Position()),
			"duration":  formatDuration(t.Duration),
		}))
	}

	return telegram.ErrEndGroup
}

func handleJump(m *telegram.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.ErrEndGroup
	}

	chatID := m.ChannelID()
	t := r.Track()

	if !r.IsActiveChat() || t.ID == "" {
		m.Reply(F(chatID, "jump_no_active"))
		return telegram.ErrEndGroup
	}

	args := strings.Fields(m.Text())
	if len(args) < 2 {
		m.Reply(F(chatID, "jump_usage", locales.Arg{
			"cmd": getCommand(m),
		}))
		return telegram.ErrEndGroup
	}

	seconds, err := strconv.Atoi(args[1])
	if err != nil || seconds < 0 {
		m.Reply(F(chatID, "jump_invalid_position", locales.Arg{
			"cmd": getCommand(m),
		}))
		return telegram.ErrEndGroup
	}

	if t.Duration-seconds <= 10 {
		m.Reply(F(chatID, "jump_too_close_end", locales.Arg{
			"position": formatDuration(seconds),
		}))
		return telegram.ErrEndGroup
	}

	if err := r.Seek(seconds - r.Position()); err != nil {
		m.Reply(F(chatID, "jump_failed", locales.Arg{
			"position": formatDuration(seconds),
			"error":    err,
		}))
	} else {
		m.Reply(F(chatID, "jump_success", locales.Arg{
			"position": formatDuration(seconds),
			"duration": formatDuration(t.Duration),
		}))
	}

	return telegram.ErrEndGroup
}
