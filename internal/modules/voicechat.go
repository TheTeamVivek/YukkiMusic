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
	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	"main/internal/database"
	"main/internal/utils"
)

func handleActions(m *telegram.NewMessage) error {
	if !isValidChatType(m) {
		warnAndLeave(m.Client, m.ChannelID())
		return telegram.ErrEndGroup
	}

	if action, ok := m.Action.(*telegram.MessageActionGroupCall); ok {
		return handleVoiceChatAction(m, action)
	}

	return telegram.ErrEndGroup
}

func handleVoiceChatAction(
	m *telegram.NewMessage,
	action *telegram.MessageActionGroupCall,
) error {
	if isMaint, _ := database.IsMaintenance(); isMaint {
		return telegram.ErrEndGroup
	}

	chatID := m.ChannelID()
	core.DeleteRoom(chatID)

	s, err := core.GetChatState(chatID)
	if err != nil {
		gologging.Error("Failed to get chat state: " + err.Error())
		return telegram.ErrEndGroup
	}
	isActive := action.Duration == 0
	msgKey := utils.IfElse(isActive, "voicechat_started", "voicechat_ended")

	s.SetVoiceChatActive(isActive)
	m.Respond(F(chatID, msgKey))
	gologging.Debug("Voice chat " + msgKey + " in " + utils.IntToStr(chatID))

	return telegram.ErrEndGroup
}

func isValidChatType(m *telegram.NewMessage) bool {
	return m.ChatType() != telegram.EntityChat ||
		(m.Channel != nil && m.Channel.Megagroup)
}
