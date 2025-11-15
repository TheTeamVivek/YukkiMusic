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
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

var (
	autoLeaveCtx    context.Context
	autoLeaveCancel context.CancelFunc
	autoLeaveMu     sync.Mutex
	limit           = 30
)

func init() {
	helpTexts["autoleave"] = fmt.Sprintf(`<i>Automatically makes the bot leave inactive or unnecessary chats every 15 minutes.</i>

<u>Usage:</u>
<b>/autoleave </b>— Shows current auto-leave status (enabled/disabled).  
<b>/autoleave enable</b> — Enable auto-leave mode.  
<b>/autoleave disable</b> — Disable auto-leave mode.

<b>🧠 Details:</b>
Once enabled, the bot checks all joined groups/channels every <b>15 minutes</b> and leaves up to <b>%d chats per cycle</b> that are not in the active room list.

<b>⚠️ Restrictions:</b>
This command can only be used by <b>owners</b> or <b>sudo users</b>.`, limit)
}

func autoLeaveHandler(m *tg.NewMessage) error {
	args := strings.Fields(m.Text())
	chatID := m.ChatID()

	currentState, err := database.GetAutoLeave()
	if err != nil {
		m.Reply(F(chatID, "autoleave_fetch_fail"))
		return tg.EndGroup
	}

	status := F(chatID, utils.IfElse(currentState, "enabled", "disabled"))

	if len(args) < 2 {
		m.Reply(F(chatID, "autoleave_status", locales.Arg{
			"cmd":    getCommand(m),
			"action": status,
		}))
		return tg.EndGroup
	}

	newState, err := utils.ParseBool(args[1])
	if err != nil {
		m.Reply(F(chatID, "invalid_bool"))
		return tg.EndGroup
	}

	if newState == currentState {
		m.Reply(F(chatID, "autoleave_already", locales.Arg{
			"action": status,
		}))
		return tg.EndGroup
	}

	if err := database.SetAutoLeave(newState); err != nil {
		m.Reply(F(chatID, "autoleave_update_fail"))
		return tg.EndGroup
	}

	newStatus := F(chatID, utils.IfElse(newState, "enabled", "disabled"))
	m.Reply(F(chatID, "autoleave_updated", locales.Arg{
		"action": newStatus,
	}))

	autoLeaveMu.Lock()
	defer autoLeaveMu.Unlock()

	if newState {
		go startAutoLeave()
	} else if autoLeaveCancel != nil {
		autoLeaveCancel()
		autoLeaveCtx = nil
		autoLeaveCancel = nil
	}
	return tg.EndGroup
}

func startAutoLeave() {
	autoLeaveMu.Lock()
	if autoLeaveCtx != nil {
		autoLeaveMu.Unlock()
		return
	}
	autoLeaveCtx, autoLeaveCancel = context.WithCancel(context.Background())
	autoLeaveMu.Unlock()

	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	exists := make(map[int64]struct{})

	for {
		select {
		case <-ticker.C:
			// Refresh set of valid room IDs
			for k := range exists {
				delete(exists, k)
			}
			for _, chatID := range core.GetAllRoomIDs() {
				exists[chatID] = struct{}{}
			}
			iterCtx, iterCancel := context.WithCancel(context.Background())
			dialogCh, errCh := core.UBot.IterDialogs(&tg.DialogOptions{Limit: int32(limit * 3), Context: iterCtx})
			leaveCount := 0

		loop:
			for {
				select {
				case d, ok := <-dialogCh:
					if !ok {
						break loop // No more dialogs, break inner loop
					}

					select {
					case <-autoLeaveCtx.Done():
						return // Exit if context was cancelled
					default:
					}

					chatID, err := utils.GetPeerID(core.UBot, d.Peer)
					if err != nil {
						gologging.Error("[Autoleave] Failed to get peer, Error: " + err.Error())
						continue

					}

					if chatID == 0 || chatID == config.LoggerID || chatID > 0 {
						continue
					}
					if _, ok := exists[chatID]; ok {
						continue
					}

					if err := core.UBot.LeaveChannel(chatID); err != nil {
						if strings.Contains(err.Error(), "USER_NOT_PARTICIPANT") || strings.Contains(err.Error(), "CHANNEL_PRIVATE") {
							continue
						}
						logger.WarnF("AutoLeave: failed to leave chat %d: %v", chatID, err)
						continue
					}

					leaveCount++
					logger.InfoF("AutoLeave: left chat %d (%d/%d)", chatID, leaveCount, limit)

					select {
					case <-time.After(2500 * time.Millisecond):
					case <-autoLeaveCtx.Done():
						return
					}

					if leaveCount >= limit {
						break loop
					}

				case err, ok := <-errCh:
					if !ok {
						break loop
					}
					logger.WarnF("AutoLeave: IterDialogs error: %v", err)
					break loop

				case <-autoLeaveCtx.Done():
					return
				}
			}
			iterCancel()

		case <-autoLeaveCtx.Done():
			return
		}
	}
}
