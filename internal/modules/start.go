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
	"html"
	"strings"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
)

func startHandler(c *td.Client, m *td.Message) error {
	if isChannel(c, m) {
		return nil
	}

	if !m.IsPrivate() {
		database.AddServedChat(m.ChatID())
		_, err := m.ReplyText(c, F(m.ChatID(), "start_group"), nil)
		return err
	}

	database.AddServedUser(m.ChatID())
	sender, _ := m.GetUser(c)

	if m.Args() == "pm_help" {
		if err := helpHandler(c, m); err != nil {
			return err
		}
	} else if err := handlePmStart(c, m, sender); err != nil {
		return err
	}

	logStart(c, m, sender)
	return nil
}

// startRich builds the rich-formatted private start message, embedding the
// configured start image (if any) as a media block.
func startRich(c *td.Client, chatID int64, sender *td.User, fallbackID int64) *td.InputRichMessage {
	text := F(chatID, "start_private", locales.Arg{
		"user": mention(sender, fallbackID),
		"bot":  td.Mention(c.Me.FirstName, c.Me.Id, true, true),
	})

	if image := config.StartImage(); image != "" {
		text = `<figure><img src="` + html.EscapeString(image) + `"/></figure>` + "\n" + text
	}

	return &td.InputRichMessage{
		Source: &td.RichMessageSourceHtml{Text: text},
	}
}

func handlePmStart(c *td.Client, m *td.Message, sender *td.User) error {
	chatID := m.ChatID()

	_, err := c.SendRichMessage(chatID, startRich(c, chatID, sender, m.SenderID()), &td.SendTextMessageOpts{
		ReplyMarkup: core.GetStartMarkup(chatID),
	})
	return err
}

func startCB(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	cb.Answer(c, 0, false, "", "")

	msg, err := cb.GetMessage(c)
	if err != nil {
		return err
	}

	chatID := cb.ChatId
	user, _ := c.GetUser(cb.SenderUserId)

	rich := &td.InputMessageRichMessage{Message: startRich(c, chatID, user, cb.SenderUserId)}

	_, err = c.EditContent(chatID, msg.Id, rich, core.GetStartMarkup(chatID))
	return err
}

func isPhotoMessage(m *td.Message) bool {
	_, ok := m.Content.(*td.MessagePhoto)
	return ok
}

// isChannelPost reports whether m came from a broadcast channel.
func isChannel(c *td.Client, m *td.Message) bool {
	chat, err := m.GetChat(c)
	if err != nil {
		return false
	}
	sg, ok := chat.Type.(*td.ChatTypeSupergroup)
	return ok && sg.IsChannel
}

// mention builds an HTML mention for a gotdbot user, falling back to a plain
// "user" mention when the user is nil.
func mention(user *td.User, fallbackId int64) string {
	name := ""
	id := fallbackId
	if user != nil {
		id = user.Id
		name = user.FirstName + " " + user.LastName
	}

	if name = strings.TrimSpace(name); name == "" {
		name = "user"
	}
	return td.Mention(name, id, true, true)
}

// logStart reports a bot start to the configured logger chat, if enabled.
func logStart(c *td.Client, m *td.Message, sender *td.User) {
	if config.LoggerID == 0 || !isLoggerEnabled() || m.SenderID() == config.OwnerID {
		return
	}

	uName := "N/A"
	if sender != nil && sender.Usernames != nil && len(sender.Usernames.ActiveUsernames) > 0 {
		uName = "@" + sender.Usernames.ActiveUsernames[0]
	}

	msg := F(config.LoggerID, "logger_bot_started", locales.Arg{
		"mention":       mention(sender, m.SenderID()),
		"user_id":       m.SenderID(),
		"user_username": uName,
	})

	if _, err := c.SendTextMessage(config.LoggerID, msg, nil); err != nil {
		c.Logger.Error("Logger send failed", "error", err)
	}
}
