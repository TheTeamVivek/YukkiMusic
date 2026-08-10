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

	"yukkimusic/internal/logger"

	td "github.com/AshokShau/gotdbot"
	tg "github.com/amarnathcjd/gogram/telegram"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

// autoLeaveSvc periodically makes the assistants leave chats that are not playing.
var autoLeaveSvc = &autoLeaveService{
	limit:         50,
	interval:      10 * time.Minute,
	preLeaveDelay: 3 * time.Second,
}

type autoLeaveService struct {
	mu   sync.Mutex
	stop chan struct{} // non-nil => loop is running

	limit         int
	interval      time.Duration
	preLeaveDelay time.Duration
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
		autoLeaveSvc.limit,
	)
}

func autoLeaveHandler(c *td.Client, m *td.Message) error {
	if !checkSudo(c, m) {
		return nil
	}
	chatID := m.ChatID()

	current, err := database.AutoLeave()
	if err != nil {
		_, err := m.ReplyText(c, F(chatID, "autoleave_fetch_fail"), nil)
		return err
	}

	args := strings.Fields(m.Text())
	if len(args) < 2 {
		_, err :=
			// no argument => show current status
			m.ReplyText(c, F(chatID, "autoleave_status", locales.Arg{
				"cmd":    getCommand(m),
				"action": F(chatID, utils.IfElse(current, "enabled", "disabled")),
			}), nil)
		return err
	}

	enabled, err := utils.ParseBool(args[1])
	if err != nil {
		_, err := m.ReplyText(c, F(chatID, "invalid_bool"), nil)
		return err
	}

	action := F(chatID, utils.IfElse(enabled, "enabled", "disabled"))

	if enabled == current {
		_, err := m.ReplyText(c, F(chatID, "autoleave_already", locales.Arg{
			"action": action,
		}), nil)
		return err
	}

	if err := database.SetAutoLeave(enabled); err != nil {
		_, err := m.ReplyText(c, F(chatID, "autoleave_update_fail"), nil)
		return err
	}
	if _, err := m.ReplyText(c, F(chatID, "autoleave_updated", locales.Arg{
		"action": action,
	}), nil); err != nil {
		return err
	}

	autoLeaveSvc.SetEnabled(enabled)
	return nil
}

// Start enables the loop if autoleave is on in the database.
func (s *autoLeaveService) Start() {
	on, err := database.AutoLeave()
	if err == nil && on {
		s.SetEnabled(true)
	}
}

// SetEnabled starts or stops the background loop.
func (s *autoLeaveService) SetEnabled(on bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if on == (s.stop != nil) {
		return // already in the requested state
	}

	if on {
		s.stop = make(chan struct{})
		go s.loop(s.stop)
		return
	}

	close(s.stop)
	s.stop = nil
}

func (s *autoLeaveService) loop(stop chan struct{}) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			s.runCycle()
		}
	}
}

// runCycle scans every assistant's chats and leaves the inactive ones.
func (s *autoLeaveService) runCycle() {
	activeRooms := core.GetAllRooms()
	core.Assistants.ForEach(func(a *core.Assistant) {
		if a == nil || a.Client == nil {
			return
		}
		go s.leaveInactiveChats(a, activeRooms)
	})
}

func (s *autoLeaveService) leaveInactiveChats(
	ass *core.Assistant,
	activeRooms map[int64]*core.RoomState,
) {
	left := 0

	err := ass.Client.IterDialogs(func(d *tg.TLDialog) error {
		chatID := d.GetChannelID()
		if d.IsUser() || chatID == 0 || chatID == config.LoggerID {
			return nil
		}
		if _, playing := activeRooms[chatID]; playing {
			return nil
		}

		time.Sleep(s.preLeaveDelay)
		if err := ass.Client.LeaveChannel(chatID); err != nil {
			if wait := tg.GetFloodWait(err); wait > 0 {
				logger.Errorf("FloodWait detected (%ds). Sleeping...", wait)
				time.Sleep(time.Duration(wait) * time.Second)
				return nil
			}
			if !isLeaveSafeError(err) {
				logger.Warnf(
					"AutoLeave (Assistant %d) failed to leave %d: %v",
					ass.Index, chatID, err,
				)
			}
			return nil
		}

		left++
		logger.Infof(
			"AutoLeave: Assistant %d left %d (%d/%d)",
			ass.Index, chatID, left, s.limit,
		)
		if left >= s.limit {
			return tg.ErrStopIteration
		}
		return nil
	}, &tg.DialogOptions{})

	if err != nil && err != tg.ErrStopIteration {
		logger.Warnf(
			"AutoLeave: IterDialogs error (assistant %d): %v",
			ass.Index, err,
		)
	}
}

// isLeaveSafeError reports leave errors that can be safely ignored.
func isLeaveSafeError(err error) bool {
	return strings.Contains(err.Error(), "USER_NOT_PARTICIPANT") ||
		strings.Contains(err.Error(), "CHANNEL_PRIVATE")
}
