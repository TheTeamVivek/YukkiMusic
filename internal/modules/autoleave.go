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
	limit        = 50
	autoLeaveSvc = newAutoLeaveService(limit, 10*time.Minute, 3*time.Second)
)

type autoLeaveService struct {
	mu          sync.Mutex
	loopRunning bool
	stopCh      chan struct{}

	limit         int
	interval      time.Duration
	preLeaveDelay time.Duration
}

func newAutoLeaveService(
	limit int,
	interval time.Duration,
	preLeaveDelay time.Duration,
) *autoLeaveService {
	return &autoLeaveService{
		limit:         limit,
		interval:      interval,
		preLeaveDelay: preLeaveDelay,
	}
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
		limit,
	)
}

func autoLeaveHandler(m *tg.NewMessage) error {
	args := strings.Fields(m.Text())
	chatID := m.ChannelID()

	currentState, err := database.AutoLeave()
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

	autoLeaveSvc.SetEnabled(newState)

	return tg.ErrEndGroup
}

func (s *autoLeaveService) Start() {
	enabled, err := database.AutoLeave()
	if err != nil || !enabled {
		return
	}

	s.SetEnabled(true)
}

func (s *autoLeaveService) SetEnabled(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if enabled {
		if s.loopRunning {
			return
		}
		s.loopRunning = true
		s.stopCh = make(chan struct{})
		go s.runLoop()
		return
	}

	if s.stopCh != nil {
		close(s.stopCh)
		s.stopCh = nil
	}
}

func (s *autoLeaveService) runLoop() {
	s.mu.Lock()
	stopCh := s.stopCh
	s.mu.Unlock()

	timer := time.NewTimer(s.interval)
	defer timer.Stop()

	for {
		select {
		case <-stopCh:
			s.mu.Lock()
			s.stopCh = nil
			s.loopRunning = false
			s.mu.Unlock()
			return
		case <-timer.C:
			s.runCycle()
			timer.Reset(s.interval)
		}
	}
}

func (s *autoLeaveService) runCycle() {
	activeRooms := core.GetAllRooms()
	core.Assistants.ForEach(func(a *core.Assistant) {
		if a == nil || a.Client == nil {
			return
		}
		go s.autoLeaveAssistant(a, activeRooms)
	})
}

func (s *autoLeaveService) autoLeaveAssistant(
	ass *core.Assistant,
	activeRooms map[int64]*core.RoomState,
) {
	leaveCount := 0
	err := ass.Client.IterDialogs(func(d *tg.TLDialog) error {
		if d.IsUser() {
			return nil
		}
		chatID := d.GetChannelID()

		if chatID == 0 || chatID == config.LoggerID ||
			d.GetID() == config.LoggerID {
			return nil
		}

		if _, ok := activeRooms[chatID]; ok {
			return nil
		}

		time.Sleep(s.preLeaveDelay)
		if err := ass.Client.LeaveChannel(chatID); err != nil {
			if wait := tg.GetFloodWait(err); wait > 0 {
				gologging.ErrorF(
					"FloodWait detected (%ds). Sleeping...", wait,
				)
				time.Sleep(time.Duration(wait) * time.Second)
				return nil
			}

			if strings.Contains(err.Error(), "USER_NOT_PARTICIPANT") ||
				strings.Contains(err.Error(), "CHANNEL_PRIVATE") {
				return nil
			}

			gologging.WarnF(
				"AutoLeave (Assistant %d) failed to leave %d: %v",
				ass.Index, chatID, err,
			)
			return nil
		}

		leaveCount++
		gologging.InfoF(
			"AutoLeave: Assistant %d left %d (%d/%d)",
			ass.Index, chatID, leaveCount, s.limit,
		)

		if leaveCount >= s.limit {
			return tg.ErrStopIteration
		}

		return nil
	}, &tg.DialogOptions{
		Limit: 0,
	})

	if err != nil && err != tg.ErrStopIteration {
		gologging.WarnF(
			"AutoLeave: IterDialogs error (assistant %d): %v",
			ass.Index, err,
		)
	}
}
