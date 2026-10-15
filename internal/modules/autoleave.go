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
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	td "github.com/AshokShau/gotdbot"
	tg "github.com/amarnathcjd/gogram/telegram"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	"yukkimusic/internal/database"
	"yukkimusic/internal/utils"
)

const (
	autoLeaveInterval = 10 * time.Minute
	autoLeaveLimit    = 50
	autoLeaveDelay    = 3 * time.Second
)

// autoLeaveSvc makes the assistants periodically leave chats that are not playing.
var autoLeaveSvc = &autoLeaveService{}

type autoLeaveService struct {
	enabled atomic.Bool
}

func init() {
	helpTexts["autoleave"] = fmt.Sprintf(
		`<i>Automatically makes the assistant leave inactive or unnecessary chats every 10 minutes.</i>

<u>Usage:</u>
<b>/autoleave </b>— Shows current auto-leave status (enabled/disabled).  
<b>/autoleave enable</b> — Enable auto-leave mode.  
<b>/autoleave disable</b> — Disable auto-leave mode.

<b>🧠 Details:</b>
Once enabled, the bot checks all joined groups/channels every <b>10 minutes</b> and leaves up to <b>%d chats per cycle</b> that are not in the active room  list.

<b>⚠️ Restrictions:</b>
This command can only be used by <b>owners</b> or <b>sudo users</b>.`,
		autoLeaveLimit,
	)
}

func autoLeaveHandler(c *td.Client, m *td.Message) error {
	if !checkSudo(c, m) {
		return nil
	}
	chatID := m.ChatID()

	args := strings.Fields(m.Text())
	if len(args) < 2 {
		return sendDashboard(c, m, "autoleave")
	}

	current, err := database.AutoLeave()
	if err != nil {
		_, err := m.ReplyText(c, F(chatID, "autoleave_fetch_fail"), nil)
		return err
	}

	enabled, err := utils.ParseBool(args[1])
	if err != nil {
		_, err := m.ReplyText(c, F(chatID, "invalid_bool"), nil)
		return err
	}

	if enabled != current {
		if err := database.SetAutoLeave(enabled); err != nil {
			_, err := m.ReplyText(c, F(chatID, "autoleave_update_fail"), nil)
			return err
		}
		autoLeaveSvc.SetEnabled(enabled)
	}

	return sendDashboard(c, m, "autoleave")
}

// Start restores the auto-leave state from the database and launches the loop.
func (s *autoLeaveService) Start() {
	if on, err := database.AutoLeave(); err == nil {
		s.enabled.Store(on)
	}
	go s.loop()
}

func (s *autoLeaveService) SetEnabled(on bool) {
	s.enabled.Store(on)
}

func (s *autoLeaveService) loop() {
	ticker := time.NewTicker(autoLeaveInterval)
	defer ticker.Stop()

	for range ticker.C {
		if !s.enabled.Load() {
			continue
		}

		// Scan every assistant's chats and leave the inactive ones.
		activeRooms := core.Rooms.All()
		core.Assistants.ForEach(func(a *core.Assistant) {
			if a == nil || a.Client == nil {
				return
			}
			go s.leaveInactiveChats(a, activeRooms)
		})
	}
}

func (s *autoLeaveService) leaveInactiveChats(
	ass *core.Assistant,
	activeRooms map[int64]*core.Room,
) {
	left := 0

	err := ass.Client.IterDialogs(func(d *tg.TLDialog) error {
		if !s.enabled.Load() {
			return tg.ErrStopIteration
		}

		chatID := d.GetChannelID()
		if d.IsUser() || chatID == 0 || chatID == config.LoggerID {
			return nil
		}
		if _, playing := activeRooms[chatID]; playing {
			return nil
		}

		time.Sleep(autoLeaveDelay)
		if err := ass.Client.LeaveChannel(chatID); err != nil {
			if wait := tg.GetFloodWait(err); wait > 0 {
				utils.Log.Errorf("FloodWait detected (%ds). Sleeping...", wait)
				time.Sleep(time.Duration(wait) * time.Second)
				return nil
			}
			if !isLeaveSafeError(err) {
				utils.Log.Warnf(
					"AutoLeave (Assistant %d) failed to leave %d: %v",
					ass.Index, chatID, err,
				)
			}
			return nil
		}

		left++
		utils.Log.Infof(
			"AutoLeave: Assistant %d left %d (%d/%d)",
			ass.Index, chatID, left, autoLeaveLimit,
		)
		if left >= autoLeaveLimit {
			return tg.ErrStopIteration
		}
		return nil
	}, &tg.DialogOptions{})

	if err != nil && err != tg.ErrStopIteration {
		utils.Log.Warnf(
			"AutoLeave: IterDialogs error (assistant %d): %v",
			ass.Index,
			err,
		)
	}
}

// isLeaveSafeError reports leave errors that can be safely ignored.
func isLeaveSafeError(err error) bool {
	return strings.Contains(err.Error(), "USER_NOT_PARTICIPANT") ||
		strings.Contains(err.Error(), "CHANNEL_PRIVATE")
}
