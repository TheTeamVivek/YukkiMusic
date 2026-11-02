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
	"strings"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/database"
	"yukkimusic/internal/utils"
)

func init() {
	dashboards["logger"] = dashboardDef{
		title:       func(chatID int64) string { return F(chatID, "logger_title") },
		description: func(chatID int64) string { return F(chatID, "logger_desc") },
		canUse: func(c *td.Client, u *td.UpdateNewCallbackQuery) bool {
			if isOwnerOrSudo(u.SenderUserId) {
				return true
			}
			u.Answer(c, F(u.ChatId, "only_sudo_cb"), &td.CallbackAnswerOpts{ShowAlert: true})
			return false
		},
		get: func(int64) (bool, error) { return database.IsLoggerEnabled() },
		set: func(_ int64, on bool) error { return database.SetLoggerEnabled(on) },
	}
}

func handleLogger(c *td.Client, m *td.Message) error {
	if !checkSudo(c, m) {
		return nil
	}

	if args := strings.Fields(m.Text()); len(args) > 1 {
		if enabled, err := utils.ParseBool(args[1]); err == nil {
			if current, cerr := database.IsLoggerEnabled(); cerr == nil && enabled != current {
				if err := database.SetLoggerEnabled(enabled); err != nil {
					return err
				}
			}
		}
	}

	return sendDashboard(c, m, "logger")
}
