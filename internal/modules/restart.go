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
	"os"
	"path/filepath"
	"syscall"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	"main/internal/utils"
)

func handleRestart(m *tg.NewMessage) error {
	mystic, err := m.Reply("♻️ Restarting...")
	if err != nil {
		gologging.Error("Failed to send restart message: " + err.Error())
	}

	exePath, err := os.Executable()
	if err != nil {
		utils.EOR(mystic, "❌ Failed to get executable path: "+err.Error())
		return tg.EndGroup
	}

	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		utils.EOR(mystic, "❌ Failed to resolve executable path: "+err.Error())
		return tg.EndGroup
	}

	for _, chatID := range core.GetAllRoomIDs() {
		if r, _ := core.GetRoom(chatID); r != nil {
			r.Stop()
			m.Client.SendMessage(chatID, utils.MentionHTML(core.BUser)+" has just restarted herself. Sorry for the issues.\n\nStart playing again after 10–15 seconds.")
		}
	}

	utils.EOR(mystic, "✅ Restart initiated successfully!")

	os.RemoveAll("downloads")

	if err := syscall.Exec(exePath, os.Args, os.Environ()); err != nil {
		utils.EOR(mystic, "❌ Failed to restart: "+err.Error())
	}

	return tg.EndGroup
}
