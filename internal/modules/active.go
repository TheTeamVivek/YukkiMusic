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
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	"main/internal/locales"
)

func init() {
	helpTexts["/active"] = `<i>Show all active voice chat sessions.</i>

<u>Usage:</u>
<b>/active</b> or <b>/ac</b> — List active chats

<b>📊 Information Shown:</b>
• Total active chats
• Active NTGCalls connections
• Broken/stale sessions

<b>🔒 Restrictions:</b>
• <b>Sudo users</b> only

<b>💡 Use Case:</b>
Monitor bot usage and identify issues.`

	keys := []string{"/ac", "/activevc", "/activevoice"}
	for _, k := range keys {
		helpTexts[k] = helpTexts["/active"]
	}
}

func activeHandler(m *telegram.NewMessage) error {
	chatID := m.ChannelID()

	allChats := core.GetAllRoomIDs()
	activeCount := len(allChats)

	ntgChats := make(map[int64]struct{})

	core.Assistants.ForEach(func(a *core.Assistant) {
		if a == nil || a.Ntg == nil {
			return
		}
		for id := range a.Ntg.Calls() {
			ntgChats[id] = struct{}{}
		}
	})

	brokenCount := 0
	for _, id := range allChats {
		if _, ok := ntgChats[id]; !ok {
			brokenCount++
		}
	}

	msg := F(chatID, "active_chats_info", locales.Arg{
		"count": activeCount,
	})

	if brokenCount > 0 {
		msg = F(chatID, "active_chats_info_with_broken", locales.Arg{
			"count":  activeCount,
			"broken": brokenCount,
		})
	}

	m.Reply(msg)
	return telegram.ErrEndGroup
}
