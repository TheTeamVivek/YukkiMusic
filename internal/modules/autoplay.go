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
	"errors"
	"math/rand"
	"strings"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	state "yukkimusic/internal/core/models"
	"yukkimusic/internal/platforms"
	"yukkimusic/internal/utils"
)

var errNoRoom = errors.New("no active room")

func init() {
	helpTexts["/autoplay"] = `<i>Automatically play recommended songs when the queue ends.</i>

<u>Usage:</u>
<b>/autoplay</b> — Open the autoplay control panel
<b>/autoplay enable</b> — Enable autoplay
<b>/autoplay disable</b> — Disable autoplay

<b>⚙️ Behavior:</b>
• When enabled, once the queue finishes the bot keeps playing similar recommended tracks from YouTube.
• Recommendations are based on the last played song and skip tracks already in the queue.
• Autoplay is tied to the current session and resets once playback stops.

<b>🔒 Restrictions:</b>
• Only <b>chat admins</b> or <b>authorized users</b> can use this`

	dashboards["autoplay"] = dashboardDef{
		title:       func(chatID int64) string { return F(chatID, "autoplay_title") },
		description: func(chatID int64) string { return F(chatID, "autoplay_desc") },
		canUse: func(c *td.Client, u *td.UpdateNewCallbackQuery) bool {
			if !canUseAdminCallback(c, u) {
				return false
			}
			if r, ok := core.Rooms.Get(u.ChatId); !ok || !r.Active() {
				u.Answer(c, F(u.ChatId, "room_no_active"), &td.CallbackAnswerOpts{ShowAlert: true})
				return false
			}
			return true
		},
		get: func(chatID int64) (bool, error) {
			r, ok := core.Rooms.Get(chatID)
			if !ok {
				return false, errNoRoom
			}
			return r.Autoplay(), nil
		},
		set: func(chatID int64, on bool) error {
			r, ok := core.Rooms.Get(chatID)
			if !ok {
				return errNoRoom
			}
			r.SetAutoplay(on)
			return nil
		},
	}
}

func autoplayHandler(c *td.Client, m *td.Message) error {
	if !isSuperGroup(c, m) || !filterAuthUsers(c, m) {
		return nil
	}

	chatID := m.ChatID()

	r, ok := core.Rooms.Get(chatID)
	if !ok || !r.Active() {
		_, err := m.ReplyText(c, F(chatID, "room_no_active"), nil)
		return err
	}

	if args := strings.Fields(m.Text()); len(args) > 1 {
		if enabled, err := utils.ParseBool(args[1]); err == nil && enabled != r.Autoplay() {
			r.SetAutoplay(enabled)
		}
	}

	return sendDashboard(c, m, "autoplay")
}

// pickAutoplayTrack resolves a recommended track to keep playback going once
// the user queue is empty. It returns nil when autoplay is disabled, the last
// track is not from YouTube, or no suitable recommendation was found.
func pickAutoplayTrack(r *core.Room, last *state.Track) *state.Track {
	if r == nil || last == nil || !r.Autoplay() ||
		last.Source != platforms.PlatformYouTube {
		return nil
	}

	limit := config.QueueLimit
	if limit <= 0 {
		limit = 10
	}

	candidates, err := platforms.AutoplayTracks(last, limit)
	if err != nil || len(candidates) == 0 {
		utils.Log.Warnf(
			"[autoplay] no recommendations for %s: %v",
			last.ID,
			err,
		)
		return nil
	}

	seen := make(map[string]struct{}, limit+1)
	seen[last.ID] = struct{}{}
	for _, t := range r.Queue() {
		if t == nil {
			continue
		}
		seen[t.ID] = struct{}{}
		seen[t.URL] = struct{}{}
	}

	requester := F(r.ControlChatID, "autoplay_requester")
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	for _, t := range candidates {
		if t == nil || t.ID == "" ||
			config.DurationLimit > 0 && t.Duration > config.DurationLimit {
			continue
		}
		if _, dup := seen[t.ID]; dup {
			continue
		}
		t.Requester = requester
		return t
	}

	utils.Log.Infof("[autoplay] all %d recommendations were unsuitable", len(candidates))
	return nil
}
