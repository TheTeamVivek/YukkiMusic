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
	"strings"
	"sync"
	"time"

	td "github.com/AshokShau/gotdbot"
	"yukkimusic/internal/logger"

	"yukkimusic/internal/core"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
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

func handleMaintenance(c *td.Client, m *td.Message) error {
	if !checkOwner(c, m) {
		return nil
	}
	args := strings.Fields(m.Text())
	chatID := m.ChatID()
	current, err := database.IsMaintenanceEnabled()
	if err != nil {
		_, _ = m.ReplyText(
			c,
			F(chatID, "maint_check_fail", locales.Arg{"error": err.Error()}),
			nil,
		)
		return nil
	}

	if len(args) < 2 {
		return showMaintenanceStatus(c, m, current)
	}

	enable, err := utils.ParseBool(args[1])
	if err != nil {
		_, _ = m.ReplyText(c, F(chatID, "invalid_bool"), nil)
		return nil
	}

	reason := strings.Join(args[2:], " ")
	if current == enable {
		return handleSameMaintenanceState(c, m, enable, reason)
	}

	return applyMaintenanceState(c, m, enable, reason)
}

func showMaintenanceStatus(c *td.Client, m *td.Message, current bool) error {
	chatID := m.ChatID()
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
	_, _ = m.ReplyText(c, F(chatID, "maint_usage", locales.Arg{
		"cmd":    getCommand(m),
		"status": status,
	}), nil)
	return nil
}

func handleSameMaintenanceState(
	c *td.Client,
	m *td.Message,
	enable bool,
	reason string,
) error {
	chatID := m.ChatID()
	if !enable {
		_, _ = m.ReplyText(c, F(chatID, "maint_already_disabled"), nil)
		return nil
	}

	oldReason, _ := database.MaintenanceReason()
	switch {
	case reason == oldReason:
		_, _ = m.ReplyText(c, F(chatID, "maint_already_reason_same"), nil)
	case reason == "" && oldReason != "":
		_ = database.SetMaintenance(true, "")
		_, _ = m.ReplyText(c, F(chatID, "maint_reason_removed"), nil)
	case reason != "" && reason != oldReason:
		_ = database.SetMaintenance(true, reason)
		_, _ = m.ReplyText(
			c,
			F(chatID, "maint_reason_updated", locales.Arg{"reason": reason}),
			nil,
		)
	default:
		_, _ = m.ReplyText(c, F(chatID, "maint_already_enabled"), nil)
	}
	return nil
}

func applyMaintenanceState(c *td.Client, m *td.Message, enable bool, reason string) error {
	chatID := m.ChatID()
	database.SetMaintenance(enable, reason)
	logger.Infof(
		"User %d set maintenance: %v (reason: %s)",
		m.SenderID(),
		enable,
		reason,
	)

	maintCancel.Lock()
	maintCancel.cancel = !enable
	maintCancel.Unlock()

	if enable {
		go notifyMaintenanceStart(c, reason)
		msgKey := "maint_enabled"
		args := locales.Arg{}
		if reason != "" {
			msgKey = "maint_enabled_reason"
			args["reason"] = reason
		}
		_, _ = m.ReplyText(c, F(chatID, msgKey, args), nil)
	} else {
		_, _ = m.ReplyText(c, F(chatID, "maint_disabled"), nil)
	}

	return nil
}

func notifyMaintenanceStart(c *td.Client, reason string) {
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
		_, _ = c.SendTextMessage(chatID, msg, nil)
		time.Sleep(time.Second)
	}
}
