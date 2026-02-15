/*
  - This file is part of YukkiMusic.
    *

  - YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
  - Copyright (C) 2025 TheTeamVivek
    *
  - This program is free software: you can redistribute it and/or modify
  - it under the terms of the GNU General Public License as published by
  - the Free Software Foundation, either version 3 of the License, or
  - (at your option) any later version.
    *
  - This program is distributed in the hope that it will be useful,
  - but WITHOUT ANY WARRANTY; without even the implied warranty of
  - MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
  - GNU General Public License for more details.
    *
  - You should have received a copy of the GNU General Public License
  - along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package modules

import (
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/autoplay"] = `<i>Enable or disable AutoPlay in the current chat.</i>

<u>Usage:</u>
<b>/autoplay [on|off]</b> — Toggle AutoPlay

<b>⚙️ Behavior:</b>
• When enabled, the bot automatically plays recommended tracks when the queue is empty.
• Recommendations are currently only supported for YouTube tracks.
• AutoPlay is room-level and needs to be enabled for each session.`
}

func autoplayHandler(m *tg.NewMessage) error {
	return handleAutoplay(m, false)
}

func cautoplayHandler(m *tg.NewMessage) error {
	return handleAutoplay(m, true)
}

func handleAutoplay(m *tg.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return tg.ErrEndGroup
	}
	args := strings.Fields(m.Text())
	chatID := m.ChannelID()

	if len(args) == 1 {
		action := F(chatID, "disabled")
		if r.Autoplay() {
			action = F(chatID, "enabled")
		}
		m.Reply(F(chatID, "autoplay_usage", locales.Arg{
			"cmd":    args[0],
			"action": action,
		}))
		return tg.ErrEndGroup
	}

	if !r.IsActiveChat() {
		m.Reply(F(chatID, "room_no_active"))
		return tg.ErrEndGroup
	}

	enabled, err := utils.ParseBool(args[1])
	if err != nil {
		m.Reply(F(chatID, "invalid_bool"))
		return telegram.ErrEndGroup
	}

	r.DeleteData("rec_cache")
	r.SetAutoplay(enabled)

	action := F(chatID, "disabled")
	if enabled {
		action = F(chatID, "enabled")
	}

	m.Reply(F(chatID, "autoplay_updated", locales.Arg{
		"user":   m.Sender.FirstName,
		"action": action,
	}))

	return tg.ErrEndGroup
}
