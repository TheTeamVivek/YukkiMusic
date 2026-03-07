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
package utils

import (
	"fmt"
"strings"
	"math"

	"github.com/amarnathcjd/gogram/telegram"
)

func GetProgress(mystic *telegram.NewMessage) *telegram.ProgressManager {
	pm := telegram.NewProgressManager(2)

	if mystic == nil {
		return pm
	}

	var opts *telegram.SendOptions
	if replyMarkup := mystic.ReplyMarkup(); replyMarkup != nil {
		opts = &telegram.SendOptions{ReplyMarkup: *replyMarkup}
	}

	pm.WithCallback(func(pi *telegram.ProgressInfo) {
		mystic.Edit(buildProgressUI(pi), opts)
	})

	return pm
}

func buildProgressUI(pi *telegram.ProgressInfo) string {
	filled := int(pi.Percentage / 100 * 12)
	if filled > 12 {
		filled = 12
	}
	bar := strings.Repeat("▰", filled) + strings.Repeat("▱", 12-filled)

	var phase string
	switch {
	case pi.Percentage == 0:
		phase = "🔗 Connecting..."
	case pi.Percentage < 10:
		phase = "🚀 Starting"
	case pi.Percentage < 50:
		phase = "📥 Downloading"
	case pi.Percentage < 90:
		phase = "⚙️ Transferring"
	case pi.Percentage < 100:
		phase = "🏁 Almost done"
	default:
		phase = "✅ Complete"
	}

	return fmt.Sprintf(
		"%s <b>%s</b>\n\n"+
			"<code>%s</code>  <b>%.1f%%</b>\n\n"+
			"📦 <b>Size:</b>  <code>%s / %s</code>\n"+
			"⚡ <b>Speed:</b>  <code>%s</code>  │  <i>avg</i> <code>%s</code>\n"+
			"⏱ <b>ETA:</b>  <code>%s</code>  │  <b>elapsed</b> <code>%s</code>",
		phase,
		ShortTitle(pi.FileName, 32),
		bar,
		pi.Percentage,
		humanBytes(pi.Current),
		humanBytes(pi.TotalSize),
		pi.SpeedString(),
		pi.AvgSpeedString(),
		pi.ETAString(),
		pi.ElapsedString(),
	)
}

func humanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %s", float64(b)/float64(div), []string{"KB", "MB", "GB", "TB"}[exp])
}

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
