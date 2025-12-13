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
	"sync"
	"time"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

var (
	autoLeaveMu      sync.Mutex
	autoLeaveRunning bool
	limit            = 50
)

func init() {
	helpTexts["autoleave"] = fmt.Sprintf(`<i>Automatically makes the assistant leave inactive or unnecessary chats every 10 minutes.</i>

<u>Usage:</u>
<b>/autoleave </b>— Shows current auto-leave status (enabled/disabled).  
<b>/autoleave enable</b> — Enable auto-leave mode.  
<b>/autoleave disable</b> — Disable auto-leave mode.

<b>🧠 Details:</b>
Once enabled, the bot checks all joined groups/channels every <b>10 minutes</b> and leaves up to <b>%d chats per cycle</b> that are not in the active room  list.

<b>⚠️ Restrictions:</b>
This command can only be used by <b>owners</b> or <b>sudo users</b>.`, limit)
}

func autoLeaveHandler(m *tg.NewMessage) error {
	args := strings.Fields(m.Text())
	chatID := m.ChannelID()

	currentState, err := database.GetAutoLeave()
	if err != nil {
		m.Reply(F(chatID, "autoleave_fetch_fail"))
		return tg.ErrEndGroup
	}

	status := F(chatID, utils.IfElse(currentState, "enabled", "disabled"))

	if len(args) < 2 {
		m.Reply(F(chatID, "autoleave_status", locales.Arg{
			"cmd":    getCommand(m),
			"action": status,
		}))
		return tg.ErrEndGroup
	}

	newState, err := utils.ParseBool(args[1])
	if err != nil {
		m.Reply(F(chatID, "invalid_bool"))
		return tg.ErrEndGroup
	}

	if newState == currentState {
		m.Reply(F(chatID, "autoleave_already", locales.Arg{
			"action": status,
		}))
		return tg.ErrEndGroup
	}

	if err := database.SetAutoLeave(newState); err != nil {
		m.Reply(F(chatID, "autoleave_update_fail"))
		return tg.ErrEndGroup
	}

	newStatus := F(chatID, utils.IfElse(newState, "enabled", "disabled"))
	m.Reply(F(chatID, "autoleave_updated", locales.Arg{
		"action": newStatus,
	}))

	autoLeaveMu.Lock()
	if newState {
		if !autoLeaveRunning {
			go startAutoLeave()
		}
	} else {
		autoLeaveRunning = false
	}
	autoLeaveMu.Unlock()

	return tg.ErrEndGroup
}

func startAutoLeave() {
	autoLeaveMu.Lock()
	if autoLeaveRunning {
		autoLeaveMu.Unlock()
		return
	}
	autoLeaveRunning = true
	autoLeaveMu.Unlock()

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		autoLeaveMu.Lock()
		if !autoLeaveRunning {
			autoLeaveMu.Unlock()
			return
		}
		autoLeaveMu.Unlock()

		activeRooms := make(map[int64]struct{})
		for _, id := range core.GetAllRoomIDs() {
			activeRooms[id] = struct{}{}
		}

		core.Assistants.ForEach(func(a *core.Assistant) {
			if a == nil || a.Client == nil {
				return
			}
			go autoLeaveAssistant(a, activeRooms, limit)
		})

		<-ticker.C
	}
}

func autoLeaveAssistant(
	ass *core.Assistant,
	activeRooms map[int64]struct{},
	limit int,
) {
	leaveCount := 0

	err := ass.Client.IterDialogs(func(d *tg.TLDialog) error {
		if d.IsUser() {
			return nil
		}

		chatID, err := utils.GetPeerID(ass.Client, d.Peer)
		if err != nil {
			gologging.Error("[Autoleave] Peer error: " + err.Error())
			return nil
		}

		if chatID == 0 || chatID == config.LoggerID {
			return nil
		}

		if _, ok := activeRooms[chatID]; ok {
			return nil
		}

		if err := ass.Client.LeaveChannel(chatID); err != nil {
			if strings.Contains(err.Error(), "USER_NOT_PARTICIPANT") ||
				strings.Contains(err.Error(), "CHANNEL_PRIVATE") {
				return nil
			}

			logger.WarnF(
				"AutoLeave (Assistant %d) failed to leave %d: %v",
				ass.Index, chatID, err,
			)
			return nil
		}

		leaveCount++
		logger.InfoF(
			"AutoLeave: Assistant %d left %d (%d/%d)",
			ass.Index, chatID, leaveCount, limit,
		)

		time.Sleep(3 * time.Second)

		if leaveCount >= limit {
			return tg.ErrStopIteration
		}

		return nil
	}, &tg.DialogOptions{
		Limit: 0,
	})

	if err != nil && err != tg.ErrStopIteration {
		logger.WarnF(
			"AutoLeave: IterDialogs error (assistant %d): %v",
			ass.Index, err,
		)
	}
}
