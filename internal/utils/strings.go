/*
 * This file is part of YukkiMusic.
 *
 * YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
 * Copyright (C) 2025 TheTeamVivek
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 */
package utils

import (
	"fmt"
	"html"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func ShortTitle(title string, max ...int) string {
	limit := 25
	if len(max) > 0 {
		limit = max[0]
	}
	runes := []rune(title)
	if len(runes) <= limit {
		return title
	}
	return string(runes[:limit]) + "..."
}

func CleanURL(raw string) string {
	parts := strings.SplitN(raw, "?", 2)
	return parts[0]
}

func MentionHTML(u *telegram.UserObj) string {
	if u == nil {
		return "Unknown"
	}

	fullName := u.FirstName
	if u.LastName != "" {
		fullName += " " + u.LastName
	}

	if fullName == "" {
		fullName = "User"
	}
	fullName = html.EscapeString(ShortTitle(fullName, 15))

	return fmt.Sprintf(`<a href="tg://user?id=%d">%s</a>`, u.ID, fullName)
}
