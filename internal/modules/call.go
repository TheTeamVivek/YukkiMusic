/*
  - This file is part of YukkiMusic.
    *

  - YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
  - Copyright (C) 2025 TheTeamVivek
    *
  - This program is free software: you can redistribute it and/or modify
  - it under the terms of the GNU General Public License as published by
  - the Free Software Foundation, either version 3 of the License, or
  - (at your option) any later version.
    *
  - This program is distributed in the hope that it will be useful,
  - but WITHOUT ANY WARRANTY; without even the implied warranty of
  - MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
  - GNU General Public License for more details.
    *
  - You should have received a copy of the GNU General Public License
  - along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package modules

import (
	"context"
	"html"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	state "main/internal/core/models"
	"main/internal/locales"
	"main/internal/platforms"
	"main/internal/utils"
)

func onStreamEndHandler(chatID int64) {
	ass, err := core.Assistants.ForChat(chatID)
	if err != nil {
		gologging.ErrorF("Failed to get Assistant for %d: %v", chatID, err)
		return
	}
	r, ok := core.GetRoom(chatID, ass)
	if !ok {
		return
	}

	cid := r.EffectiveChatID()
	r.Parse()

	var t *state.Track
	if len(r.Queue()) == 0 && r.Loop() == 0 {
		if r.Autoplay() {
			lastTrack := r.Track()
			if lastTrack != nil {
				p := platforms.GetPlatform(lastTrack.Source)
				if p != nil && p.CanGetRecommendations() {
					recs, err := p.GetRecommendations(
						lastTrack,
					)
					if err == nil && len(recs) > 0 {
						t = recs[0]
						t.Requester = "AutoPlay"
						r.PrepareForAutoPlay()
					} else {
						gologging.ErrorF("got error: %v", err)
					}
				}
			}
		}

		if t == nil {
			core.DeleteRoom(chatID)
			core.Bot.SendMessage(cid, F(cid, "stream_queue_finished"))
			return
		}
	} else {
		t = r.NextTrack()
	}
	mystic, err := core.Bot.SendMessage(
		cid,
		F(cid, "stream_downloading_next"),
	)
	if err != nil {
		gologging.ErrorF("[call.go] Failed to send msg: %v", err)
	}

	filePath, err := platforms.Download(context.Background(), t, mystic)
	if err != nil {
		gologging.ErrorF("Download failed for %s: %v", t.URL, err)
		utils.EOR(mystic, F(cid, "stream_download_fail", locales.Arg{
			"error": err.Error(),
		}))
		core.DeleteRoom(chatID)

		return
	}

	if err := r.Play(t, filePath); err != nil {
		utils.EOR(mystic, F(cid, "stream_play_fail"))
		core.DeleteRoom(chatID)

		return
	}

	title := utils.ShortTitle(t.Title, 25)
	safeTitle := html.EscapeString(title)

	msgText := F(cid, "stream_now_playing", locales.Arg{
		"url":      t.URL,
		"title":    safeTitle,
		"duration": formatDuration(t.Duration),
		"by":       t.Requester,
	})

	opt := &telegram.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: core.GetPlayMarkup(cid, r, false),
	}

	if t.Artwork != "" && shouldShowThumb(chatID) {
		opt.Media = utils.CleanURL(t.Artwork)
	}

	mystic, _ = utils.EOR(mystic, msgText, opt)
	r.SetMystic(mystic)
}
