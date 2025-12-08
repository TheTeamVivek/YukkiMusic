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
	"main/internal/locales"
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

	chatID := r.ChatID()
	userID := m.SenderID()
	floodKey := fmt.Sprintf("reload:%d%d", chatID, userID)
	floodDuration := 5 * time.Minute

	if remaining := utils.GetFlood(floodKey); remaining > 0 {
		return m.E(m.Reply(F(
			chatID,
			"flood_minutes",
			locales.Arg{
				"duration": formatDuration(int(remaining.Seconds())),
			},
		)))
	}

	mystic, err := m.Reply(F(chatID, "reload_start"))
	if err != nil {
		return err
	}

	summary := ""

	admins, adminErr := utils.ReloadChatAdmin(m.Client, chatID)
	if adminErr != nil {
		summary += F(chatID, "reload_admin_cache_fail", locales.Arg{
			"error": adminErr.Error(),
		}) + "\n"
	} else {
		summary += F(chatID, "reload_admin_cache_ok") + "\n"
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
		floodDuration = 2 * time.Minute
	}
	utils.SetFlood(floodKey, floodDuration)

	cs, err := core.GetChatState(chatID)
	if err != nil {
		summary += F(chatID, "reload_assistant_fail", locales.Arg{
			"error": err.Error(),
		}) + "\n"
		utils.EOR(mystic, F(chatID, "reload_done", locales.Arg{
			"summary": summary,
		}))
		return nil
	}

	activeVC, vcErr := cs.IsActiveVC(true)
	if vcErr != nil {
		switch {
		case errors.Is(vcErr, core.ErrAdminPermissionRequired):
			summary += F(chatID, "reload_voice_admin_required") + "\n"
		default:
			summary += F(chatID, "reload_voice_fail", locales.Arg{
				"error": vcErr.Error(),
			}) + "\n"
		}
	} else if activeVC {
		summary += F(chatID, "reload_voice_active") + "\n"
	} else {
		summary += F(chatID, "reload_voice_inactive") + "\n"
	}

	banned, assErr := cs.IsAssistantBanned(true)
	if assErr != nil {
		switch {
		case errors.Is(assErr, core.ErrAdminPermissionRequired):
			summary += F(chatID, "reload_assistant_admin_required") + "\n"
		default:
			summary += F(chatID, "reload_assistant_fail", locales.Arg{
				"error": assErr.Error(),
			}) + "\n"
		}
	} else if banned {
		summary += F(chatID, "reload_assistant_banned") + "\n"
	} else {
		present, assErr2 := cs.IsAssistantPresent()
		if assErr2 != nil {
			switch {
			case errors.Is(assErr2, core.ErrAdminPermissionRequired):
				summary += F(chatID, "reload_assistant_admin_required") + "\n"
			default:
				summary += F(chatID, "reload_assistant_fail", locales.Arg{
					"error": assErr2.Error(),
				}) + "\n"
			}
		} else if present {
			summary += F(chatID, "reload_assistant_present") + "\n"
		} else {
			summary += F(chatID, "reload_assistant_not_present") + "\n"
		}
	}

	if isAdmin {
		if room, ok := core.GetRoom(chatID); ok {
			room.Destroy()
			summary += F(chatID, "reload_room_reset") + "\n"
		}
	}

	utils.EOR(mystic, F(chatID, "reload_done", locales.Arg{
		"summary": summary,
	}))

	return nil
}
