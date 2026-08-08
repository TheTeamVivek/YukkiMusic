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
	"strconv"
	"strings"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
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

func seekHandler(c *td.Client, m *td.Message) error {
	return handleSeek(c, m, false, false)
}

func cseekHandler(c *td.Client, m *td.Message) error {
	return handleSeek(c, m, true, false)
}

func seekbackHandler(c *td.Client, m *td.Message) error {
	return handleSeek(c, m, false, true)
}

func cseekbackHandler(c *td.Client, m *td.Message) error {
	return handleSeek(c, m, true, true)
}

func jumpHandler(c *td.Client, m *td.Message) error {
	return handleJump(c, m, false)
}

func cjumpHandler(c *td.Client, m *td.Message) error {
	return handleJump(c, m, true)
}

func handleSeek(c *td.Client, m *td.Message, cplay, isBack bool) error {
	if !isSuperGroupTd(c, m) {
		return nil
	}

	if !filterAuthUsersTd(c, m) {
		return nil
	}

	r, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		m.ReplyText(c, err.Error(), nil)
		return nil
	}
	chatID := m.ChatID()
	t := r.Track()
	if !r.IsActiveChat() || t == nil {
		m.ReplyText(c, F(chatID, "seek_no_active"), nil)
		return nil
	}

	args := strings.Fields(m.Text())
	if len(args) < 2 {
		m.ReplyText(c, F(chatID, "seek_usage", locales.Arg{
			"cmd": getCommandTd(m),
		}), nil)
		return nil
	}

	seconds, err := strconv.Atoi(args[1])
	if err != nil {
		m.ReplyText(c, F(chatID, "seek_invalid_seconds", locales.Arg{
			"cmd": getCommandTd(m),
		}), nil)
		return nil
	}

	var direction, emoji string
	var seekErr error

	if isBack {
		if (r.Position() - seconds) <= 10 {
			m.ReplyText(c, F(chatID, "seek_too_close_start", locales.Arg{
				"seconds": seconds,
			}), nil)
			return nil
		}
		seekErr = r.Seek(-seconds)
		direction = "backward"
		emoji = "⏪"
	} else {
		if (t.Duration - (r.Position() + seconds)) <= 10 {
			m.ReplyText(c, F(chatID, "seek_too_close_end", locales.Arg{
				"seconds": seconds,
			}), nil)
			return nil
		}
		seekErr = r.Seek(seconds)
		direction = "forward"
		emoji = "⏩"
	}

	if seekErr != nil {
		m.ReplyText(c, F(chatID, "seek_failed", locales.Arg{
			"direction": direction,
			"seconds":   seconds,
			"error":     seekErr,
		}), nil)
	} else {
		m.ReplyText(c, F(chatID, "seek_success", locales.Arg{
			"emoji":     emoji,
			"direction": direction,
			"position":  utils.FormatDuration(r.Position()),
			"duration":  utils.FormatDuration(t.Duration),
		}), nil)
	}

	return nil
}

func handleJump(c *td.Client, m *td.Message, cplay bool) error {
	if !isSuperGroupTd(c, m) {
		return nil
	}

	if !filterAuthUsersTd(c, m) {
		return nil
	}

	r, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		m.ReplyText(c, err.Error(), nil)
		return nil
	}

	chatID := m.ChatID()
	t := r.Track()

	if !r.IsActiveChat() || t == nil {
		m.ReplyText(c, F(chatID, "jump_no_active"), nil)
		return nil
	}

	args := strings.Fields(m.Text())
	if len(args) < 2 {
		m.ReplyText(c, F(chatID, "jump_usage", locales.Arg{
			"cmd": getCommandTd(m),
		}), nil)
		return nil
	}

	seconds, err := strconv.Atoi(args[1])
	if err != nil || seconds < 0 {
		m.ReplyText(c, F(chatID, "jump_invalid_position", locales.Arg{
			"cmd": getCommandTd(m),
		}), nil)
		return nil
	}

	if t.Duration-seconds <= 10 {
		m.ReplyText(c, F(chatID, "jump_too_close_end", locales.Arg{
			"position": utils.FormatDuration(seconds),
		}), nil)
		return nil
	}

	if err := r.Seek(seconds - r.Position()); err != nil {
		m.ReplyText(c, F(chatID, "jump_failed", locales.Arg{
			"position": utils.FormatDuration(seconds),
			"error":    err,
		}), nil)
	} else {
		m.ReplyText(c, F(chatID, "jump_success", locales.Arg{
			"position": utils.FormatDuration(seconds),
			"duration": utils.FormatDuration(t.Duration),
		}), nil)
	}

	return nil
}
