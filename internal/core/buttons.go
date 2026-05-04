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

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/locales"
	"main/internal/utils"
)

var F func(chatID int64, key string, values ...locales.Arg) string // overwritten from main.go

func styleBtn(text, cb, colour string) tg.KeyboardButton {
	b := tg.Button.Data(text, cb)

	if config.DisableColour {
		return b
	}

	switch strings.ToLower(colour) {
	case "red":
		b.Danger()
	case "blue":
		b.Primary()
	case "green":
		b.Success()
	}
	return b
}

func AddMeMarkup(chatID int64) tg.ReplyMarkup {
	return tg.NewKeyboard().
		AddRow(
			tg.Button.URL(
				F(chatID, "ADD_ME_BTN"),
				"https://t.me/"+Bot.Me().Username+"?startgroup&admin=invite_users",
			),
		).
		Build()
}

func GetCancelKeyboard(chatID int64) *tg.ReplyInlineMarkup {
	return tg.NewKeyboard().
		AddRow(
			tg.Button.Data(F(chatID, "DOWNLOAD_CANCEL_BTN"), "cancel"),
		).
		Build()
}

func GetBroadcastCancelKeyboard(chatID int64) *tg.ReplyInlineMarkup {
	return tg.NewKeyboard().
		AddRow(
			tg.Button.Data(F(chatID, "BROADCAST_CANCEL_BTN"), "bcast_cancel"),
		).
		Build()
}

func SuppMarkup(chatID int64) tg.ReplyMarkup {
	return tg.NewKeyboard().
		AddRow(
			tg.Button.URL(F(chatID, "SUPPORT_BTN"), config.SupportChat),
		).
		Build()
}

func GetStopConfirmMarkup(
	chatID int64,
	r *RoomState,
	isPaused bool,
) tg.ReplyMarkup {
	btn := tg.NewKeyboard()
	prefix := fmt.Sprintf("room:%d:", r.ID)

	text, cb := "CONFIRM_UNMUTE_BTN", prefix+"unmute"

	if isPaused {
		text, cb = "CONFIRM_RESUME_BTN", prefix+"resume"
	}

	btn.AddRow(
		styleBtn(F(chatID, text), cb, "green"),
		styleBtn(F(chatID, "CONFIRM_STOP_BTN"), prefix+"stop", "red"),
	)

	return btn.Build()
}

func GetPlayMarkup(chatID int64, r *RoomState, queued bool) tg.ReplyMarkup {
	btn := tg.NewKeyboard()
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

	if !queued {
		btn.AddRow(
			tg.Button.Data(progress, "progress"),
		)
	}
	btn.AddRow(
		tg.Button.Data("▷", prefix+"resume"),
		tg.Button.Data("II", prefix+"pause"),
		tg.Button.Data("‣‣I", prefix+"skip"),
		tg.Button.Data("▢", prefix+"stop"),
	)

	btn.AddRow(
		tg.Button.Data("↩ 15s", prefix+"seekback_15"),
		tg.Button.Data("⟳", prefix+"replay"),
		tg.Button.Data("15s ↪", prefix+"seek_15"),
	)

	btn.AddRow(
		tg.Button.Data(F(chatID, "CLOSE_BTN"), "close"),
	)

	return btn.Build()
}

func GetGroupHelpKeyboard(chatID int64) *tg.ReplyInlineMarkup {
	return tg.NewKeyboard().
		AddRow(
			tg.Button.URL(F(chatID, "GC_HELP_BTN"), "https://t.me/"+Bot.Me().Username+"?start=pm_help"),
		).
		Build()
}

func GetStartMarkup(chatID int64) tg.ReplyMarkup {
	return tg.NewKeyboard().
		AddRow(
			tg.Button.URL(
				F(chatID, "ADD_ME_BTN"),
				"https://t.me/"+Bot.Me().Username+"?startgroup&admin=invite_users",
			),
		).
		AddRow(
			tg.Button.Data(
				F(chatID, "START_HELP_BTN"),
				"help_cb",
			),
		).
		AddRow(
			tg.Button.URL(
				F(chatID, "UPDATES_BTN"),
				config.SupportChannel,
			),
			tg.Button.URL(
				F(chatID, "SUPPORT_BTN"),
				config.SupportChat,
			),
		).
		Build()
}

func GetHelpKeyboard(chatID int64) *tg.ReplyInlineMarkup {
	return tg.NewKeyboard().
		AddRow(
			tg.Button.Data(
				F(chatID, "HELP_ADMINS_BTN"),
				"help:admins",
			),
			tg.Button.Data(
				F(chatID, "HELP_PUBLIC_BTN"),
				"help:public",
			),
		).
		AddRow(
			tg.Button.Data(
				F(chatID, "HELP_OWNER_BTN"),
				"help:owner",
			),
			tg.Button.Data(
				F(chatID, "HELP_SUDOERS_BTN"),
				"help:sudoers",
			),
		).
		AddRow(
			styleBtn(F(chatID, "BACK_BTN"), "start", ""),
		).
		Build()
}

func GetBackKeyboard(chatID int64) *tg.ReplyInlineMarkup {
	return tg.NewKeyboard().
		AddRow(
			styleBtn(F(chatID, "BACK_BTN"), "help:main", "blue"),
		).
		Build()
}

func GetRestartConfirmMarkup(chatID int64) *tg.ReplyInlineMarkup {
	return tg.NewKeyboard().
		AddRow(
			styleBtn(F(chatID, "restart_btn_bot"), "restart:bot", "red"),
			styleBtn(F(chatID, "restart_btn_replay"), "restart:replay", "green"),
		).
		Build()
}
