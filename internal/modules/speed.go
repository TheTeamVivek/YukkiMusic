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
	"fmt"
	"strconv"
	"strings"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func init() {
	helpTexts["/speed"] = `<i>Control playback speed (tempo).</i>

<u>Usage:</u>
<b>/speed</b> — Show current speed
<b>/speed [multiplier]</b> — Set speed (0.5-4.0x)
<b>/speed normal</b> or <b>/speed reset</b> — Reset to 1.0x

<b>⚙️ Features:</b>
• Range: 0.50x to 4.00x
• Pitch preservation
• Real-time adjustment

<b>🔒 Restrictions:</b>
• Only <b>chat admins</b> or <b>authorized users</b> can use this

<b>💡 Examples:</b>
<code>/speed 1.5</code> — Play 1.5x faster
<code>/speed 0.75</code> — Play slower (0.75x)
<code>/speed normal</code> — Reset to normal speed

<b>⚠️ Notes:</b>
• Speed affects duration calculations
• Suffix 'x' is optional: <code>1.5</code> = <code>1.5x</code>`
}

func speedHandler(c *td.Client, m *td.Message) error {
	return handleSpeed(c, m, false)
}

func cspeedHandler(c *td.Client, m *td.Message) error {
	return handleSpeed(c, m, true)
}

func handleSpeed(c *td.Client, m *td.Message, cplay bool) error {
	if !isSuperGroup(c, m) {
		return nil
	}

	if !filterAuthUsers(c, m) {
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
		m.ReplyText(c, F(chatID, "room_no_active"), nil)
		return nil
	}

	args := strings.Fields(m.Text())

	// No args -> show current speed or usage hint
	if len(args) < 2 {
		if r.Speed() != 1.0 {
			m.ReplyText(c, F(chatID, "speed_current", locales.Arg{
				"speed": fmt.Sprintf("%.2f", r.Speed()),
				"title": utils.EscapeHTML(utils.ShortTitle(t.Title, 25)),
				"cmd":   getCommand(m),
			}), nil)
		} else {
			m.ReplyText(c, F(chatID, "speed_usage", locales.Arg{
				"cmd": getCommand(m),
			}), nil)
		}
		return nil
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
			m.ReplyText(c, F(chatID, "speed_invalid_value", locales.Arg{
				"cmd": getCommand(m),
			}), nil)
			return nil
		}
		if s < 0.50 || s > 4.0 {
			m.ReplyText(c, F(chatID, "speed_invalid_range"), nil)
			return nil
		}
		newSpeed = s
	}

	// Same speed → give info
	if newSpeed == r.Speed() {
		m.ReplyText(c, F(chatID, "speed_already_set", locales.Arg{
			"speed": fmt.Sprintf("%.2f", newSpeed),
			"title": utils.EscapeHTML(utils.ShortTitle(t.Title, 25)),
		}), nil)
		return nil
	}

	// Apply speed
	if err := r.SetSpeed(newSpeed); err != nil {
		m.ReplyText(c, F(chatID, "speed_failed", locales.Arg{
			"speed": fmt.Sprintf("%.2f", newSpeed),
			"error": err.Error(),
		}), nil)
		return nil
	}

	sender, _ := m.GetUser(c)
	mention := mentionOf(sender, m.SenderID())

	if newSpeed == 1.0 {
		m.ReplyText(c, F(chatID, "speed_reset_success", locales.Arg{
			"user": mention,
		}), nil)
	} else {
		m.ReplyText(c, F(chatID, "speed_set", locales.Arg{
			"speed": fmt.Sprintf("%.2f", newSpeed),
			"user":  mention,
		}), nil)
	}

	return nil
}
