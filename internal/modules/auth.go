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
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"

	"github.com/TheTeamVivek/YukkiMusic/config"
	"github.com/TheTeamVivek/YukkiMusic/internal/core"
	"github.com/TheTeamVivek/YukkiMusic/internal/database"
	"github.com/TheTeamVivek/YukkiMusic/internal/utils"
)

func addAuthHandler(m *telegram.NewMessage) error {
	chatID := m.ChannelID()

	if m.Args() == "" && !m.IsReply() {
		m.Reply("⚠️ Please provide a user — use:\n" + getCommand(m) + " [user_id]</code> or reply to a user's message.")
		return telegram.EndGroup
	}

	if au, _ := database.GetAuthUsers(chatID); len(au) >= config.MaxAuthUsers {
		m.Reply("⚠️ This chat has reached the maximum authorized users limit (" + strconv.Itoa(config.MaxAuthUsers) + "). Please remove someone before adding a new one.")
		return telegram.EndGroup
	}

	userID, err := utils.ExtractUser(m)
	if err != nil {
		m.Reply("<b>⚠️ Unable to get that user:</b> <i>" + err.Error() + "</i>")
		return telegram.EndGroup
	}

	if userID == config.OwnerID {
		m.Reply("👑 The owner is already implicitly authorized — no need to add manually.")
		return telegram.EndGroup
	}

	if userID == core.BUser.ID {
		m.Reply("🤖 You cannot add the bot as an authorized user.")
		return telegram.EndGroup
	}

	if userID == m.SenderID() {
		m.Reply("⚠️ You can’t authorize yourself.")
		return telegram.EndGroup
	}

	if ok, _ := database.IsAuthUser(chatID, userID); ok {
		m.Reply("⚠️ That user is already authorized — no need to add them again.")
		return telegram.EndGroup
	}

	if ok, _ := utils.IsChatAdmin(m.Client, chatID, userID); ok {
		m.Reply("⚠️ That user is already a chat admin — adding them to the auth list isn’t necessary.")
		return telegram.EndGroup
	}

	user, err := m.Client.GetUser(userID)
	if err != nil || user == nil {
		msg := "❌ Failed to fetch user info."
		if err != nil {
			msg += " <code>" + err.Error() + "</code>"
		}
		m.Reply(msg)
		return telegram.EndGroup
	}

	if user.Bot {
		m.Reply("🤖 You can’t add a bot to the auth list — sudoers must be human!")
		return telegram.EndGroup
	}

	if err := database.AddAuthUser(chatID, userID); err != nil {
		m.Reply("❌ Failed to add authorized user: <code>" + err.Error() + "</code>")
		return telegram.EndGroup
	}

	uname := utils.MentionHTML(user)
	if user.Username != "" {
		uname += " (@" + user.Username + ")"
	}

	if au, _ := database.GetAuthUsers(chatID); len(au) > 0 {
		m.Reply(fmt.Sprintf("✅ Added %s.\nNow %d/%d authorized users.",
			uname, len(au), config.MaxAuthUsers))
	} else {
		m.Reply("✅ Successfully added " + uname + " to the authorized users list.")
	}

	return telegram.EndGroup
}

func delAuthHandler(m *telegram.NewMessage) error {
	chatID := m.ChannelID()
	if m.Args() == "" && !m.IsReply() {
		m.Reply("⚠️ Please provide a user — use:\n" + getCommand(m) + " [user_id]</code> or reply to a user's message.")
		return telegram.EndGroup
	}
	userID, err := utils.ExtractUser(m)
	if err != nil {
		m.Reply("<b>⚠️ Unable to get that user:</b> <i>" + err.Error() + "</i>")
		return telegram.EndGroup
	}

	if ok, err := database.IsAuthUser(chatID, userID); !ok && err == nil {
		m.Reply("⚠️ That user isn’t authorized — nothing to remove.")
		return telegram.EndGroup
	}

	user, err := m.Client.GetUser(userID)
	if err != nil || user == nil {
		msg := "❌ Failed to fetch user info."
		if err != nil {
			msg += " <code>" + err.Error() + "</code>"
		}
		m.Reply(msg)
		return telegram.EndGroup
	}

	if err := database.RemoveAuthUser(chatID, userID); err != nil {
		m.Reply("❌ Failed to remove authorized user: <code>" + err.Error() + "</code>")
		return telegram.EndGroup
	}

	uname := utils.MentionHTML(user)
	if user.Username != "" {
		uname += " (@" + user.Username + ")"
	}

	m.Reply("✅ Successfully removed " + uname + " from the authorized users list.")
	return telegram.EndGroup
}

func authListHandler(m *telegram.NewMessage) error {
	chatID := m.ChannelID()
	authUsers, err := database.GetAuthUsers(chatID)
	if err != nil {
		m.Reply("❌ Failed to get authorized users list: <code>" + err.Error() + "</code>")
		return nil
	}

	if len(authUsers) == 0 {
		m.Reply("ℹ️ There are no authorized users in this chat.")
		return nil
	}

	mystic, err := m.Reply("⏳ Fetching authorized users list...")
	if err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString("<b>Authorized Users:</b>\n")

	for i, userID := range authUsers {
		user, err := m.Client.GetUser(userID)
		if err != nil || user == nil {
			sb.WriteString(fmt.Sprintf("%d. <code>%d</code> (Could not fetch user info)\n", i+1, userID))
			continue
		}

		uname := utils.MentionHTML(user)
		if user.Username != "" {
			uname += " (@" + user.Username + ")"
		}

		sb.WriteString(fmt.Sprintf("%d. %s (<code>%d</code>)\n", i+1, uname, user.ID))
	}
	sb.WriteString(fmt.Sprintf("\n<b>Total:</b> %d users", len(authUsers)))

	utils.EOR(mystic, sb.String())
	return telegram.EndGroup
}
