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

	"main/internal/config"
	"main/internal/core"
	state "main/internal/core/models"
	"main/internal/locales"
	"main/internal/platforms"
	"main/internal/utils"
	"main/ntgcalls"
)

type RecCache struct {
	Tracks []*state.Track
	Index  int
}

func getRecCache(r *core.RoomState) (*RecCache, bool) {
	ok, v := r.GetData("rec_cache")
	if !ok {
		return nil, false
	}

	cache, ok := v.(*RecCache)
	return cache, ok
}

func setRecCache(r *core.RoomState, tracks []*state.Track, start int) {
	r.SetData("rec_cache", &RecCache{
		Tracks: tracks,
		Index:  start,
	})
}

func nextCachedRec(r *core.RoomState) *state.Track {
	cache, ok := getRecCache(r)
	if !ok || cache == nil {
		return nil
	}

	if cache.Index >= len(cache.Tracks) {
		r.DeleteData("rec_cache")
		return nil
	}

	t := cache.Tracks[cache.Index]
	r.SetData("rec_cache", &RecCache{
		Tracks: cache.Tracks,
		Index:  cache.Index + 1,
	})

	return t
}

func streamEndHandler(
	chatID int64,
	streamType ntgcalls.StreamType,
	_ ntgcalls.StreamDevice,
) {
	if streamType == ntgcalls.VideoStream {
		gologging.Debug("[onStreamEndHandler] Video stream ended, returning")
		return
	}

	gologging.DebugF("[onStreamEndHandler] Stream ended in chat %d", chatID)
	ass, err := core.Assistants.ForChat(chatID)
	if err != nil {
		gologging.ErrorF("Failed to get Assistant for %d: %v", chatID, err)
		return
	}
	r, ok := core.GetRoom(chatID, ass)
	if !ok {
		return
	}

	if ok, v := r.GetData("is_transitioning"); ok {
		if ok, v := v.(bool); ok && v {
			return
		}
	}

	r.SetData("is_transitioning", true)
	defer r.DeleteData("is_transitioning")

	cid := r.EffectiveChatID()
	r.Parse()

	var t *state.Track
	var wasLooping bool
	if len(r.Queue()) == 0 && r.Loop() == 0 {
		if r.Autoplay() {
			gologging.DebugF(
				"[onStreamEndHandler] AutoPlay is ON for chat %d",
				chatID,
			)

			t = nextCachedRec(r)
			if t != nil {
				gologging.DebugF(
					"[onStreamEndHandler] Found next track in cache: %s",
					t.Title,
				)
				t.Requester = "AutoPlay"
				r.PrepareForAutoPlay()
			}

			if t == nil {
				lastTrack := r.Track()
				if lastTrack != nil {
					if lastTrack.Requester == "AutoPlay" {
						gologging.DebugF(
							"[onStreamEndHandler] AutoPlay batch finished for chat %d, stopping as per request",
							chatID,
						)
					} else {
						gologging.DebugF(
							"[onStreamEndHandler] Cache empty, fetching recommendations for: %s",
							lastTrack.Title,
						)

						p := platforms.GetPlatform(lastTrack.Source)
						if p != nil && p.CanGetRecommendations() {

							recs, err := p.GetRecommendations(lastTrack)
							if err == nil && len(recs) > 0 {
								gologging.DebugF(
									"[onStreamEndHandler] Fetched %d new recommendations",
									len(recs),
								)

								if config.QueueLimit > 0 && len(recs) > config.QueueLimit {
									recs = recs[:config.QueueLimit]
									gologging.DebugF("[onStreamEndHandler] Truncated recommendations to %d (QueueLimit)", config.QueueLimit)
								}

								setRecCache(r, recs, 1)

								// play first
								t = recs[0]
								t.Requester = "AutoPlay"
								r.PrepareForAutoPlay()

							} else {
								gologging.ErrorF("[onStreamEndHandler] recommendation error: %v", err)
							}
						} else {
							gologging.DebugF("[onStreamEndHandler] Platform %s does not support recommendations", lastTrack.Source)
						}
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
		wasLooping = r.Loop() > 0
		t = r.NextTrack()
	}

	statusMsg := F(cid, "stream_downloading_next")
	if wasLooping && t != nil && r.FilePath() != "" {
		statusMsg = F(cid, "cb_replaying")
	}

	mystic, err := core.Bot.SendMessage(
		cid,
		statusMsg,
	)
	if err != nil {
		gologging.ErrorF("[call.go] Failed to send msg: %v", err)
	}

	var filePath string
	if wasLooping && t != nil && r.FilePath() != "" {
		filePath = r.FilePath()
	} else {
		filePath, err = platforms.Download(context.Background(), t, mystic)
	}

	if err != nil {
		gologging.ErrorF(
			"[onStreamEndHandler] Download failed for %s: %v",
			t.URL,
			err,
		)
		utils.EOR(mystic, F(cid, "stream_download_fail", locales.Arg{
			"error": err.Error(),
		}))
		core.DeleteRoom(chatID)

		return
	}

	if err := r.Play(t, filePath, true); err != nil {
		gologging.ErrorF(
			"[onStreamEndHandler] Play failed for %s: %v",
			t.URL,
			err,
		)
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
