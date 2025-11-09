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
	"strconv"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/locales"
	"main/internal/utils"
)

func pauseHandler(m *telegram.NewMessage) error {
	return handlePause(m, false)
}

func cpauseHandler(m *telegram.NewMessage) error {
	return handlePause(m, true)
}

func handlePause(m *tg.NewMessage, cplay bool) error {
	chatID := m.ChannelID()
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return tg.EndGroup
	}

	if !r.IsActiveChat() {
		m.Reply(F(chatID, "room_no_active"))
		return tg.EndGroup
	}

	if r.IsPaused() {
		remaining := r.RemainingResumeDuration()
		autoResumeLine := ""
		if remaining > 0 {
			autoResumeLine = F(chatID, "auto_resume_line", locales.Arg{
				"seconds": formatDuration(int(remaining.Seconds())),
			})
		}
		m.Reply(F(chatID, "pause_already", locales.Arg{
			"auto_resume_line": autoResumeLine,
		}))
		return tg.EndGroup
	}

	args := strings.Fields(m.Text())
	var autoResumeDuration time.Duration
	if len(args) >= 2 {
		raw := strings.ToLower(strings.TrimSpace(args[1]))
		raw = strings.TrimSuffix(raw, "s")
		if sec, convErr := strconv.Atoi(raw); convErr == nil {
			if sec < 5 || sec > 3600 {
				m.Reply(F(chatID, "pause_invalid_duration"))
				return tg.EndGroup
			}
			autoResumeDuration = time.Duration(sec) * time.Second
		} else {
			m.Reply(F(chatID, "pause_invalid_format", locales.Arg{
				"cmd": getCommand(m),
			}))
			return tg.EndGroup
		}
	}

	var pauseErr error
	if autoResumeDuration > 0 {
		_, pauseErr = r.Pause(autoResumeDuration)
	} else {
		_, pauseErr = r.Pause()
	}
	if pauseErr != nil {
		m.Reply(F(chatID, "room_pause_failed", locales.Arg{
			"error": pauseErr.Error(),
		}))
		return tg.EndGroup
	}

	mention := utils.MentionHTML(m.Sender)
	title := html.EscapeString(utils.ShortTitle(r.Track.Title, 25))

	autoResumeLine := ""
	if autoResumeDuration > 0 {
		autoResumeLine = F(chatID, "auto_resume_line", locales.Arg{
			"seconds": int(autoResumeDuration.Seconds()),
		})
	}

	msg := F(chatID, "pause_success", locales.Arg{
		"title":            title,
		"position":         formatDuration(r.Position),
		"duration":         formatDuration(r.Track.Duration),
		"user":             mention,
		"auto_resume_line": autoResumeLine,
	})

	if sp := r.GetSpeed(); sp != 1.0 {
		msg += "\n" + F(chatID, "speed_line", locales.Arg{
			"speed": fmt.Sprintf("%.2f", sp),
		})
	}

	m.Reply(msg)
	return tg.EndGroup
}
