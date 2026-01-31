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
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/restart"] = `<i>Restart the bot process.</i>

<u>Usage:</u>
<b>/restart</b> — Restart bot

<b>⚙️ Behavior:</b>
• Stops all active rooms
• Notifies all active chats
• Restarts bot process
• Clears download cache

<b>🔒 Restrictions:</b>
• <b>Owner only</b> command

<b>⚠️ Warning:</b>
All playback will be interrupted. Bot will be offline for a few seconds.`
}

func handleRestart(m *tg.NewMessage) error {
	chatID := m.ChannelID()

	mystic, err := m.Reply(F(chatID, "restart"))
	if err != nil {
		gologging.Error("Failed to send restart message: " + err.Error())
	}

	exePath, err := os.Executable()
	if err != nil {
		utils.EOR(mystic, F(chatID, "restart_exepath_fail", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		utils.EOR(mystic, F(chatID, "restart_symlink_fail", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	for chatID := range core.GetAllRooms() {
		core.DeleteRoom(chatID)
		m.Client.SendMessage(chatID, F(chatID, "restart_service", locales.Arg{
			"bot": utils.MentionHTML(core.BUser),
		}))
		time.Sleep(time.Second)

	}

	utils.EOR(mystic, F(chatID, "restart_initiated"))

	_ = os.RemoveAll("downloads")
	_ = os.RemoveAll("cache")

	if err := syscall.Exec(exePath, os.Args, os.Environ()); err != nil {
		utils.EOR(mystic, F(chatID, "restart_fail", locales.Arg{
			"error": err.Error(),
		}))
	}

	return tg.ErrEndGroup
}
