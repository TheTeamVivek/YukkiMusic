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
	"log"

	"github.com/amarnathcjd/gogram/telegram"

	"github.com/TheTeamVivek/YukkiMusic/config"
	"github.com/TheTeamVivek/YukkiMusic/internal/core"
	"github.com/TheTeamVivek/YukkiMusic/internal/database"
	"github.com/TheTeamVivek/YukkiMusic/internal/utils"
)

var startMSG = "⚡️Pika Pika, %s!\n⚡️  Welcome to <b>%s</b> \n🎶  I’m here to help you play, stream, and manage music right here on Telegram. 🎵"

func startHandler(m *telegram.NewMessage) error {
	if m.ChatType() != telegram.EntityUser {
		database.AddServed(m.ChannelID())
		m.Reply("🎶 I'm all set!\n▶️ Drop a command to light up the chat with music.")
		return telegram.EndGroup
	}

	arg := m.Args()
	database.AddServed(m.ChannelID(), true)

	switch arg {

	case "help":
		helpHandler(m)

	default:

		caption := fmt.Sprintf(startMSG, utils.MentionHTML(m.Sender), utils.MentionHTML(core.BUser))

		if _, err := m.RespondMedia(config.StartImage, telegram.MediaOptions{
			Caption:     caption,
			NoForwards:  true,
			ReplyMarkup: core.GetStartMarkup(),
		}); err != nil {
			log.Printf("Error responding start in chat: %v", err)
			return err
		}
	}

	return telegram.EndGroup
}

func startCB(c *telegram.CallbackQuery) error {
	c.Answer("")

	caption := fmt.Sprintf(startMSG, utils.MentionHTML(c.Sender), utils.MentionHTML(core.BUser))

	opt := &telegram.SendOptions{
		ReplyMarkup: core.GetStartMarkup(),
		NoForwards:  true,
	}

	if config.StartImage != "" {
		opt.Media = config.StartImage
	}
	c.Edit(caption, opt)
	return telegram.EndGroup
}
