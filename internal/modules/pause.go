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

	"github.com/TheTeamVivek/YukkiMusic/internal/utils"
)

func pauseHandler(m *telegram.NewMessage) error {
	return handlePause(m, false)
}

func cpauseHandler(m *telegram.NewMessage) error {
	return handlePause(m, true)
}

func handlePause(m *telegram.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.EndGroup
	}
	if !r.IsActiveChat() {
		m.Reply("⚠️ <b>No active playback.</b>\nThere’s nothing playing right now.")
		return telegram.EndGroup
	}
	if r.IsMuted() {
		m.Reply("⚠️ <b>Oops!</b>\nThe room is muted. Please unmute it first to pause playback.")
		return telegram.EndGroup
	}
	if r.IsPaused() {
		remaining := r.RemainingResumeDuration()
		if remaining > 0 {
			m.Reply(fmt.Sprintf("⏸️ <b>Already Paused</b>\nThe music is already paused.\nAuto-resume in <b>%s</b>", formatDuration(int(remaining.Seconds()))))
		} else {
			m.Reply("⏸️ <b>Already Paused</b>\nThe music is already paused. Would you like to resume it?")
		}
		return telegram.EndGroup
	}
	args := strings.Fields(m.Text())
	var autoResumeDuration time.Duration
	if len(args) >= 2 {
		rawDuration := strings.ToLower(strings.TrimSpace(args[1]))
		rawDuration = strings.TrimSuffix(rawDuration, "s")
		seconds, err := strconv.Atoi(rawDuration)
		if err != nil || seconds < 5 || seconds > 3600 {
			m.Reply("⚠️ Invalid duration for auto-resume. It must be between <b>5</b> and <b>3600</b> seconds.")
			return telegram.EndGroup
		}
		autoResumeDuration = time.Duration(seconds) * time.Second
	}
	var pauseErr error
	if autoResumeDuration > 0 {
		_, pauseErr = r.Pause(autoResumeDuration)
	} else {
		_, pauseErr = r.Pause()
	}
	if pauseErr != nil {
		m.Reply(fmt.Sprintf("❌ <b>Playback Pause Failed</b>\nError: <code>%v</code>", pauseErr))
		return telegram.EndGroup
	}
	mention := utils.MentionHTML(m.Sender)
	trackTitle := html.EscapeString(utils.ShortTitle(r.Track.Title, 25))
	msg := fmt.Sprintf("⏸️ <b>Paused playback</b>\n\nTrack: \"%s\"\nPosition: %s / %s\nPaused by: %s\n",
		trackTitle, formatDuration(r.Position), formatDuration(r.Track.Duration), mention)
	if sp := r.GetSpeed(); sp != 1.0 {
		msg += fmt.Sprintf("⚙️ Speed: <b>%.2fx</b>\n", sp)
	}
	if autoResumeDuration > 0 {
		msg += fmt.Sprintf("\n<i>⏳ Will auto-resume playback after %d seconds</i>", int(autoResumeDuration.Seconds()))
	}
	m.Reply(msg)
	return telegram.EndGroup
}
