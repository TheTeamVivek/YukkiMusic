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
	"fmt"
	"html"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
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
		m.Reply("⚠️ <b>No active playback.</b>\nThere’s nothing playing right now.")
		return telegram.EndGroup
	}
	mention := utils.MentionHTML(m.Sender)
	if len(r.Queue) == 0 && r.Loop == 0 {
		r.Destroy()
		m.Reply(fmt.Sprintf("⏹️ Playback stopped. Queue is empty.\nSkipped by: %s", mention))
		return telegram.EndGroup
	}

	t := r.NextTrack()
	mystic, err := core.Bot.SendMessage(chatID, "📥 Downloading your next track...")
	if err != nil {
		gologging.ErrorF("[skip.go] Failed to send status message: %v", err)
	}

	path, err := platforms.Download(context.Background(), t, mystic)
	if err != nil {
		gologging.ErrorF("Download failed for %s: %v", t.URL, err)
		errMsg := fmt.Sprintf("❌ Failed to download next track.\nError: %v", err)
		if mystic != nil {
			utils.EOR(mystic, errMsg)
		} else {
			core.Bot.SendMessage(chatID, errMsg)
		}
		r.Destroy()
		return telegram.EndGroup
	}

	err = r.Play(t, path)
	if err != nil {
		errMsg := "❌ Failed to play song."
		if mystic != nil {
			utils.EOR(mystic, errMsg)
		} else {
			core.Bot.SendMessage(chatID, errMsg)
		}
		r.Destroy()
		return telegram.EndGroup
	}

	title := utils.ShortTitle(t.Title, 25)
	safeTitle := html.EscapeString(title)
	msgText := fmt.Sprintf(
		"<b>🎵 Now Playing:</b>\n\n<b>▫ Track:</b> <a href=\"%s\">%s</a>\n<b>▫ Duration:</b> %s\n<b>▫ Requested by:</b> %s",
		t.URL,
		safeTitle,
		formatDuration(t.Duration),
		t.BY,
	)
	opt := telegram.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: core.GetPlayMarkup(r, false),
	}
	if t.Artwork != "" {
		opt.Media = utils.CleanURL(t.Artwork)
	}

	var newMystic *telegram.NewMessage
	if mystic != nil {
		newMystic, _ = utils.EOR(mystic, msgText, opt)
	} else {
		newMystic, _ = core.Bot.SendMessage(chatID, msgText, &opt)
	}

	if newMystic != nil {
		r.SetMystic(newMystic)
	}

	return telegram.EndGroup
}
