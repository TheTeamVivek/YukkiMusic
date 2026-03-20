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
	"fmt"
	"time"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["bug"] = `<i>Report a bug, issue, or unexpected behavior directly to the bot developers.</i>

<u>Usage:</u>
<b>/bug &lt;description&gt;</b> — Send a bug report with a short explanation.  
<b>Reply + /bug</b> — Report a specific message or media as a bug.

<b>🧠 Details:</b>
When used, the bot automatically forwards your report (and the replied message if any) to the <b>owner</b> and <b>logger channels</b>.  
Flood protection is applied — you can only send one report every <b>5 minutes</b> per chat.

<b>⚠️ Note:</b>  
Reports are logged for debugging purposes only. Misuse (like spam) may restrict your access to this command.`
}

func bugHandler(m *telegram.NewMessage) error {
	chatID := m.ChannelID()
	reason := m.Args()

	if reason == "" && !m.IsReply() {
		m.Reply(F(chatID, "bug_usasge", locales.Arg{
			"cmd": getCommand(m),
		}))
		return telegram.ErrEndGroup
	}

	// Flood control
	key := fmt.Sprintf("room:%d:%d", m.SenderID(), m.ChannelID())
	if remaining := utils.GetFlood(key); remaining > 0 {
		m.Reply(F(chatID, "flood_minutes", locales.Arg{
			"minutes": formatDuration(int(remaining.Seconds())),
		}))
		return telegram.ErrEndGroup
	}
	utils.SetFlood(key, 5*time.Minute)

	// Forward the replied message if any
	if m.IsReply() {
		if config.LoggerID != 0 {
			m.Client.Forward(config.LoggerID, m.Peer, []int32{m.ReplyID()})
		}
		if config.OwnerID != 0 {
			m.Client.Forward(config.OwnerID, m.Peer, []int32{m.ReplyID()})
		}
	}

	userMention := utils.MentionHTML(m.Sender)
	chatTitle := "Private Chat"
	if m.Channel != nil {
		chatTitle = m.Channel.Title
	}
	chatMention := fmt.Sprintf("<a href=\"%s\">%s</a>", m.Link(), chatTitle)

	reportMsg := F(chatID, "bug_report_format", locales.Arg{
		"user":    userMention,
		"user_id": m.Sender.ID,
		"chat":    chatMention,
		"chat_id": m.ChannelID(),
		"report":  reason,
	})

	// Send report to dev channels
	if config.LoggerID != 0 && (reason != "" || m.IsReply()) {
		m.Client.SendMessage(config.LoggerID, reportMsg)
	}
	if config.OwnerID != 0 && (reason != "" || m.IsReply()) {
		m.Client.SendMessage(config.OwnerID, reportMsg)
	}
	m.Reply(F(chatID, "bug_thanks"))
	return telegram.ErrEndGroup
}
