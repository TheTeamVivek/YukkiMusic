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
	if Bot == nil || Bot.Me == nil || Bot.Me.Usernames == nil {
		return ""
	}
	usernames := Bot.Me.Usernames.ActiveUsernames
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

func AddMeMarkup(chatID int64) td.ReplyMarkup {
	text, iconID := buttonText(chatID, "ADD_ME_BTN")
	btn := urlBtn(text, "https://t.me/"+botUsername()+"?startgroup&admin=invite_users")
	btn.IconCustomEmojiId = iconID
	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{{btn}},
	}
}

func GetCancelKeyboard(chatID int64) td.ReplyMarkup {
	text, iconID := buttonText(chatID, "DOWNLOAD_CANCEL_BTN")
	btn := dataBtn(text, "cancel")
	btn.IconCustomEmojiId = iconID
	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{{btn}},
	}
}

func GetBroadcastCancelKeyboard(chatID int64) td.ReplyMarkup {
	text, iconID := buttonText(chatID, "BROADCAST_CANCEL_BTN")
	btn := dataBtn(text, "bcast_cancel")
	btn.IconCustomEmojiId = iconID
	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{{btn}},
	}
}

func SuppMarkup(chatID int64) td.ReplyMarkup {
	text, iconID := buttonText(chatID, "SUPPORT_BTN")
	btn := urlBtn(text, config.SupportChat)
	btn.IconCustomEmojiId = iconID
	if !config.DisableColour {
		btn.Style = td.ButtonStylePrimary{}
	}

	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{{btn}},
	}
}

func GetStopConfirmMarkup(
	chatID int64,
	r *RoomState,
	isPaused bool,
) td.ReplyMarkup {
	prefix := fmt.Sprintf("room:%d:", r.ID)

	text, cb := "CONFIRM_UNMUTE_BTN", prefix+"unmute"

	if isPaused {
		text, cb = "CONFIRM_RESUME_BTN", prefix+"resume"
	}

	resumeText, resumeIcon := buttonText(chatID, text)
	stopText, stopIcon := buttonText(chatID, "CONFIRM_STOP_BTN")

	resumeBtn := styleBtn(resumeText, cb, "green")
	resumeBtn.IconCustomEmojiId = resumeIcon

	stopBtn := styleBtn(stopText, prefix+"stop", "red")
	stopBtn.IconCustomEmojiId = stopIcon

	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{resumeBtn, stopBtn},
		},
	}
}

func GetPlayMarkup(chatID int64, r *RoomState, queued bool) td.ReplyMarkup {
	prefix := fmt.Sprintf("room:%d:", r.ID)
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
	text, iconID := buttonText(chatID, "GC_HELP_BTN")
	btn := urlBtn(text, "https://t.me/"+botUsername()+"?start=pm_help")
	btn.IconCustomEmojiId = iconID
	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{{btn}},
	}
}

func GetStartMarkup(chatID int64) td.ReplyMarkup {
	addMeText, addMeIcon := buttonText(chatID, "ADD_ME_BTN")
	addMeBtn := urlBtn(addMeText, "https://t.me/"+botUsername()+"?startgroup&admin=invite_users")
	addMeBtn.IconCustomEmojiId = addMeIcon

	helpText, helpIcon := buttonText(chatID, "START_HELP_BTN")
	helpBtn := dataBtn(helpText, "help_cb")
	helpBtn.IconCustomEmojiId = helpIcon

	updatesText, updatesIcon := buttonText(chatID, "UPDATES_BTN")
	updatesBtn := urlBtn(updatesText, config.SupportChannel)
	updatesBtn.IconCustomEmojiId = updatesIcon

	supportText, supportIcon := buttonText(chatID, "SUPPORT_BTN")
	supportBtn := urlBtn(supportText, config.SupportChat)
	supportBtn.IconCustomEmojiId = supportIcon

	sourceText, sourceIcon := buttonText(chatID, "SOURCE_BTN")
	sourceBtn := urlBtn(sourceText, "https://github.com/TheTeamVivek/YukkiMusic")
	sourceBtn.IconCustomEmojiId = sourceIcon

	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{addMeBtn},
			{helpBtn},
			{updatesBtn, supportBtn},
			{sourceBtn},
		},
	}
}

func GetHelpKeyboard(chatID int64) td.ReplyMarkup {
	adminsText, adminsIcon := buttonText(chatID, "HELP_ADMINS_BTN")
	adminsBtn := dataBtn(adminsText, "help:admins")
	adminsBtn.IconCustomEmojiId = adminsIcon

	publicText, publicIcon := buttonText(chatID, "HELP_PUBLIC_BTN")
	publicBtn := dataBtn(publicText, "help:public")
	publicBtn.IconCustomEmojiId = publicIcon

	ownerText, ownerIcon := buttonText(chatID, "HELP_OWNER_BTN")
	ownerBtn := dataBtn(ownerText, "help:owner")
	ownerBtn.IconCustomEmojiId = ownerIcon

	sudoersText, sudoersIcon := buttonText(chatID, "HELP_SUDOERS_BTN")
	sudoersBtn := dataBtn(sudoersText, "help:sudoers")
	sudoersBtn.IconCustomEmojiId = sudoersIcon

	backText, backIcon := buttonText(chatID, "BACK_BTN")
	backBtn := styleBtn(backText, "start", "")
	backBtn.IconCustomEmojiId = backIcon

	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{adminsBtn, publicBtn},
			{ownerBtn, sudoersBtn},
			{backBtn},
		},
	}
}

func GetBackKeyboard(chatID int64) td.ReplyMarkup {
	text, iconID := buttonText(chatID, "BACK_BTN")
	btn := styleBtn(text, "help:main", "blue")
	btn.IconCustomEmojiId = iconID
	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{{btn}},
	}
}

func GetRestartConfirmMarkup(chatID int64) td.ReplyMarkup {
	botText, botIcon := buttonText(chatID, "restart_btn_bot")
	replayText, replayIcon := buttonText(chatID, "restart_btn_replay")

	botBtn := styleBtn(botText, "restart:bot", "red")
	botBtn.IconCustomEmojiId = botIcon

	replayBtn := styleBtn(replayText, "restart:replay", "green")
	replayBtn.IconCustomEmojiId = replayIcon

	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{botBtn, replayBtn},
		},
	}
}
