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
 */

package modules

import (
	"strings"
	"sync"
	"time"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

var maintCancel = struct {
	sync.Mutex
	cancel bool
}{}

func handleMaintenance(m *tg.NewMessage) error {
	args := strings.Fields(m.Text())
	chatID := m.ChannelID()
	current, err := database.IsMaintenance()
	if err != nil {
		m.Reply(F(chatID, "maint_check_fail", locales.Arg{"error": err.Error()}))
		return tg.ErrEndGroup
	}

	// show current status if no args
	if len(args) < 2 {
		reason, _ := database.GetMaintReason()
		status := F(chatID, "disabled")
		if current {
			if reason != "" {
				status = F(chatID, "enabled_with_reason", locales.Arg{"reason": reason})
			} else {
				status = F(chatID, "enabled")
			}
		}
		m.Reply(F(chatID, "maint_usage", locales.Arg{
			"cmd":    getCommand(m),
			"status": status,
		}))
		return tg.ErrEndGroup
	}

	enable, err := utils.ParseBool(args[1])
	if err != nil {
		m.Reply(F(chatID, "invalid_bool"))
		return tg.ErrEndGroup
	}
	reason := strings.Join(args[2:], " ")
	oldReason, _ := database.GetMaintReason()

	// no change in state
	if current == enable {
		if enable {
			switch {
			case reason == oldReason:
				m.Reply(F(chatID, "maint_already_reason_same"))
			case reason == "" && oldReason != "":
				_ = database.SetMaintenance(true, "")
				m.Reply(F(chatID, "maint_reason_removed"))
			case reason != "" && reason != oldReason:
				_ = database.SetMaintenance(true, reason)
				m.Reply(F(chatID, "maint_reason_updated", locales.Arg{"reason": reason}))
			default:
				m.Reply(F(chatID, "maint_already_enabled"))
			}
		} else {
			m.Reply(F(chatID, "maint_already_disabled"))
		}
		return tg.ErrEndGroup
	}

	// apply new state
	database.SetMaintenance(enable, reason)
	logger.InfoF("User %d set maintenance: %v (reason: %s)", m.SenderID(), enable, reason)

	if enable {
		maintCancel.Lock()
		maintCancel.cancel = false
		maintCancel.Unlock()

		go func(c *tg.Client, reason string) {
			for _, id := range core.GetAllRoomIDs() {
				maintCancel.Lock()
				if maintCancel.cancel {
					maintCancel.Unlock()
					break
				}
				maintCancel.Unlock()
				ass, err := core.Assistants.ForChat(id)
				if err != nil {
					gologging.ErrorF("Failed to get Assistant for %d: %v", id, err)
					continue
				}

				if r, ok := core.GetRoom(id, ass); ok {
					r.Destroy()
					msg := F(id, "maint_entering")
					if reason != "" {
						msg += "\n" + F(id, "maint_reason", locales.Arg{"reason": reason})
					}
					c.SendMessage(id, msg)
					time.Sleep(time.Second)
				}
			}
		}(m.Client, reason)

		args := locales.Arg{}
		if reason != "" {
			args["reason"] = reason
			m.Reply(F(chatID, "maint_enabled_reason", args))
		} else {
			m.Reply(F(chatID, "maint_enabled"))
		}
		return tg.ErrEndGroup
	}

	// disable maintenance
	maintCancel.Lock()
	maintCancel.cancel = true
	maintCancel.Unlock()

	m.Reply(F(chatID, "maint_disabled"))
	return tg.ErrEndGroup
}
