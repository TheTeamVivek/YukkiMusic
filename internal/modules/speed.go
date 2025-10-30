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

	"main/internal/utils"
)

func speedHandler(m *telegram.NewMessage) error {
	return handleSpeed(m, false)
}

func cspeedHandler(m *telegram.NewMessage) error {
	return handleSpeed(m, true)
}

func handleSpeed(m *telegram.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.EndGroup
	}

	if !r.IsActiveChat() {
		m.Reply("🎵 No track is currently playing.")
		return telegram.EndGroup
	}

	args := strings.Fields(m.Text())

	if len(args) < 2 {
		if r.Speed != 1.0 {
			remaining := r.RemainingSpeedDuration()
			if remaining > 0 {
				m.Reply(fmt.Sprintf(
					"🎧 <b>Current Speed:</b> <code>%.2fx</code>\n\n"+
						"🎵 <b>Track:</b> <u>%s</u>\n"+
						"⏳ <b>Auto-reset in:</b> %s\n\n"+
						"💡 Use <code>%s reset</code> to restore normal playback speed instantly.",
					r.Speed,
					html.EscapeString(utils.ShortTitle(r.Track.Title, 25)),
					formatDuration(int(remaining.Seconds())),
					getCommand(m),
				))
			} else {
				m.Reply(fmt.Sprintf(
					"🎧 <b>Current Speed:</b> <code>%.2fx</code>\n\n"+
						"🎵 <b>Track:</b> <u>%s</u>\n"+
						"💡 Use <code>%s reset</code> to restore normal playback speed.",
					r.Speed,
					html.EscapeString(utils.ShortTitle(r.Track.Title, 25)),
					getCommand(m),
				))
			}
		} else {
			m.Reply(
				"💡 <b>Usage:</b>\n" +
					fmt.Sprintf("<code>%s [speed] [duration]</code> — change playback speed\n\n", getCommand(m)) +
					"Allowed speed range: <b>0.50x → 4.00x</b>\n" +
					"Specify a duration to auto-reset speed after given seconds (optional).\n\n" +
					fmt.Sprintf("Example:\n<code>%s 1.25 45</code> — play at 1.25x for 45 seconds, then return to normal speed.", getCommand(m)),
			)
		}
		return telegram.EndGroup
	}

	// Parse speed argument
	inputSpeed := strings.ToLower(strings.TrimSpace(args[1]))
	inputSpeed = strings.TrimSuffix(inputSpeed, "x")
	inputSpeed = strings.TrimSuffix(inputSpeed, "×")

	var newSpeed float64
	if inputSpeed == "normal" || inputSpeed == "reset" || inputSpeed == "default" {
		newSpeed = 1.0
	} else {
		speed, err := strconv.ParseFloat(inputSpeed, 64)
		if err != nil {
			m.Reply(fmt.Sprintf("❌ Invalid speed value.\nExample: <code>%s 1.5</code> or <code>%s 2x</code>", getCommand(m), getCommand(m)))
			return telegram.EndGroup
		}
		if speed < 0.50 || speed > 4.0 {
			m.Reply("⚠️ Speed must be between <b>0.50x</b> and <b>4.00x</b>.")
			return telegram.EndGroup
		}
		newSpeed = speed
	}

	// Parse optional duration argument for auto reset
	var durationAfterNormal time.Duration
	if len(args) >= 3 {
		rawDuration := strings.ToLower(strings.TrimSpace(args[2]))
		rawDuration = strings.TrimSuffix(rawDuration, "s")
		seconds, err := strconv.Atoi(rawDuration)
		if err != nil || seconds < 5 || seconds > 3600 {
			m.Reply("⚠️ Invalid duration value. It must be between <b>5</b> and <b>3600</b> seconds. Suffix 's' is optional.")
			return telegram.EndGroup
		}
		durationAfterNormal = time.Duration(seconds) * time.Second
	}

	if newSpeed == r.Speed {
		if durationAfterNormal == 0 {
			m.Reply(fmt.Sprintf(
				"ℹ️ Playback speed is already set to <b>%.2fx</b>\n🎵 Track: <u>%s</u>",
				newSpeed,
				html.EscapeString(utils.ShortTitle(r.Track.Title, 25)),
			))
		} else if newSpeed != 1.0 {
			m.Reply(fmt.Sprintf(
				"ℹ️ Playback speed is already set to <b>%.2fx</b>\n🎵 Track: <u>%s</u>\n\n<b>Use <code>%s reset</code> for resetting speed</b>",
				newSpeed,
				html.EscapeString(utils.ShortTitle(r.Track.Title, 25)),
				getCommand(m),
			))
		}
		return telegram.EndGroup
	}

	// Set speed with optional duration
	var setErr error
	if durationAfterNormal > 0 && newSpeed != 1.0 {
		setErr = r.SetSpeed(newSpeed, durationAfterNormal)
	} else {
		setErr = r.SetSpeed(newSpeed)
	}

	if setErr != nil {
		m.Reply(fmt.Sprintf("❌ Failed to change speed to <b>%.2fx</b>.\nError: %v", newSpeed, setErr))
		return telegram.EndGroup
	}

	mention := utils.MentionHTML(m.Sender)
	if newSpeed == 1.0 {
		m.Reply("✅ Playback speed reset to <b>1.0x</b> by " + mention)
	} else {
		msg := fmt.Sprintf("🚀 Playback speed set to <b>%.2fx</b> by %s.", newSpeed, mention)
		if durationAfterNormal > 0 {
			msg += fmt.Sprintf("\n\n<i>⏳ Will reset to normal after %d seconds</i>", int(durationAfterNormal.Seconds()))
		}
		m.Reply(msg)
	}

	return telegram.EndGroup
}
