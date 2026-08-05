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
		m.ReplyText(c, F(m.ChatID(), "start_group"), nil)
		return nil
	}

	database.AddServedUser(m.ChatID())
	sender, _ := m.GetUser(c)

	var err error
	if m.Args() == "pm_help" {
		err = helpHandler(c, m)
	} else {
		err = handlePmStart(c, m, sender)
	}
	if err != nil {
		return err
	}

	logStart(c, m, sender)
	return nil
}

func handlePmStart(c *td.Client, m *td.Message, sender *td.User) error {
	caption := F(m.ChatID(), "start_private", locales.Arg{
		"user": mentionOf(sender, m.SenderID()),
		"bot":  td.Mention(c.Me.FirstName, c.Me.Id, true, true),
	})

	if image := config.StartImage(); image != "" {
		if _, err := m.ReplyPhoto(c, td.InputFileRemote{Id: image}, &td.SendPhotoOpts{
			Caption:     caption,
			ReplyMarkup: core.GetStartMarkup(m.ChatID()),
		}); err == nil {
			return nil
		}
	}

	_, err := m.ReplyText(c, caption, &td.SendTextMessageOpts{
		ReplyMarkup: core.GetStartMarkup(m.ChatID()),
	})
	return err
}

func startCB(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	cb.Answer(c, 0, false, "", "")

	msg, err := cb.GetMessage(c)
	if err != nil {
		return err
	}
	sender, _ := msg.GetUser(c)

	caption := F(cb.ChatId, "start_private", locales.Arg{
		"user": mentionOf(sender, msg.SenderID()),
		"bot":  td.Mention(c.Me.FirstName, c.Me.Id, true, true),
	})

	_, err = cb.EditMessageText(c, caption, &td.EditTextMessageOpts{
		ReplyMarkup: core.GetStartMarkup(cb.ChatId),
	})
	return err
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

// mentionOf builds an HTML mention, falling back to a plain "user" mention
// if sender is nil (e.g. GetUser failed).
func mentionOf(sender *td.User, fallbackId int64) string {
	if sender == nil {
		return td.Mention("user", fallbackId, true, true)
	}
	name := strings.TrimSpace(sender.FirstName + " " + sender.LastName)
	if name == "" {
		name = "user"
	}
	return td.Mention(name, sender.Id, true, true)
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
		"mention":       mentionOf(sender, m.SenderID()),
		"user_id":       m.SenderID(),
		"user_username": uName,
	})

	if _, err := c.SendTextMessage(config.LoggerID, msg, nil); err != nil {
		c.Logger.Error("Logger send failed", "error", err)
	}
}