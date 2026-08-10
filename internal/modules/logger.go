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
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func handleLogger(c *td.Client, m *td.Message) error {
	if !checkSudo(c, m) {
		return nil
	}
	args := strings.Fields(m.Text())
	chatID := m.ChatID()

	current, dbErr := database.IsLoggerEnabled()

	action := F(chatID, utils.IfElse(current, "enabled", "disabled"))

	if len(args) < 2 {
		if dbErr == nil {
			if _, err := m.ReplyText(c, F(chatID, "logger_usage", locales.Arg{
				"cmd": getCommand(m),
				"status": F(chatID, "logger_status", locales.Arg{
					"action": action,
				}),
			}), nil); err != nil {
				return err
			}
		} else {
			if _, err := m.ReplyText(c, F(chatID, "logger_usage", locales.Arg{
				"cmd":    getCommand(m),
				"status": "",
			}), nil); err != nil {
				return err
			}
		}
		return nil
	}

	enable, err := utils.ParseBool(args[1])
	if err != nil {
		_, err := m.ReplyText(c, F(chatID, "invalid_bool"), nil)
		return err
	}

	action = F(chatID, utils.IfElse(enable, "enabled", "disabled"))
	if dbErr != nil {
		_, err := m.ReplyText(
			c,
			F(chatID, "logger_check_fail", locales.Arg{"error": dbErr.Error()}),
			nil,
		)
		return err
	}

	if current == enable {
		_, err := m.ReplyText(c, F(chatID, "logger_already", locales.Arg{"action": action}), nil)
		return err
	}

	if err := database.SetLoggerEnabled(enable); err != nil {
		_, err := m.ReplyText(
			c,
			F(chatID, "logger_update_fail", locales.Arg{"error": err.Error()}),
			nil,
		)
		return err
	}
	_, rerr := m.ReplyText(c, F(chatID, "logger_updated", locales.Arg{"action": action}), nil)
	return rerr
}
