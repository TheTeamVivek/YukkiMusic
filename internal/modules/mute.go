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
	"html"
	"strconv"
	"strings"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/locales"
	"main/internal/utils"
)

func muteHandler(m *tg.NewMessage) error {
	return handleMute(m, false)
}

func cmuteHandler(m *tg.NewMessage) error {
	return handleMute(m, true)
}

func handleMute(m *tg.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return tg.EndGroup
	}

	if !r.IsActiveChat() {
		m.Reply(F(m.ChatID(), "room_no_active"))
		return tg.EndGroup
	}

	if r.IsMuted() {
		remaining := r.RemainingUnmuteDuration()
		if remaining > 0 {
			m.Reply(F(m.ChatID(), "mute_already_muted_with_time", locales.Arg{
				"duration": formatDuration(int(remaining.Seconds())),
			}))
		} else {
			m.Reply(F(m.ChatID(), "mute_already_muted"))
		}
		return tg.EndGroup
	}

	mention := utils.MentionHTML(m.Sender)
	args := strings.Fields(m.Text())
	var autoUnmuteDuration time.Duration

	if len(args) >= 2 {
		rawDuration := strings.ToLower(strings.TrimSpace(args[1]))
		rawDuration = strings.TrimSuffix(rawDuration, "s")

		if seconds, err := strconv.Atoi(rawDuration); err == nil {
			if seconds < 5 || seconds > 3600 {
				m.Reply(F(m.ChatID(), "mute_invalid_duration"))
				return tg.EndGroup
			}
			autoUnmuteDuration = time.Duration(seconds) * time.Second
		} else {
			m.Reply(F(m.ChatID(), "mute_invalid_format", locales.Arg{
				"cmd": getCommand(m),
			}))
			return tg.EndGroup
		}
	}

	var muteErr error
	if autoUnmuteDuration > 0 {
		_, muteErr = r.Mute(autoUnmuteDuration)
	} else {
		_, muteErr = r.Mute()
	}

	if muteErr != nil {
		m.Reply(F(m.ChatID(), "mute_failed", locales.Arg{
			"error": muteErr.Error(),
		}))
		return tg.EndGroup
	}

	msgArgs := locales.Arg{
		"title": html.EscapeString(utils.ShortTitle(r.Track.Title, 25)),
		"user":  mention,
	}

	if autoUnmuteDuration > 0 {
		msgArgs["duration"] = int(autoUnmuteDuration.Seconds())
		m.Reply(F(m.ChatID(), "mute_success_with_time", msgArgs))
	} else {
		m.Reply(F(m.ChatID(), "mute_success", msgArgs))
	}

	return tg.EndGroup
}
