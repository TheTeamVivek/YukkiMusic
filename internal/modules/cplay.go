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
	"strconv"
	"strings"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

// setCPlayHandler handles /setcplay <chat_id | @username | username | linked | off>.
func setCPlayHandler(c *td.Client, m *td.Message) error {
	if !isSuperGroup(c, m) || !filterAuthUsers(c, m) {
		return nil
	}

	args := strings.Fields(m.Text())
	if len(args) < 2 {
		replyKey(c, m, "cplay_usage")
		return nil
	}

	chatID := m.ChatID()
	var channelID int64 // 0 turns channel play off
	doneKey := "cplay_disabled"

	if on, err := utils.ParseBool(args[1]); err != nil || on {
		id, key := resolveChannel(c, chatID, args[1])
		if key == "" {
			key = checkBotRights(c, id)
		}
		if key != "" {
			replyKey(c, m, key)
			return nil
		}
		channelID, doneKey = id, "cplay_enabled"
	}

	if err := database.LinkChannel(chatID, channelID); err != nil {
		utils.Log.Errorf("cplay: failed to save for chat %d: %v", chatID, err)
		replyKey(c, m, "cplay_save_error")
		return nil
	}
	_, err := m.ReplyText(c, F(chatID, doneKey, locales.Arg{"channel_id": channelID}), nil)
	return err
}

// resolveChannel finds the channel named by arg. On failure it returns the
// locale key explaining why.
func resolveChannel(c *td.Client, chatID int64, arg string) (int64, string) {
	var chat *td.Chat
	var err error

	switch id, parseErr := strconv.ParseInt(arg, 10, 64); {
	case strings.EqualFold(arg, "linked"):
		return linkedChannel(c, chatID)
	case parseErr == nil:
		chat, err = c.GetChat(id)
	default:
		chat, err = c.SearchPublicChat(strings.TrimPrefix(arg, "@"))
	}

	if err != nil || chat == nil {
		utils.Log.Errorf("cplay: cannot resolve %q for chat %d: %v", arg, chatID, err)
		return 0, "cplay_channel_not_accessible"
	}
	if _, ok := chat.Type.(*td.ChatTypeSupergroup); !ok {
		return 0, "cplay_invalid_target"
	}
	return chat.Id, ""
}

func linkedChannel(c *td.Client, chatID int64) (int64, string) {
	chat, err := c.GetChat(chatID)
	if err != nil {
		return 0, "cplay_resolve_peer_fail"
	}
	sg, ok := chat.Type.(*td.ChatTypeSupergroup)
	if !ok {
		return 0, "cplay_resolve_peer_fail"
	}

	full, err := c.GetSupergroupFullInfo(sg.SupergroupId)
	if err != nil || full == nil || full.LinkedChatId == 0 {
		return 0, "cplay_channel_not_linked"
	}
	return full.LinkedChatId, ""
}

// checkBotRights returns the locale key of the problem if the bot is not an
// admin of the channel with the right to invite users, else "".
func checkBotRights(c *td.Client, channelID int64) string {
	me := c.Me()
	if me == nil || me.Id == 0 {
		var err error
		if me, err = c.GetMe(); err != nil || me == nil {
			return "cplay_channel_not_accessible"
		}
	}

	member, err := c.GetChatMember(channelID, &td.MessageSenderUser{UserId: me.Id})
	if err != nil || member == nil {
		utils.Log.Errorf("cplay: cannot get bot member of %d: %v", channelID, err)
		return "cplay_channel_not_accessible"
	}

	switch st := member.Status.(type) {
	case *td.ChatMemberStatusCreator:
		return ""
	case *td.ChatMemberStatusAdministrator:
		if st.Rights != nil && st.Rights.CanInviteUsers {
			return ""
		}
		return "cplay_bot_invite_permission_missing"
	}
	return "cplay_channel_not_accessible"
}
