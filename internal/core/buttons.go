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

package core

import (
	"fmt"
	"strings"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

var F func(chatID int64, key string, values ...locales.Arg) string // overwritten from main.go

// Bot is the gotdbot client, set from main.go.
var Bot *td.Client

func botUsername() string {
	if Bot == nil || Bot.Me() == nil || Bot.Me().Usernames == nil {
		return ""
	}
	usernames := Bot.Me().Usernames.ActiveUsernames
	if len(usernames) == 0 {
		return ""
	}
	return usernames[0]
}

func buttonText(chatID int64, localeKey string) (string, int64) {
	lang, err := database.Language(chatID)
	if err != nil {
		lang = config.DefaultLang
	}
	return locales.GetButton(lang, localeKey, nil)
}

func urlBtn(text, url string) td.InlineKeyboardButton {
	return td.InlineKeyboardButton{
		Text: text,
		Type: &td.InlineKeyboardButtonTypeUrl{Url: url},
	}
}

func dataBtn(text, cb string) td.InlineKeyboardButton {
	return td.InlineKeyboardButton{
		Text: text,
		Type: &td.InlineKeyboardButtonTypeCallback{Data: []byte(cb)},
	}
}

func styleBtn(text, cb, colour string) td.InlineKeyboardButton {
	b := dataBtn(text, cb)

	if config.DisableColour {
		return b
	}

	switch strings.ToLower(colour) {
	case "red":
		b.Style = td.ButtonStyleDanger{}
	case "blue":
		b.Style = td.ButtonStylePrimary{}
	case "green":
		b.Style = td.ButtonStyleSuccess{}
	}

	return b
}

func DataBtnT(chatID int64, key, data string) td.InlineKeyboardButton {
	text, icon := buttonText(chatID, key)
	btn := dataBtn(text, data)
	btn.IconCustomEmojiId = icon
	return btn
}

func urlBtnT(chatID int64, key, url string) td.InlineKeyboardButton {
	text, icon := buttonText(chatID, key)
	btn := urlBtn(text, url)
	btn.IconCustomEmojiId = icon
	return btn
}

func styleBtnT(chatID int64, key, data, colour string) td.InlineKeyboardButton {
	text, icon := buttonText(chatID, key)
	btn := styleBtn(text, data, colour)
	btn.IconCustomEmojiId = icon
	return btn
}

func row(btns ...td.InlineKeyboardButton) []td.InlineKeyboardButton {
	return btns
}

func keyboard(rows ...[]td.InlineKeyboardButton) td.ReplyMarkup {
	return &td.ReplyMarkupInlineKeyboard{Rows: rows}
}

func GetCancelKeyboard(chatID int64) td.ReplyMarkup {
	return keyboard(row(DataBtnT(chatID, "DOWNLOAD_CANCEL_BTN", "cancel")))
}

func GetBroadcastCancelKeyboard(chatID int64) td.ReplyMarkup {
	return keyboard(row(DataBtnT(chatID, "BROADCAST_CANCEL_BTN", "bcast_cancel")))
}

func SuppMarkup(chatID int64) td.ReplyMarkup {
	btn := urlBtnT(chatID, "SUPPORT_BTN", config.SupportChat)
	if !config.DisableColour {
		btn.Style = td.ButtonStylePrimary{}
	}

	return keyboard(row(btn))
}

// getActionToggle returns a single-button keyboard that toggles the given
// playback action ("pause" or "mute") based on the room's current state.
func getActionToggle(chatID int64, r *Room, action string) td.ReplyMarkup {
	key := strings.ToUpper(action) + "_BTN"

	switch {
	case action == "pause" && r.Paused():
		key, action = "RESUME_BTN", "resume"
	case action == "mute" && r.Muted():
		key, action = "UNMUTE_BTN", "unmute"
	}

	return keyboard(row(DataBtnT(chatID, key, fmt.Sprintf("room:%d:%s", r.ChatID, action))))
}

// GetPauseToggle returns a pause/resume toggle keyboard for the room.
func GetPauseToggle(chatID int64, r *Room) td.ReplyMarkup {
	return getActionToggle(chatID, r, "pause")
}

// GetMuteToggle returns a mute/unmute toggle keyboard for the room.
func GetMuteToggle(chatID int64, r *Room) td.ReplyMarkup {
	return getActionToggle(chatID, r, "mute")
}

func GetStopConfirmMarkup(
	chatID int64,
	r *Room,
	isPaused bool,
) td.ReplyMarkup {
	prefix := fmt.Sprintf("room:%d:", r.ChatID)

	key, cb := "CONFIRM_UNMUTE_BTN", prefix+"unmute"

	if isPaused {
		key, cb = "CONFIRM_RESUME_BTN", prefix+"resume"
	}

	return keyboard(row(
		styleBtnT(chatID, key, cb, "green"),
		styleBtnT(chatID, "CONFIRM_STOP_BTN", prefix+"stop", "red"),
	))
}

func GetPlayMarkup(chatID int64, r *Room, queued bool) td.ReplyMarkup {
	prefix := fmt.Sprintf("room:%d:", r.ChatID)
	track := r.Track()
	duration := 0
	if track != nil {
		duration = track.Duration
	}

	progress := utils.GetProgressBar(r.Position(), duration)
	progress = utils.FormatTime(
		r.Position(),
	) + " " + progress + " " + utils.FormatTime(
		duration,
	)

	rows := make([][]td.InlineKeyboardButton, 0, 4)

	if !queued {
		rows = append(rows, []td.InlineKeyboardButton{
			dataBtn(progress, "progress"),
		})
	}
	// Toggle: show play when paused, pause while playing.
	toggle, toggleCB := "II", prefix+"pause"
	if r.Paused() {
		toggle, toggleCB = "▷", prefix+"resume"
	}

	rows = append(rows, []td.InlineKeyboardButton{
		dataBtn(toggle, toggleCB),
		dataBtn("⟳", prefix+"replay"),
		dataBtn("‣‣I", prefix+"skip"),
		dataBtn("▢", prefix+"stop"),
	})

	rows = append(rows, []td.InlineKeyboardButton{
		dataBtn(F(chatID, "CLOSE_BTN"), "close"),
	})

	return &td.ReplyMarkupInlineKeyboard{Rows: rows}
}

func GetGroupHelpKeyboard(chatID int64) td.ReplyMarkup {
	return keyboard(row(urlBtnT(chatID, "GC_HELP_BTN", "https://t.me/"+botUsername()+"?start=pm_help")))
}

func GetStartMarkup(chatID int64) td.ReplyMarkup {
	return keyboard(
		row(urlBtnT(chatID, "ADD_ME_BTN", "https://t.me/"+botUsername()+"?startgroup&admin=invite_users")),
		row(DataBtnT(chatID, "START_HELP_BTN", "help_cb")),
		row(
			urlBtnT(chatID, "UPDATES_BTN", config.SupportChannel),
			urlBtnT(chatID, "SUPPORT_BTN", config.SupportChat),
		),
		row(
			DataBtnT(chatID, "settings_btn_lang", "lang:select"),
			urlBtnT(chatID, "SOURCE_BTN", "https://github.com/TheTeamVivek/YukkiMusic"),
		),
	)
}

func GetHelpKeyboard(chatID int64) td.ReplyMarkup {
	return keyboard(
		row(
			DataBtnT(chatID, "HELP_ADMINS_BTN", "help:admins"),
			DataBtnT(chatID, "HELP_PUBLIC_BTN", "help:public"),
		),
		row(
			DataBtnT(chatID, "HELP_OWNER_BTN", "help:owner"),
			DataBtnT(chatID, "HELP_SUDOERS_BTN", "help:sudoers"),
		),
		row(styleBtnT(chatID, "BACK_BTN", "start", "")),
	)
}

func GetBackKeyboard(chatID int64) td.ReplyMarkup {
	return keyboard(row(styleBtnT(chatID, "BACK_BTN", "help:main", "blue")))
}

func GetRestartConfirmMarkup(chatID int64) td.ReplyMarkup {
	return keyboard(row(
		styleBtnT(chatID, "restart_btn_bot", "restart:bot", "red"),
		styleBtnT(chatID, "restart_btn_replay", "restart:replay", "green"),
	))
}

// GetLangKeyboard lists the available languages two per row, marking lang as
// selected. A non-empty back adds a back button with that callback data.
func GetLangKeyboard(lang, back string) td.ReplyMarkup {
	var rows [][]td.InlineKeyboardButton
	var cur []td.InlineKeyboardButton

	for _, l := range locales.GetAvailableLanguages() {
		name := locales.Get(l, "name", nil)
		if l == lang {
			name = "✔️ " + name
		}
		cur = append(cur, dataBtn(name, "lang:"+l))
		if len(cur) == 2 {
			rows = append(rows, cur)
			cur = nil
		}
	}
	if len(cur) > 0 {
		rows = append(rows, cur)
	}

	if back != "" {
		text, icon := locales.GetButton(lang, "BACK_BTN", nil)
		btn := dataBtn(text, back)
		btn.IconCustomEmojiId = icon
		rows = append(rows, row(btn))
	}

	return keyboard(rows...)
}
