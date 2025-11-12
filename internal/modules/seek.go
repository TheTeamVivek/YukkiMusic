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
	chatID := m.ChannelID()

	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.EndGroup
	}

	if !r.IsActiveChat() {
		m.Reply(F(chatID, "seek_no_active"))
		return telegram.EndGroup
	}

	args := strings.Fields(m.Text())
	if len(args) < 2 {
		m.Reply(F(chatID, "seek_usage", locales.Arg{
			"cmd": getCommand(m),
		}))
		return telegram.EndGroup
	}

	seconds, err := strconv.Atoi(args[1])
	if err != nil {
		m.Reply(F(chatID, "seek_invalid_seconds", locales.Arg{
			"cmd": getCommand(m),
		}))
		return telegram.EndGroup
	}

	var direction, emoji string
	var seekErr error

	if isBack {
		if (r.Position - seconds) <= 10 {
			m.Reply(F(chatID, "seek_too_close_start", locales.Arg{
				"seconds": seconds,
			}))
			return telegram.EndGroup
		}
		seekErr = r.Seek(-seconds)
		direction = "backward"
		emoji = "⏪"
	} else {
		if (r.Track.Duration - (r.Position + seconds)) <= 10 {
			m.Reply(F(chatID, "seek_too_close_end", locales.Arg{
				"seconds": seconds,
			}))
			return telegram.EndGroup
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
			"position":  formatDuration(r.Position),
			"duration":  formatDuration(r.Track.Duration),
		}))
	}

	return telegram.EndGroup
}

func handleJump(m *telegram.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.EndGroup
	}

	chatID := m.ChannelID()

	if !r.IsActiveChat() || r.Track == nil {
		m.Reply(F(chatID, "jump_no_active"))
		return telegram.EndGroup
	}

	args := strings.Fields(m.Text())
	if len(args) < 2 {
		m.Reply(F(chatID, "jump_usage", locales.Arg{
			"cmd": getCommand(m),
		}))
		return telegram.EndGroup
	}

	seconds, err := strconv.Atoi(args[1])
	if err != nil || seconds < 0 {
		m.Reply(F(chatID, "jump_invalid_position", locales.Arg{
			"cmd": getCommand(m),
		}))
		return telegram.EndGroup
	}

	if r.Track.Duration-seconds <= 10 {
		m.Reply(F(chatID, "jump_too_close_end", locales.Arg{
			"position": formatDuration(seconds),
		}))
		return telegram.EndGroup
	}

	if err := r.Seek(seconds - r.Position); err != nil {
		m.Reply(F(chatID, "jump_failed", locales.Arg{
			"position": formatDuration(seconds),
			"error":    err,
		}))
	} else {
		m.Reply(F(chatID, "jump_success", locales.Arg{
			"position": formatDuration(seconds),
			"duration": formatDuration(r.Track.Duration),
		}))
	}

	return telegram.EndGroup
}
