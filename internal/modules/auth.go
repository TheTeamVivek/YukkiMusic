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
	"fmt"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/addauth"] = fmt.Sprintf(`<i>Grant permission to a regular user to control playback and other admin-level features without making them a Telegram admin.</i>

<u>Usage:</u>
<b>/addauth [reply to user]</b> — Add a user by replying to their message.  
<b>/addauth &lt;user_id / username&gt;</b> — Add a user directly by ID or @username.

<b>⚙️ Notes:</b>
• Only <b>chat admins</b> can use this command.  
• Auth users can control playback with commands like <code>/pause</code>, <code>/resume</code>, <code>/skip</code>, <code>/seek</code>, <code>/mute</code>, etc.  
• 🤖 Bots cannot be added as auth users.  
• 🔢 You can have up to <b>%d</b> auth users per chat.  
• 👑 The <b>Bot Owner</b>, <b>Assistant</b>, and all <b>Sudoers</b> are <b>already authorized by default</b> — they do not appear in the list and cannot be removed.

For related commands, see <code>/delauth</code> and <code>/authlist</code>.`, config.MaxAuthUsers)

	helpTexts["/delauth"] = `<i>Revoke permission from a user who was previously authorized to control playback.</i>

<u>Usage:</u>
<b>/delauth [reply to user]</b> — Remove by replying to their message.  
<b>/delauth &lt;user_id / username&gt; </b>— Remove by ID or @username.

<b>⚙️ Notes:</b>
• Only <b>chat admins</b> can use this command.  
• Use this to revoke access from misbehaving users.  
• To check who’s currently authorized, use <code>/authlist</code>.`

	helpTexts["/authlist"] = `<u>Usage:</u>
<b>/authlist</b> - <i>Displays all users currently authorized to control playback in this chat.</i>

<b>⚙️ Notes:</b>
• Anyone in the chat can use this command.  
• Shows only manually added auth users — the Owner, Assistant, and Sudoers are not listed but are always authorized.`
}

func addAuthHandler(m *telegram.NewMessage) error {
	chatID := m.ChannelID()

	if m.Args() == "" && !m.IsReply() {
		m.Reply(F(chatID, "auth_no_user", locales.Arg{
			"cmd": getCommand(m),
		}))
		return telegram.ErrEndGroup
	}

	if au, _ := database.GetAuthUsers(chatID); len(au) >= config.MaxAuthUsers {
		m.Reply(F(chatID, "auth_limit_reached", locales.Arg{
			"limit": config.MaxAuthUsers,
		}))
		return telegram.ErrEndGroup
	}

	userID, err := utils.ExtractUser(m)
	if err != nil {
		m.Reply(F(chatID, "user_extract_fail", locales.Arg{
			"error": err.Error(),
		}))
		return telegram.ErrEndGroup
	}

	// owner, bot, self, already auth, or admin — all treated the same
	if userID == config.OwnerID || userID == core.BUser.ID || userID == m.SenderID() {
		m.Reply(F(chatID, "cannot_authorize_user"))
		return telegram.ErrEndGroup
	}

	if ok, _ := database.IsAuthUser(chatID, userID); ok {
		m.Reply(F(chatID, "already_authed"))
		return telegram.ErrEndGroup
	}

	if ok, _ := utils.IsChatAdmin(m.Client, chatID, userID); ok {
		m.Reply(F(chatID, "addauth_user_is_admin"))
		return telegram.ErrEndGroup
	}

	user, err := m.Client.GetUser(userID)
	if err != nil || user == nil {
		m.Reply(F(chatID, "user_extract_fail", locales.Arg{
			"error": utils.IfElse(err != nil, err.Error(), ""),
		}))
		return telegram.ErrEndGroup
	}

	if user.Bot {
		m.Reply(F(chatID, "addauth_bot_user"))
		return telegram.ErrEndGroup
	}

	if err := database.AddAuthUser(chatID, userID); err != nil {
		m.Reply(F(chatID, "addauth_add_fail", locales.Arg{
			"error": err.Error(),
		}))
		return telegram.ErrEndGroup
	}

	uname := utils.MentionHTML(user)
	if user.Username != "" {
		uname += " (@" + user.Username + ")"
	}

	if au, _ := database.GetAuthUsers(chatID); len(au) > 0 {
		m.Reply(F(chatID, "addauth_success_with_count", locales.Arg{
			"user":  uname,
			"count": len(au),
			"limit": config.MaxAuthUsers,
		}))
	} else {
		m.Reply(F(chatID, "addauth_success", locales.Arg{
			"user": uname,
		}))
	}

	return telegram.ErrEndGroup
}

func delAuthHandler(m *telegram.NewMessage) error {
	chatID := m.ChannelID()

	if m.Args() == "" && !m.IsReply() {
		m.Reply(F(chatID, "auth_no_user", locales.Arg{
			"cmd": getCommand(m),
		}))
		return telegram.ErrEndGroup
	}

	userID, err := utils.ExtractUser(m)
	if err != nil {
		m.Reply(F(chatID, "user_extract_fail", locales.Arg{
			"error": utils.IfElse(err != nil, err.Error(), "unknown error"),
		}))
		return telegram.ErrEndGroup
	}

	if ok, err := database.IsAuthUser(chatID, userID); !ok && err == nil {
		m.Reply(F(chatID, "del_auth_not_authorized", nil))
		return telegram.ErrEndGroup
	}

	user, err := m.Client.GetUser(userID)
	if err != nil || user == nil {
		m.Reply(F(chatID, "user_extract_fail", locales.Arg{
			"error": utils.IfElse(err != nil, err.Error(), "unknown error"),
		}))
		return telegram.ErrEndGroup
	}

	if err := database.RemoveAuthUser(chatID, userID); err != nil {
		m.Reply(F(chatID, "del_auth_remove_fail", locales.Arg{
			"error": err.Error(),
		}))
		return telegram.ErrEndGroup
	}

	uname := utils.MentionHTML(user)
	if user.Username != "" {
		uname += " (@" + user.Username + ")"
	}

	m.Reply(F(chatID, "del_auth_success", locales.Arg{
		"user": uname,
	}))
	return telegram.ErrEndGroup
}

func authListHandler(m *telegram.NewMessage) error {
	chatID := m.ChannelID()

	authUsers, err := database.GetAuthUsers(chatID)
	if err != nil {
		m.Reply(F(chatID, "authlist_fetch_fail", locales.Arg{
			"error": err.Error(),
		}))
		return nil
	}

	if len(authUsers) == 0 {
		m.Reply(F(chatID, "authlist_empty", nil))
		return nil
	}

	mystic, err := m.Reply(F(chatID, "authlist_fetching", nil))
	if err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString(F(chatID, "authlist_header", nil) + "\n")

	for i, userID := range authUsers {
		user, err := m.Client.GetUser(userID)
		if err != nil || user == nil {
			sb.WriteString(F(chatID, "authlist_user_fail", locales.Arg{
				"index":   i + 1,
				"user_id": userID,
			}) + "\n")
			continue
		}

		uname := utils.MentionHTML(user)
		if user.Username != "" {
			uname += " (@" + user.Username + ")"
		}

		sb.WriteString(F(chatID, "authlist_user_entry", locales.Arg{
			"index":   i + 1,
			"user":    uname,
			"user_id": user.ID,
		}) + "\n")
	}

	sb.WriteString("\n" + F(chatID, "authlist_total", locales.Arg{
		"count": len(authUsers),
	}))

	utils.EOR(mystic, sb.String())
	return telegram.ErrEndGroup
}
