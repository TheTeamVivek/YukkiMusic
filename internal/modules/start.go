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
	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/start"] = `<i>Start the bot and show main menu.</i>`
}

func startHandler(m *tg.NewMessage) error {
	if m.ChatType() != tg.EntityUser {
		database.AddServedChat(m.ChannelID())
		m.Reply(
			F(m.ChannelID(), "start_group"),
		)
		return tg.ErrEndGroup
	}

	arg := m.Args()
	database.AddServedUser(m.ChannelID())

	if arg != "" {
		gologging.Info(
			"Got Start parameter: " + arg + " in ChatID: " + utils.IntToStr(
				m.ChannelID(),
			),
		)
	}

	switch arg {
	case "pm_help":
		gologging.Info("User requested help via start param")
		helpHandler(m)

	default:
		caption := F(m.ChannelID(), "start_private", locales.Arg{
			"user": utils.MentionHTML(m.Sender),
			"bot":  utils.MentionHTML(m.Client.Me()),
		})

		if err := sendStartResponse(m, caption); err != nil {
			return err
		}
	}

	if config.LoggerID != 0 && isLoggerEnabled() {
		uName := "N/A"
		if m.Sender.Username != "" {
			uName = "@" + m.Sender.Username
		}

		msg := F(config.LoggerID, "logger_bot_started", locales.Arg{
			"mention":       utils.MentionHTML(m.Sender),
			"user_id":       m.SenderID(),
			"user_username": uName,
		})

		var err error
		_, err = m.Client.SendMessage(config.LoggerID, msg)
		if err != nil {
			gologging.ErrorF("[start] logger send failed: %v", err)
		}
	}

	return tg.ErrEndGroup
}

func sendStartResponse(m *tg.NewMessage, caption string) error {
	sendOpt := &tg.SendOptions{ReplyMarkup: core.GetStartMarkup(m.ChannelID())}
	if effectID := config.GetRandomEffectID(); effectID != 0 {
		sendOpt.Effect = effectID
	}
	if startImage := config.GetRandomStartImage(); startImage != "" {
		sendOpt.Media = startImage
	}

	_, err := m.Respond(caption, sendOpt)
	if err != nil && sendOpt.Effect != 0 && tg.MatchError(err, "EFFECT_ID_INVALID") {
		gologging.WarnF("[start] invalid effect id %d, retrying without effect", sendOpt.Effect)
		sendOpt.Effect = 0
		_, err = m.Respond(caption, sendOpt)
	}
	if err == nil {
		return nil
	}
	if sendOpt.Media == "" {
		gologging.ErrorF("[start] text send failed: %v", err)
		return err
	}

	gologging.ErrorF("[start] image send failed: %v", err)
	sendOpt.Media = ""
	_, textErr := m.Respond(caption, sendOpt)
	if textErr != nil {
		gologging.ErrorF("[start] text send failed: %v", textErr)
		return textErr
	}
	return nil
}

func startCB(cb *tg.CallbackQuery) error {
	cb.Answer("")

	caption := F(cb.ChannelID(), "start_private", locales.Arg{
		"user": utils.MentionHTML(cb.Sender),
		"bot":  utils.MentionHTML(cb.Client.Me()),
	})

	sendOpt := &tg.SendOptions{
		ReplyMarkup: core.GetStartMarkup(cb.ChannelID()),
	}

	cb.Edit(caption, sendOpt)
	return tg.ErrEndGroup
}
