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
	"context"
	"html"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	"main/internal/locales"
	"main/internal/platforms"
	"main/internal/utils"
)

func skipHandler(m *telegram.NewMessage) error {
	return handleSkip(m, false)
}

func cskipHandler(m *telegram.NewMessage) error {
	return handleSkip(m, true)
}

func handleSkip(m *telegram.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.EndGroup
	}

	chatID := m.ChannelID()
	if !r.IsActiveChat() {
		m.Reply(F(chatID, "room_no_active"))
		return telegram.EndGroup
	}

	mention := utils.MentionHTML(m.Sender)

	if len(r.Queue()) == 0 && r.Loop() == 0 {
		r.Destroy()
		m.Reply(F(chatID, "skip_stopped", locales.Arg{
			"user": mention,
		}))
		return telegram.EndGroup
	}

	t := r.NextTrack()

	mystic, err := core.Bot.SendMessage(chatID, F(chatID, "stream_downloading_next"))
	if err != nil {
		gologging.ErrorF("[skip.go] err: %v", err)
	}

	path, err := platforms.Download(context.Background(), t, mystic)
	if err != nil {
		txt := F(chatID, "stream_download_fail", locales.Arg{
			"error": err.Error(),
		})

		if mystic != nil {
			utils.EOR(mystic, txt)
		} else {
			core.Bot.SendMessage(chatID, txt)
		}

		r.Destroy()
		return telegram.EndGroup
	}

	if err := r.Play(t, path); err != nil {
		txt := F(chatID, "stream_play_fail")
		if mystic != nil {
			utils.EOR(mystic, txt)
		} else {
			core.Bot.SendMessage(chatID, txt)
		}
		r.Destroy()
		return telegram.EndGroup
	}

	title := utils.ShortTitle(t.Title, 25)
	safeTitle := html.EscapeString(title)

	msg := F(chatID, "stream_now_playing", locales.Arg{
		"url":      t.URL,
		"title":    safeTitle,
		"duration": formatDuration(t.Duration),
		"by":       t.Requester,
	})

	opt := &telegram.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: core.GetPlayMarkup(r, false),
	}
	if t.Artwork != "" {
		opt.Media = utils.CleanURL(t.Artwork)
	}

	var newMystic *telegram.NewMessage
	if mystic != nil {
		newMystic, _ = utils.EOR(mystic, msg, opt)
	} else {
		newMystic, _ = core.Bot.SendMessage(chatID, msg, opt)
	}

	if newMystic != nil {
		r.SetMystic(newMystic)
	}

	return telegram.EndGroup
}
