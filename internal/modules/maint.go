/*
 * ● YukkiMusic
 * ○ A high-performance engine for streaming music in Telegram voicechats.
 * ________________________________________________________________________________________
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
 * ________________________________________________________________________________________
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

func init() {
	helpTexts["/maintenance"] = `<i>Toggle maintenance mode.</i>

<u>Usage:</u>
<b>/maintenance</b> — Show current status
<b>/maintenance on [reason]</b> — Enable maintenance
<b>/maintenance off</b> — Disable maintenance

<b>⚙️ Behavior When Active:</b>
• Stops all active rooms
• Blocks non-owner/sudo commands
• Shows maintenance message to users

<b>🔒 Restrictions:</b>
• <b>Owner only</b> command

<b>💡 Examples:</b>
<code>/maintenance on Server upgrade</code>
<code>/maintenance off</code>

<b>⚠️ Notes:</b>
• Owner and sudoers can still use bot
• All rooms are destroyed when enabled
• Users see maintenance message with reason`
	helpTexts["/maint"] = helpTexts["/maintenance"]
}

var maintCancel = struct {
	sync.Mutex
	cancel bool
}{}

func handleMaintenance(m *tg.NewMessage) error {
	args := strings.Fields(m.Text())
	chatID := m.ChannelID()
	current, err := database.IsMaintenanceEnabled()
	if err != nil {
		m.Reply(
			F(chatID, "maint_check_fail", locales.Arg{"error": err.Error()}),
		)
		return tg.ErrEndGroup
	}

	if len(args) < 2 {
		return showMaintenanceStatus(m, current)
	}

	enable, err := utils.ParseBool(args[1])
	if err != nil {
		m.Reply(F(chatID, "invalid_bool"))
		return tg.ErrEndGroup
	}

	reason := strings.Join(args[2:], " ")
	if current == enable {
		return handleSameMaintenanceState(m, enable, reason)
	}

	return applyMaintenanceState(m, enable, reason)
}

func showMaintenanceStatus(m *tg.NewMessage, current bool) error {
	chatID := m.ChannelID()
	reason, _ := database.MaintenanceReason()
	status := F(chatID, "disabled")
	if current {
		if reason != "" {
			status = F(
				chatID,
				"enabled_with_reason",
				locales.Arg{"reason": reason},
			)
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

func handleSameMaintenanceState(
	m *tg.NewMessage,
	enable bool,
	reason string,
) error {
	chatID := m.ChannelID()
	if !enable {
		m.Reply(F(chatID, "maint_already_disabled"))
		return tg.ErrEndGroup
	}

	oldReason, _ := database.MaintenanceReason()
	switch {
	case reason == oldReason:
		m.Reply(F(chatID, "maint_already_reason_same"))
	case reason == "" && oldReason != "":
		_ = database.SetMaintenance(true, "")
		m.Reply(F(chatID, "maint_reason_removed"))
	case reason != "" && reason != oldReason:
		_ = database.SetMaintenance(true, reason)
		m.Reply(
			F(chatID, "maint_reason_updated", locales.Arg{"reason": reason}),
		)
	default:
		m.Reply(F(chatID, "maint_already_enabled"))
	}
	return tg.ErrEndGroup
}

func applyMaintenanceState(m *tg.NewMessage, enable bool, reason string) error {
	chatID := m.ChannelID()
	database.SetMaintenance(enable, reason)
	gologging.InfoF(
		"User %d set maintenance: %v (reason: %s)",
		m.SenderID(),
		enable,
		reason,
	)

	maintCancel.Lock()
	maintCancel.cancel = !enable
	maintCancel.Unlock()

	if enable {
		go notifyMaintenanceStart(m.Client, reason)
		msgKey := "maint_enabled"
		args := locales.Arg{}
		if reason != "" {
			msgKey = "maint_enabled_reason"
			args["reason"] = reason
		}
		m.Reply(F(chatID, msgKey, args))
	} else {
		m.Reply(F(chatID, "maint_disabled"))
	}

	return tg.ErrEndGroup
}

func notifyMaintenanceStart(c *tg.Client, reason string) {
	for chatID := range core.GetAllRooms() {
		maintCancel.Lock()
		cancelled := maintCancel.cancel
		maintCancel.Unlock()

		if cancelled {
			break
		}

		core.DeleteRoom(chatID)
		msg := F(chatID, "maint_entering")
		if reason != "" {
			msg += "\n" + F(
				chatID,
				"maint_reason",
				locales.Arg{"reason": reason},
			)
		}
		c.SendMessage(chatID, msg)
		time.Sleep(time.Second)
	}
}
