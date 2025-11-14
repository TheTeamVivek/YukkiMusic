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
	"github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

func startHandler(m *telegram.NewMessage) error {
	if m.ChatType() != telegram.EntityUser {
		database.AddServed(m.ChannelID())
		m.Reply(
			F(m.ChannelID(), "start_group"),
		)
		return telegram.EndGroup
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

		if _, err := m.RespondMedia(config.StartImage, telegram.MediaOptions{
			Caption:     caption,
			NoForwards:  true,
			ReplyMarkup: core.GetStartMarkup(),
		}); err != nil {
			gologging.Error("Error sending start media: " + err.Error())
			return err
		}
	}

	return telegram.EndGroup
}

func startCB(cb *telegram.CallbackQuery) error {
	opt := &telegram.CallbackOptions{Alert: true}

	chatID, err := getCbChatID(cb)
	if err != nil {
		gologging.ErrorF("PeerID error for %v", err)
		cb.Answer(FWithLang(config.DefaultLang, "chat_not_recognized"), opt)
		return telegram.EndGroup
	}
	cb.Answer("")
	caption := F(chatID, "start_private", locales.Arg{
		"user": utils.MentionHTML(cb.Sender),
		"bot":  utils.MentionHTML(core.BUser),
	})

	sendOpt := &telegram.SendOptions{
		ReplyMarkup: core.GetStartMarkup(),
		NoForwards:  true,
	}

	if config.StartImage != "" {
		sendOpt.Media = config.StartImage
	}

	cb.Edit(caption, sendOpt)
	return telegram.EndGroup
}
