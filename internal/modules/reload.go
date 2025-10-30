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
 * GNU General Public License for a copy of the GNU General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 */

package modules

import (
	"errors"
	"fmt"
	"time"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	"main/internal/utils"
)

func reloadHandler(m *telegram.NewMessage) error {
	return handleReload(m, false)
}

func creloadHandler(m *telegram.NewMessage) error {
	return handleReload(m, true)
}

func handleReload(m *telegram.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.EndGroup
	}
	chatID := r.ChatID
	userID := m.SenderID()
	floodKey := fmt.Sprintf("reload:%d%d", chatID, userID)
	floodDuration := 10 * time.Minute
	if remaining := utils.GetFlood(floodKey); remaining > 0 {
		return m.E(m.Reply(fmt.Sprintf(
			"⏳ Please wait %s minutes before using this command again.",
			formatDuration(int(remaining.Seconds())),
		)))
	}

	mystic, err := m.Reply("⚙️ Reloading admin cache, voice chat status, and assistant status...\n")
	if err != nil {
		return err
	}

	summary := ""
	admins, adminErr := utils.ReloadChatAdmin(m.Client, chatID)
	if adminErr != nil {
		summary += fmt.Sprintf("<b>• Admin cache:</b> <i>❌ (%v)</i>\n", adminErr)
	} else {
		summary += "<b>• Admin cache:</b> <i>✅ Done</i>\n"
	}

	isAdmin := false
	if adminErr == nil {
		for _, id := range admins {
			if id == userID {
				isAdmin = true
				break
			}
		}
	}

	if isAdmin {
		floodDuration = 5 * time.Minute
	}
	utils.SetFlood(floodKey, floodDuration)

	voiceActive, voiceErr := core.GetVoiceChatStatus(chatID, true)
	if voiceErr != nil {
		switch {
		case errors.Is(voiceErr, core.ErrNoActiveVoiceChat):
			summary += "<b>• Voice chat:</b> <i>⚪ Inactive</i>\n"
		case errors.Is(voiceErr, core.ErrAdminPermissionRequired):
			summary += "<b>• Voice chat:</b> <i>❌ Admin permission required</i>\n"
		default:
			summary += fmt.Sprintf("<b>• Voice chat:</b> <i>❌ (%v)</i>\n", voiceErr)
		}
	} else if voiceActive {
		summary += "<b>• Voice chat:</b> <i>✅ Active</i>\n"
	}

	assistantActive, assistantErr := core.GetAssistantStatus(chatID, true)
	if assistantErr != nil {
		switch {
		case errors.Is(assistantErr, core.ErrAssistantBanned):
			summary += "<b>• Assistant:</b> <i>❌ Banned</i>\n"
		case errors.Is(assistantErr, core.ErrAdminPermissionRequired):
			summary += "<b>• Assistant:</b> <i>❌ Admin permission required</i>\n"
		case errors.Is(assistantErr, core.ErrAssistantJoinRejected):
			summary += "<b>• Assistant:</b> <i>❌ Invite rejected or invalid</i>\n"
		case errors.Is(assistantErr, core.ErrAssistantJoinRateLimited):
			summary += "<b>• Assistant:</b> <i>❌ Rate limited</i>\n"
		case errors.Is(assistantErr, core.ErrAssistantJoinRequestSent):
			summary += "<b>• Assistant:</b> <i>⚪ Join request sent</i>\n"
		default:
			summary += fmt.Sprintf("<b>• Assistant:</b> <i>❌ (%v)</i>\n", assistantErr)
		}
	} else if assistantActive {
		summary += "<b>• Assistant:</b> <i>✅ Present</i>\n"
	} else {
		summary += "<b>• Assistant:</b> <i>⚪ Not present</i>\n"
	}

	// --- Destroy room if user is admin ---
	if isAdmin {
		if room, ok := core.GetRoom(chatID); ok {
			room.Destroy()
			summary += "<b>• Room:</b> <i>Reset ✅</i>\n"
		}
	}

	utils.EOR(mystic, "<b>⚙️ Reload complete:</b></u>\n\n"+summary)

	return nil
}
