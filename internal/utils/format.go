/*
 * ● YukkiMusic
 * ○ A high-performance engine for streaming music in Telegram voicechats.
 *
 * Copyright (C) 2024 TheTeamVivek
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
	"strconv"
)

// FormatDuration returns a duration string (e.g., "1d 2h", "3h 4m", "5m 6s").
func FormatDuration(sec int) string {
	if sec < 0 {
		sec = 0
	}

	const (
		day  = 86400
		hour = 3600
		min  = 60
	)

	if sec < min {
		return strconv.Itoa(sec) + "s"
	}
	if sec < hour {
		return strconv.Itoa(sec/min) + "m " + strconv.Itoa(sec%min) + "s"
	}
	if sec < day {
		return strconv.Itoa(sec/hour) + "h " + strconv.Itoa((sec%hour)/min) + "m"
	}

	return strconv.Itoa(sec/day) + "d " + strconv.Itoa((sec%day)/hour) + "h"
}

// FormatTime returns a clock-style duration string (e.g., "HH:MM:SS" or "MM:SS").
func FormatTime(sec int) string {
	if sec < 0 {
		sec = 0
	}
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60

	if h > 0 {
		return strconv.Itoa(h) + ":" + padZero(m) + ":" + padZero(s)
	}
	return padZero(m) + ":" + padZero(s)
}

func padZero(v int) string {
	if v < 10 {
		return "0" + strconv.Itoa(v)
	}
	return strconv.Itoa(v)
}
