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

	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/speed"] = `<i>Control playback speed (tempo).</i>

<u>Usage:</u>
<b>/speed</b> — Show current speed
<b>/speed [multiplier]</b> — Set speed (0.5-4.0x)
<b>/speed [multiplier] [seconds]</b> — Set with auto-reset timer
<b>/speed normal</b> or <b>/speed reset</b> — Reset to 1.0x

<b>⚙️ Features:</b>
• Range: 0.50x to 4.00x
• Auto-reset timer (5-3600 seconds)
• Pitch preservation
• Real-time adjustment

<b>🔒 Restrictions:</b>
• Only <b>chat admins</b> or <b>authorized users</b> can use this

<b>💡 Examples:</b>
<code>/speed 1.5</code> — Play 1.5x faster
<code>/speed 0.75</code> — Play slower (0.75x)
<code>/speed 2.0 300</code> — 2x speed for 5 minutes, then reset
<code>/speed normal</code> — Reset to normal speed

<b>⚠️ Notes:</b>
• Speed affects duration calculations
• Auto-reset only works for non-1.0x speeds
• Suffix 'x' is optional: <code>1.5</code> = <code>1.5x</code>`
}

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
		return telegram.ErrEndGroup
	}

	chatID := m.ChannelID()
	t := r.Track()

	if !r.IsActiveChat() {
		m.Reply(F(chatID, "room_no_active"))
		return telegram.ErrEndGroup
	}

	args := strings.Fields(m.Text())

	// No args -> show current speed or usage hint
	if len(args) < 2 {
		if r.Speed() != 1.0 {
			remaining := r.RemainingSpeedDuration()
			if remaining > 0 {
				m.Reply(F(chatID, "speed_current_with_reset", locales.Arg{
					"speed":    fmt.Sprintf("%.2f", r.Speed()),
					"title":    html.EscapeString(utils.ShortTitle(t.Title, 25)),
					"duration": formatDuration(int(remaining.Seconds())),
					"cmd":      getCommand(m),
				}))
			} else {
				m.Reply(F(chatID, "speed_current", locales.Arg{
					"speed": fmt.Sprintf("%.2f", r.Speed()),
					"title": html.EscapeString(utils.ShortTitle(t.Title, 25)),
					"cmd":   getCommand(m),
				}))
			}
		} else {
			m.Reply(F(chatID, "speed_usage", locales.Arg{
				"cmd": getCommand(m),
			}))
		}
		return telegram.ErrEndGroup
	}

	// Parse speed
	raw := strings.ToLower(strings.TrimSpace(args[1]))
	raw = strings.TrimSuffix(raw, "x")
	raw = strings.TrimSuffix(raw, "×")

	var newSpeed float64
	if raw == "normal" || raw == "reset" || raw == "default" {
		newSpeed = 1.0
	} else {
		s, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			m.Reply(F(chatID, "speed_invalid_value", locales.Arg{
				"cmd": getCommand(m),
			}))
			return telegram.ErrEndGroup
		}
		if s < 0.50 || s > 4.0 {
			m.Reply(F(chatID, "speed_invalid_range"))
			return telegram.ErrEndGroup
		}
		newSpeed = s
	}

	// Parse auto-reset duration
	var resetDuration time.Duration
	if len(args) >= 3 {
		d := strings.ToLower(strings.TrimSpace(args[2]))
		d = strings.TrimSuffix(d, "s")

		seconds, err := strconv.Atoi(d)
		if err != nil || seconds < 5 || seconds > 3600 {
			m.Reply(F(chatID, "speed_invalid_duration"))
			return telegram.ErrEndGroup
		}
		resetDuration = time.Duration(seconds) * time.Second
	}

	// Same speed → give info
	if newSpeed == r.Speed() {
		if resetDuration == 0 {
			m.Reply(F(chatID, "speed_already_set", locales.Arg{
				"speed": fmt.Sprintf("%.2f", newSpeed),
				"title": html.EscapeString(utils.ShortTitle(t.Title, 25)),
			}))
		} else if newSpeed != 1.0 {
			m.Reply(F(chatID, "speed_already_set_reset_hint", locales.Arg{
				"speed": fmt.Sprintf("%.2f", newSpeed),
				"title": html.EscapeString(utils.ShortTitle(t.Title, 25)),
				"cmd":   getCommand(m),
			}))
		}
		return telegram.ErrEndGroup
	}

	// Apply speed
	var setErr error
	if resetDuration > 0 && newSpeed != 1.0 {
		setErr = r.SetSpeed(newSpeed, resetDuration)
	} else {
		setErr = r.SetSpeed(newSpeed)
	}

	if setErr != nil {
		m.Reply(F(chatID, "speed_failed", locales.Arg{
			"speed": fmt.Sprintf("%.2f", newSpeed),
			"error": setErr.Error(),
		}))
		return telegram.ErrEndGroup
	}

	mention := utils.MentionHTML(m.Sender)

	if newSpeed == 1.0 {
		m.Reply(F(chatID, "speed_reset_success", locales.Arg{
			"user": mention,
		}))
	} else {
		if resetDuration > 0 {
			m.Reply(F(chatID, "speed_set_with_reset", locales.Arg{
				"speed":    fmt.Sprintf("%.2f", newSpeed),
				"user":     mention,
				"duration": int(resetDuration.Seconds()),
			}))
		} else {
			m.Reply(F(chatID, "speed_set", locales.Arg{
				"speed": fmt.Sprintf("%.2f", newSpeed),
				"user":  mention,
			}))
		}
	}

	return telegram.ErrEndGroup
}
