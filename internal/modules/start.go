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
		database.AddServed(m.ChannelID())
		m.Reply(
			F(m.ChannelID(), "start_group"),
		)
		return tg.ErrEndGroup
	}

	arg := m.Args()
	database.AddServed(m.ChannelID(), true)

	if arg != "" {
		gologging.Info("Got Start parameter: " + arg + " in ChatID: " + utils.IntToStr(m.ChannelID()))
	}

	switch arg {
	case "pm_help":
		gologging.Info("User requested help via start param")
		helpHandler(m)

	default:
		caption := F(m.ChannelID(), "start_private", locales.Arg{
			"user": utils.MentionHTML(m.Sender),
			"bot":  utils.MentionHTML(core.BUser),
		})

		_, err := m.RespondMedia(&tg.InputMediaWebPage{
			URL:             config.StartImage,
			ForceLargeMedia: true,
		}, &tg.MediaOptions{
			Caption:     caption,
			NoForwards:  true,
			ReplyMarkup: core.GetStartMarkup(m.ChannelID()),
		})
		if err != nil {
			gologging.Error("[start] InputMediaWebPage Reply failed: " + err.Error())

			_, err = m.RespondMedia(config.StartImage, &tg.MediaOptions{
				Caption:     caption,
				NoForwards:  true,
				ReplyMarkup: core.GetStartMarkup(m.ChannelID()),
			})
			if err != nil {
				gologging.Error("[start] URL media reply failed: " + err.Error())

				_, err = m.RespondMedia(caption, &tg.MediaOptions{
					NoForwards:  true,
					ReplyMarkup: core.GetStartMarkup(m.ChannelID()),
				})
				return err
			}
		}
	}

	if config.LoggerID != 0 && isLogger() {
		uName := "N/A"
		if m.Sender.Username != "" {
			uName = "@" + m.Sender.Username
		}
		msg := F(m.ChannelID(), "logger_bot_started", locales.Arg{
			"mention":       utils.MentionHTML(m.Sender),
			"user_id":       m.SenderID(),
			"user_username": uName,
		})
		_, err := m.Client.SendMessage(config.LoggerID, msg)
		if err != nil {
			gologging.Error("Failed to send logger_bot_started msg, Err: " + err.Error())
		}
	}
	return tg.ErrEndGroup
}

func startCB(cb *tg.CallbackQuery) error {
	cb.Answer("")

	caption := F(cb.ChannelID(), "start_private", locales.Arg{
		"user": utils.MentionHTML(cb.Sender),
		"bot":  utils.MentionHTML(core.BUser),
	})

	sendOpt := &tg.SendOptions{
		ReplyMarkup: core.GetStartMarkup(cb.ChannelID()),
		NoForwards:  true,
	}

	if config.StartImage != "" {
		sendOpt.Media = config.StartImage
	}

	cb.Edit(caption, sendOpt)
	return tg.ErrEndGroup
}
