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
	"strings"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

func handleLogger(m *telegram.NewMessage) error {
	args := strings.Fields(m.Text())
	chatID := m.ChatID()

	current, dbErr := database.IsLoggerEnabled()

	action := F(chatID, utils.IfElse(current, "enabled", "disabled"))

	if len(args) < 2 {
		if dbErr == nil {
			m.Reply(F(chatID, "logger_usage", locales.Arg{
				"cmd": getCommand(m),
				"status": F(chatID, "logger_status", locales.Arg{
					"action": action,
				}),
			}))
		} else {
			m.Reply(F(chatID, "logger_usage", locales.Arg{
				"cmd":    getCommand(m),
				"status": "",
			}))
		}
		return telegram.EndGroup
	}

	enable, err := utils.ParseBool(args[1])
	if err != nil {
		m.Reply(F(chatID, "invalid_bool"))
		return telegram.EndGroup
	}

	if dbErr != nil {
		m.Reply(F(chatID, "logger_check_fail", locales.Arg{"error": dbErr.Error()}))
		return telegram.EndGroup
	}

	if current == enable {
		m.Reply(F(chatID, "logger_already", locales.Arg{"action": action}))
		return telegram.EndGroup
	}

	if err := database.SetLoggerEnabled(enable); err != nil {
		m.Reply(F(chatID, "logger_update_fail", locales.Arg{"error": err.Error()}))
		return telegram.EndGroup
	}

	m.Reply(F(chatID, "logger_updated", locales.Arg{"action": action}))

	return telegram.EndGroup
}
