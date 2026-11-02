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
	"strconv"
	"strings"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func init() {
	helpTexts["/addauth"] = fmt.Sprintf(
		`<i>Grant permission to a regular user to control playback and other admin-level features without making them a Telegram admin.</i>

<u>Usage:</u>
<b>/addauth [reply to user]</b> — Add a user by replying to their message.  
<b>/addauth &lt;user_id / username&gt;</b> — Add a user directly by ID or @username.

<b>⚙️ Notes:</b>
• Only <b>chat admins</b> can use this command.  
• Auth users can control playback with commands like <code>/pause</code>, <code>/resume</code>, <code>/skip</code>, <code>/seek</code>, <code>/mute</code>, etc.  
• 🤖 Bots cannot be added as auth users.  
• 🔢 You can have up to <b>%d</b> auth users per chat.  
• 👑 The <b>Bot Owner</b>, <b>Assistant</b>, and all <b>Sudoers</b> are <b>already authorized by default</b> — they do not appear in the list and cannot be removed.

For related commands, see <code>/delauth</code> and <code>/authlist</code>.`,
		config.MaxAuthUsers,
	)

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

func addAuthHandler(c *td.Client, m *td.Message) error {
	chatID := m.ChatID()

	if !canManageAuth(c, chatID, m.SenderID()) {
		_, err := m.ReplyRichMessageHTML(c, F(chatID, "only_admin"), nil)
		return err
	}

	if m.Args() == "" && m.ReplyToMessageID() == 0 {
		_, err := m.ReplyText(c, F(chatID, "auth_no_user", locales.Arg{
			"cmd": getCommand(m),
		}), nil)
		return err
	}

	if au, _ := database.AuthorizedUsers(chatID); len(
		au,
	) >= config.MaxAuthUsers {
		_, err := m.ReplyText(c, F(chatID, "auth_limit_reached", locales.Arg{
			"limit": config.MaxAuthUsers,
		}), nil)
		return err
	}

	userID, err := utils.ExtractUser(c, m)
	if err != nil {
		_, err := m.ReplyText(c, F(chatID, "user_extract_fail", locales.Arg{
			"error": err.Error(),
		}), nil)
		return err
	}

	// owner, bot, self, already auth, or admin — all treated the same
	if userID == config.OwnerID || userID == c.Me().Id ||
		userID == m.SenderID() {
		_, err := m.ReplyText(c, F(chatID, "cannot_authorize_user"), nil)
		return err
	}

	if ok, _ := database.IsAuthorized(chatID, userID); ok {
		_, err := m.ReplyText(c, F(chatID, "already_authed"), nil)
		return err
	}

	if ok, _ := utils.IsChatAdmin(c, chatID, userID); ok {
		_, err := m.ReplyText(c, F(chatID, "addauth_user_is_admin"), nil)
		return err
	}

	user, err := c.GetUser(userID)
	if err != nil || user == nil {
		_, err := m.ReplyText(c, F(chatID, "user_extract_fail", locales.Arg{
			"error": utils.IfElse(err != nil, err.Error(), ""),
		}), nil)
		return err
	}

	if user.Type != nil {
		if _, isBot := user.Type.(*td.UserTypeBot); isBot {
			_, err := m.ReplyText(c, F(chatID, "addauth_bot_user"), nil)
			return err
		}
	}

	if err := database.Authorize(chatID, userID); err != nil {
		_, err := m.ReplyText(c, F(chatID, "addauth_add_fail", locales.Arg{
			"error": err.Error(),
		}), nil)
		return err
	}

	uname := mention(user, userID)
	if user.Usernames != nil && len(user.Usernames.ActiveUsernames) > 0 {
		uname += " (@" + user.Usernames.ActiveUsernames[0] + ")"
	}

	if au, _ := database.AuthorizedUsers(chatID); len(au) > 0 {
		if _, err := m.ReplyText(c, F(chatID, "addauth_success_with_count", locales.Arg{
			"user":  uname,
			"count": len(au),
			"limit": config.MaxAuthUsers,
		}), nil); err != nil {
			return err
		}
	} else {
		if _, err := m.ReplyText(c, F(chatID, "addauth_success", locales.Arg{
			"user": uname,
		}), nil); err != nil {
			return err
		}
	}

	return nil
}

func delAuthHandler(c *td.Client, m *td.Message) error {
	chatID := m.ChatID()

	if !canManageAuth(c, chatID, m.SenderID()) {
		_, err := m.ReplyRichMessageHTML(c, F(chatID, "only_admin"), nil)
		return err
	}

	if m.Args() == "" && m.ReplyToMessageID() == 0 {
		_, err := m.ReplyText(c, F(chatID, "auth_no_user", locales.Arg{
			"cmd": getCommand(m),
		}), nil)
		return err
	}

	userID, err := utils.ExtractUser(c, m)
	if err != nil {
		_, err := m.ReplyText(c, F(chatID, "user_extract_fail", locales.Arg{
			"error": utils.IfElse(err != nil, err.Error(), "unknown error"),
		}), nil)
		return err
	}

	if ok, err := database.IsAuthorized(chatID, userID); !ok && err == nil {
		_, err := m.ReplyText(c, F(chatID, "del_auth_not_authorized", nil), nil)
		return err
	}

	user, _ := c.GetUser(userID)

	if err := database.Unauthorize(chatID, userID); err != nil {
		_, err := m.ReplyText(c, F(chatID, "del_auth_remove_fail", locales.Arg{
			"error": err.Error(),
		}), nil)
		return err
	}

	var uname string
	if user != nil {
		uname = mention(user, userID)
		if user.Usernames != nil && len(user.Usernames.ActiveUsernames) > 0 {
			uname += " (@" + user.Usernames.ActiveUsernames[0] + ")"
		}
	} else {
		uname = "User (<code>" + strconv.FormatInt(userID, 10) + "</code>)"
	}
	_, rerr := m.ReplyText(c, F(chatID, "del_auth_success", locales.Arg{
		"user": uname,
	}), nil)
	return rerr
}

func authListHandler(c *td.Client, m *td.Message) error {
	chatID := m.ChatID()

	authUsers, err := database.AuthorizedUsers(chatID)
	if err != nil {
		_, err := m.ReplyText(c, F(chatID, "authlist_fetch_fail", locales.Arg{
			"error": err.Error(),
		}), nil)
		return err
	}

	_, err = m.ReplyRichMessage(c, authRich(c, chatID, authUsers), nil)
	return err
}

// authRich builds the rich "authorized users" card. Each row carries a danger
// button that revokes that user's authorization and refreshes the card.
func authRich(c *td.Client, chatID int64, users []int64) *td.InputRichMessage {
	var b strings.Builder

	fmt.Fprintf(&b, "<h3>%s</h3>\n", F(chatID, "authlist_title"))

	if len(users) == 0 {
		fmt.Fprintf(&b, "<p>%s</p>\n", F(chatID, "authlist_empty"))
		return &td.InputRichMessage{
			Source: &td.RichMessageSourceHtml{Text: b.String()},
		}
	}

	fmt.Fprintf(&b, "<p>%s</p>\n", F(chatID, "authlist_count", locales.Arg{
		"count": len(users),
		"limit": config.MaxAuthUsers,
	}))
	b.WriteString("<table bordered striped compact>\n")
	fmt.Fprintf(
		&b,
		"<tr><th>%s</th><th>%s</th><th>\U0001F5D1\uFE0F</th></tr>\n",
		F(chatID, "authlist_col_user"),
		F(chatID, "authlist_col_id"),
	)

	for _, userID := range users {
		name := F(chatID, "authlist_unknown_user")
		if user, err := c.GetUser(userID); err == nil && user != nil {
			name = mention(user, userID)
			if user.Usernames != nil && len(user.Usernames.ActiveUsernames) > 0 {
				name += " (@" + user.Usernames.ActiveUsernames[0] + ")"
			}
		}
		fmt.Fprintf(
			&b,
			"<tr><td>%s</td><td><code>%d</code></td>"+
				"<td><tg-button type=\"callback_data\" style=\"danger\" data=\"authdel:%d:%d\">\u2a2f</tg-button></td></tr>\n",
			name,
			userID,
			chatID,
			userID,
		)
	}
	b.WriteString("</table>\n")
	fmt.Fprintf(&b, "<footer>%s</footer>\n", F(chatID, "authlist_hint"))

	return &td.InputRichMessage{
		Source: &td.RichMessageSourceHtml{Text: b.String()},
	}
}

func authDeleteHandler(c *td.Client, u *td.UpdateNewCallbackQuery) error {
	chatID := u.ChatId

	parts := strings.Split(u.DataString(), ":")
	if len(parts) != 3 || parts[0] != "authdel" {
		utils.Log.Warnf("Invalid auth delete payload: %s", u.DataString())
		u.Answer(c, F(chatID, "invalid_request"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	targetID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		u.Answer(c, F(chatID, "invalid_request"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	if !canManageAuth(c, chatID, u.SenderUserId) {
		u.Answer(c, F(chatID, "only_admin_cb"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	if ok, err := database.IsAuthorized(chatID, targetID); !ok || err != nil {
		u.Answer(c, F(chatID, "del_auth_not_authorized"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	if err := database.Unauthorize(chatID, targetID); err != nil {
		u.Answer(c, F(chatID, "del_auth_remove_fail", locales.Arg{
			"error": err.Error(),
		}), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	users, err := database.AuthorizedUsers(chatID)
	if err != nil {
		users = nil
	}

	if _, err := c.EditContent(u.ChatId, u.MessageId, &td.InputMessageRichMessage{
		Message: authRich(c, chatID, users),
	}, nil); err != nil {
		utils.Log.Errorf("Failed to edit auth list: %v", err)
	}

	u.Answer(c, F(chatID, "authlist_delete_success"), &td.CallbackAnswerOpts{ShowAlert: true})
	return nil
}
