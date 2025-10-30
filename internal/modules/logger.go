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
	"strings"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/database"
	"main/internal/utils"
)

func handleLogger(m *telegram.NewMessage) error {
	args := strings.Fields(m.Text())
	current, dbErr := database.IsLoggerEnabled()

	if len(args) < 2 {
		if dbErr == nil {
			status := "🟢 Enabled"
			if !current {
				status = "🔴 Disabled"
			}
			m.Reply(
				fmt.Sprintf("⚙️ Usage: <code>%s [enable|disable]</code> - To enable or disable the logger\n\n📜 Current status: %s", getCommand(m), status),
			)
		} else {
			m.Reply(fmt.Sprintf("⚙️ Usage: <code>/%s [enable|disable]</code> - To enable or disable the logger", getCommand(m)))
		}
		return telegram.EndGroup
	}

	enable, err := utils.ParseBool(args[1])
	if err != nil {
		m.Reply("⚠️ Invalid option. Use 'enable' or 'disable'.")
		return telegram.EndGroup
	}

	if dbErr != nil {
		m.Reply("❌ Failed to check logger status: " + dbErr.Error())
		return telegram.EndGroup
	}

	if current == enable {
		status := "enabled"
		if !enable {
			status = "disabled"
		}
		m.Reply("ℹ️ Logger is already " + status + ".")
		return telegram.EndGroup
	}

	if err := database.SetLoggerEnabled(enable); err != nil {
		m.Reply("❌ Failed to update logger setting: " + err.Error())
		return telegram.EndGroup
	}

	status := "disabled"
	if enable {
		status = "enabled"
	}
	m.Reply("✅ Logger has been " + status + " successfully.")
	return telegram.EndGroup
}
