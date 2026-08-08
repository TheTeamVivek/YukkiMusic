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

var progressBarCache = [10]string{
	"◉—————————",
	"—◉————————",
	"——◉———————",
	"———◉——————",
	"————◉—————",
	"—————◉————",
	"——————◉———",
	"———————◉——",
	"————————◉—",
	"—————————◉",
}

func GetProgressBar(playedSec, durationSec int) string {
	if durationSec <= 0 || playedSec <= 0 {
		return progressBarCache[0]
	}

	if playedSec >= durationSec {
		return progressBarCache[9]
	}

	index := min((playedSec*10)/durationSec, 9)

	return progressBarCache[index]
}
