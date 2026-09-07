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
	"math/rand"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	state "yukkimusic/internal/core/models"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/logger"
	"yukkimusic/internal/platforms"
)

func init() {
	helpTexts["/autoplay"] = `<i>Automatically play recommended songs when the queue ends.</i>

<u>Usage:</u>
<b>/autoplay</b> — Open the autoplay control panel

<b>⚙️ Behavior:</b>
• When enabled, once the queue finishes the bot keeps playing similar recommended tracks from YouTube.
• Recommendations are based on the last played song and skip tracks already in the queue.
• Autoplay is tied to the current session and resets once playback stops.

<b>🔒 Restrictions:</b>
• Only <b>chat admins</b> or <b>authorized users</b> can use this`
}

func autoplayHandler(c *td.Client, m *td.Message) error {
	if !isSuperGroup(c, m) || !filterAuthUsers(c, m) {
		return nil
	}

	chatID := m.ChatID()
	r, ok := core.RoomFor(chatID)
	if !ok || !r.Active() {
		_, err := m.ReplyText(c, F(chatID, "autoplay_no_active"), nil)
		return err
	}

	text, markup := autoplayMenu(chatID, r.Autoplay())
	_, err := m.ReplyText(c, text, &td.SendTextMessageOpts{
		ParseMode:   "HTML",
		ReplyMarkup: markup,
	})
	return err
}

func autoplayCallbackHandler(c *td.Client, u *td.UpdateNewCallbackQuery) error {
	chatID := u.ChatId

	r, ok := core.RoomFor(chatID)
	if !ok || !r.Active() {
		_ = u.Answer(c, 0, true, F(chatID, "autoplay_no_active"), "")
		return nil
	}

	if !canUseAdminCommand(c, chatID, u.SenderUserId) {
		mode, err := database.GetAdminMode(chatID)
		if err == nil && mode == database.AdminModeAdminsOnly {
			u.Answer(c, 0, true, F(chatID, "only_admin_cb"), "")
		} else {
			u.Answer(c, 0, true, F(chatID, "only_admin_or_auth_cb"), "")
		}
		return nil
	}

	enabled := !r.Autoplay()
	r.SetAutoplay(enabled)

	statusKey := "autoplay_disabled"
	if enabled {
		statusKey = "autoplay_enabled"
	}
	_ = u.Answer(c, 0, false, F(chatID, statusKey), "")

	text, markup := autoplayMenu(chatID, enabled)
	_, err := u.EditMessageText(c, text, &td.EditTextMessageOpts{
		ParseMode:   "HTML",
		ReplyMarkup: markup,
	})
	return err
}

func autoplayMenu(chatID int64, enabled bool) (string, *td.ReplyMarkupInlineKeyboard) {
	state := "disabled"
	if enabled {
		state = "enabled"
	}

	stateText := F(chatID, state)
	text := F(chatID, "autoplay_menu", locales.Arg{
		"state": stateText,
	})

	markup := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: F(chatID, "autoplay_btn", locales.Arg{
						"state": stateText,
					}),
					Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("autoplay:toggle")},
				},
			},
			{
				{
					Text: F(chatID, "CLOSE_BTN"),
					Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte("close")},
				},
			},
		},
	}

	return text, markup
}

// pickAutoplayTrack resolves a recommended track to keep playback going once
// the user queue is empty. It returns nil when autoplay is disabled, the last
// track is not from YouTube, or no suitable recommendation was found.
func pickAutoplayTrack(r *core.RoomState, last *state.Track) *state.Track {
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
		logger.Warnf(
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

	requester := F(r.ChatID, "autoplay_requester")
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

	logger.Infof("[autoplay] all %d recommendations were unsuitable", len(candidates))
	return nil
}
