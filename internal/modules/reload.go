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
	"errors"
	"fmt"
	"slices"
	"time"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/core"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func init() {
	helpTexts["/reload"] = `<i>Reload admin cache and refresh voice chat state.</i>

<u>Usage:</u>
<b>/reload</b> — Refresh all cached data

<b>🔄 What Gets Reloaded:</b>
• Chat admin list
• Voice chat status
• Assistant presence status
• Assistant ban status (unbanned and rejoined when possible)

<b>🔒 Flood Protection:</b>
• Regular users: 5 minute cooldown
• Admins: 2 minute cooldown

<b>💡 When to Use:</b>
• After promoting/demoting admins
• Voice chat issues
• Permission problems
• Bot behaving incorrectly

<b>⚠️ Notes:</b>
• May reset room state if admin permissions required
`
}

func reloadHandler(c *td.Client, m *td.Message) error {
	if !isSuperGroup(c, m) {
		return nil
	}
	return handleReload(c, m, false)
}

func creloadHandler(c *td.Client, m *td.Message) error {
	if !isSuperGroup(c, m) || !filterAuthUsers(c, m) {
		return nil
	}
	return handleReload(c, m, true)
}

func handleReload(c *td.Client, m *td.Message, cplay bool) error {
	r, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		_, err := m.ReplyText(c, err.Error(), nil)
		return err
	}

	chatID, roomID := m.ChatID(), r.ChatID

	floodKey := fmt.Sprintf("reload:%d%d", roomID, m.SenderID())
	if remaining := utils.GetFlood(floodKey); remaining > 0 {
		_, err := m.ReplyText(c, F(chatID, "flood_minutes", locales.Arg{
			"duration": utils.FormatDuration(int(remaining.Seconds())),
		}), nil)
		return err
	}

	statusMsg, err := m.ReplyText(c, F(chatID, "reload_start"), nil)
	if err != nil {
		return err
	}

	summary := ""
	admins, err := utils.RefreshChatAdmin(c, roomID)
	if err != nil {
		summary += reloadLine("Admin cache", "❌", "("+utils.EscapeHTML(err.Error())+")")
	} else {
		summary += reloadLine("Admin cache", "✅", "Done")
	}

	isAdmin := slices.Contains(admins, m.SenderID())
	utils.SetFlood(floodKey, utils.IfElse(isAdmin, 2*time.Minute, 5*time.Minute))

	summary += reloadAssistant(roomID)

	if isAdmin && core.Rooms.Delete(roomID) {
		summary += reloadLine("Room", "✅", "Reset")
	}

	_, _ = utils.EOR(c, statusMsg, F(chatID, "reload_done", locales.Arg{
		"summary": summary,
	}), &td.EditMessageTextOpts{ParseMode: td.ParseModeHTML})
	return nil
}

// reloadAssistant refreshes the assistant's membership and the voice chat
// state, returning the summary lines for both. A banned assistant is unbanned
// through the bot (it cannot see the voice chat or join while banned), and an
// assistant that is not in the chat is joined.
func reloadAssistant(roomID int64) string {
	cs, err := core.ChatStateFor(roomID)
	if err != nil {
		return reloadLine("Assistant", "❌", "("+utils.EscapeHTML(err.Error())+")")
	}

	snap, err := cs.Refresh()
	if err != nil {
		text := "(" + utils.EscapeHTML(err.Error()) + ")"
		if errors.Is(err, core.ErrAdminPermissionRequired) {
			text = "Admin permission required"
		}
		return reloadLine("Voice chat", "❌", text) + reloadLine("Assistant", "❌", text)
	}

	if snap.AssistantBanned {
		if err := cs.Unban(); err != nil {
			// The voice chat cannot be queried without the assistant, so the
			// cached state is shown instead.
			utils.Log.Errorf("reload: failed to unban assistant in %d: %v", roomID, err)
			return reloadVoice(snap) + reloadLine("Assistant", "❌", "Banned")
		}
		snap.AssistantPresent = false
	}

	assistant := ""
	if !snap.AssistantPresent {
		if err := cs.Join(); err != nil {
			utils.Log.Errorf("reload: failed to join assistant in %d: %v", roomID, err)
			assistant = reloadLine("Assistant", "❌", "("+utils.EscapeHTML(err.Error())+")")
		} else {
			snap.AssistantPresent = true
			if fresh, err := cs.Refresh(); err == nil {
				snap = fresh
			}
		}
	}
	if assistant == "" {
		assistant = reloadLine("Assistant", "✅", "Present")
		if !snap.AssistantPresent {
			assistant = reloadLine("Assistant", "⚪", "Not present")
		}
	}

	return reloadVoice(snap) + assistant
}

func reloadVoice(snap core.StateSnapshot) string {
	if snap.VoiceChatActive != nil && *snap.VoiceChatActive {
		return reloadLine("Voice chat", "✅", "Active")
	}
	return reloadLine("Voice chat", "⚪", "Inactive")
}

func reloadLine(name, icon, text string) string {
	return fmt.Sprintf("<b>• %s:</b> <i>%s %s</i>\n", name, icon, text)
}
