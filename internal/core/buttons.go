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
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

var F func(chatID int64, key string, values ...locales.Arg) string // overwritten from main.go

// TDBot is the gotdbot client, set from main.go.
var TDBot *td.Client

func botUsername() string {
	if TDBot == nil || TDBot.Me == nil || TDBot.Me.Usernames == nil {
		return ""
	}
	usernames := TDBot.Me.Usernames.ActiveUsernames
	if len(usernames) == 0 {
		return ""
	}
	return usernames[0]
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
	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				urlBtn(
					F(chatID, "ADD_ME_BTN"),
					"https://t.me/"+botUsername()+"?startgroup&admin=invite_users",
				),
			},
		},
	}
}

func GetCancelKeyboard(chatID int64) td.ReplyMarkup {
	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				dataBtn(F(chatID, "DOWNLOAD_CANCEL_BTN"), "cancel"),
			},
		},
	}
}

func GetBroadcastCancelKeyboard(chatID int64) td.ReplyMarkup {
	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				dataBtn(F(chatID, "BROADCAST_CANCEL_BTN"), "bcast_cancel"),
			},
		},
	}
}

func SuppMarkup(chatID int64) td.ReplyMarkup {
	btn := urlBtn(F(chatID, "SUPPORT_BTN"), config.SupportChat)
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

	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				styleBtn(F(chatID, text), cb, "green"),
				styleBtn(F(chatID, "CONFIRM_STOP_BTN"), prefix+"stop", "red"),
			},
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
	rows = append(rows, []td.InlineKeyboardButton{
		dataBtn("▷", prefix+"resume"),
		dataBtn("II", prefix+"pause"),
		dataBtn("‣‣I", prefix+"skip"),
		dataBtn("▢", prefix+"stop"),
	})

	rows = append(rows, []td.InlineKeyboardButton{
		dataBtn("↩ 15s", prefix+"seekback_15"),
		dataBtn("⟳", prefix+"replay"),
		dataBtn("15s ↪", prefix+"seek_15"),
	})

	rows = append(rows, []td.InlineKeyboardButton{
		dataBtn(F(chatID, "CLOSE_BTN"), "close"),
	})

	return &td.ReplyMarkupInlineKeyboard{Rows: rows}
}

func GetGroupHelpKeyboard(chatID int64) td.ReplyMarkup {
	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				urlBtn(
					F(chatID, "GC_HELP_BTN"),
					"https://t.me/"+botUsername()+"?start=pm_help",
				),
			},
		},
	}
}

func GetStartMarkup(chatID int64) td.ReplyMarkup {
	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				urlBtn(
					F(chatID, "ADD_ME_BTN"),
					"https://t.me/"+botUsername()+"?startgroup&admin=invite_users",
				),
			},
			{
				dataBtn(F(chatID, "START_HELP_BTN"), "help_cb"),
			},
			{
				urlBtn(F(chatID, "UPDATES_BTN"), config.SupportChannel),
				urlBtn(F(chatID, "SUPPORT_BTN"), config.SupportChat),
			},
		},
	}
}

func GetHelpKeyboard(chatID int64) td.ReplyMarkup {
	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				dataBtn(F(chatID, "HELP_ADMINS_BTN"), "help:admins"),
				dataBtn(F(chatID, "HELP_PUBLIC_BTN"), "help:public"),
			},
			{
				dataBtn(F(chatID, "HELP_OWNER_BTN"), "help:owner"),
				dataBtn(F(chatID, "HELP_SUDOERS_BTN"), "help:sudoers"),
			},
			{
				styleBtn(F(chatID, "BACK_BTN"), "start", ""),
			},
		},
	}
}

func GetBackKeyboard(chatID int64) td.ReplyMarkup {
	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				styleBtn(F(chatID, "BACK_BTN"), "help:main", "blue"),
			},
		},
	}
}

func GetRestartConfirmMarkup(chatID int64) td.ReplyMarkup {
	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				styleBtn(F(chatID, "restart_btn_bot"), "restart:bot", "red"),
				styleBtn(F(chatID, "restart_btn_replay"), "restart:replay", "green"),
			},
		},
	}
}
