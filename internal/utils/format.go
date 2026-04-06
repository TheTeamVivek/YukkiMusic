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
	"fmt"
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
		return fmt.Sprintf("%ds", sec)
	}
	if sec < hour {
		return fmt.Sprintf("%dm %ds", sec/min, sec%min)
	}
	if sec < day {
		return fmt.Sprintf("%dh %dm", sec/hour, (sec%hour)/min)
	}

	return fmt.Sprintf(
		"%dd %dh",
		sec/day,
		(sec%day)/hour,
	)
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
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}
