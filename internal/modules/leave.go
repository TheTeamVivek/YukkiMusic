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
	"github.com/amarnathcjd/gogram/telegram"

	"github.com/TheTeamVivek/YukkiMusic/config"
	"github.com/TheTeamVivek/YukkiMusic/internal/core"
	"github.com/TheTeamVivek/YukkiMusic/internal/database"
	"github.com/TheTeamVivek/YukkiMusic/internal/utils"
)

var (
	autoLeaveCtx    context.Context
	autoLeaveCancel context.CancelFunc
	autoLeaveMu     sync.Mutex
	limit           = 30
)

func autoLeaveHandler(m *telegram.NewMessage) error {
	args := strings.Fields(m.Text())

	currentState, err := database.GetAutoLeave()
	if err != nil {
		m.Reply("⚠️ <b>Failed to fetch AutoLeave state.</b>")
		return telegram.EndGroup
	}

	status := "❌ Disabled"
	if currentState {
		status = "✅ Enabled"
	}

	if len(args) < 2 {
		m.Reply(fmt.Sprintf("🏃‍♂️ <b>AutoLeave Control</b>\n\nUsage: %s [enable|disable]\nCurrent state: "+status, getCommand(m)))
		return telegram.EndGroup
	}

	mystic, err := m.Reply("⚙️ <b>Updating AutoLeave...</b>")
	if err != nil {
		return err
	}

	newState, err := utils.ParseBool(args[1])
	if err != nil {
		utils.EOR(mystic, "⚠️ <b>Invalid value.</b>\nUse 'enable' or 'disable'.")
		return telegram.EndGroup
	}

	if newState == currentState {
		utils.EOR(mystic, "⚠️ AutoLeave is already "+status)
		return telegram.EndGroup
	}

	if err := database.SetAutoLeave(newState); err != nil {
		utils.EOR(mystic, "⚠️ <b>Failed to update AutoLeave state.</b>")
		return telegram.EndGroup
	}

	newStatus := "❌ Disabled"
	if newState {
		newStatus = "✅ Enabled"
	}

	utils.EOR(mystic, "🏃‍♂️ AutoLeave has been "+newStatus)

	autoLeaveMu.Lock()
	defer autoLeaveMu.Unlock()

	if newState {
		go startAutoLeave()
	} else if autoLeaveCancel != nil {
		autoLeaveCancel()
		autoLeaveCtx = nil
		autoLeaveCancel = nil
	}

	return telegram.EndGroup
}

func startAutoLeave() {
	autoLeaveMu.Lock()
	if autoLeaveCtx != nil {
		autoLeaveMu.Unlock()
		return
	}
	autoLeaveCtx, autoLeaveCancel = context.WithCancel(context.Background())
	autoLeaveMu.Unlock()

	ticker := time.NewTicker(120 * time.Second)
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

			dialogCh, errCh := core.UBot.IterDialogs(&telegram.DialogOptions{Limit: int32(limit * 2)})
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
					case <-time.After(1 * time.Second):
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

		case <-autoLeaveCtx.Done():
			return
		}
	}
}
