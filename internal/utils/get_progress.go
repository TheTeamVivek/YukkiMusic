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

package utils

import (
	"fmt"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func GetProgress(statusMsg *telegram.NewMessage) *telegram.ProgressManager {
	pm := telegram.NewProgressManager(2)

	if statusMsg == nil {
		return pm
	}

	var opts *telegram.SendOptions
	if replyMarkup := statusMsg.ReplyMarkup(); replyMarkup != nil {
		opts = &telegram.SendOptions{ReplyMarkup: *replyMarkup}
	}

	pm.WithCallback(func(pi *telegram.ProgressInfo) {
		text := fmt.Sprintf(
			"<b>📥 Downloading your track...</b>\n"+
				"<pre>"+
				"Progress : %6.2f%%\n"+
				"Speed    : %s\n"+
				"Eta      : %s\n"+
				"Elapsed  : %s"+
				"</pre>",
			pi.Percentage,
			pi.SpeedString(),
			pi.ETAString(),
			pi.ElapsedString(),
		)
		statusMsg.Edit(text, opts)
	})

	return pm
}

/*
func GetProgressBar(playedSec, durationSec int) string {
	if durationSec == 0 || playedSec <= 0 {
		return "◉—————————"
	}

	percentage := (float64(playedSec) / float64(durationSec)) * 100
	umm := math.Floor(percentage)

	var bar string

	switch {
	case umm >= 0 && umm <= 10:
		bar = "◉—————————"
	case umm > 10 && umm < 20:
		bar = "—◉————————"
	case umm >= 20 && umm < 30:
		bar = "——◉———————"
	case umm >= 30 && umm < 40:
		bar = "———◉——————"
	case umm >= 40 && umm < 50:
		bar = "————◉—————"
	case umm >= 50 && umm < 60:
		bar = "—————◉————"
	case umm >= 60 && umm < 70:
		bar = "——————◉———"
	case umm >= 70 && umm < 80:
		bar = "———————◉——"
	case umm >= 80 && umm < 90:
		bar = "————————◉—"
	case umm >= 90 && umm <= 100:
		bar = "—————————◉"
	default:
		bar = "—————————◉"
	}
	return bar
}
*/

func GetProgressBar(playedSec, durationSec int) string {
	if durationSec <= 0 || playedSec <= 0 {
		return "◉—————————"
	}

	if playedSec >= durationSec {
		return "—————————◉"
	}

	index := (playedSec * 10) / durationSec
	if index > 9 {
		index = 9
	}

	return strings.Repeat("—", index) + "◉" + strings.Repeat("—", 9-index)
}
