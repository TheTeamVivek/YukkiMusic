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
	"html"
	"strconv"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func ShortTitle(title string, max ...int) string {
	title = html.UnescapeString(title)
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

// EscapeHTML escapes characters for Telegram HTML mode (only &, <, >).
func EscapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func CleanURL(raw string) string {
	before, _, _ := strings.Cut(raw, "?")
	return before
}

func MentionHTML(u *tg.UserObj) string {
	if u == nil {
		return "Unknown"
	}

	fullName := strings.TrimSpace(u.FirstName + " " + u.LastName)
	if fullName == "" {
		fullName = "User"
	}
	fullName = EscapeHTML(ShortTitle(fullName, 15))

	return fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a>", u.ID, fullName)
}

// IfElse returns `a` if condition is true, else returns `b`.
func IfElse[T any](condition bool, a, b T) T {
	if condition {
		return a
	}
	return b
}

// ParseBool converts strings like "on", "off", "enable", "disable", "true", "false"
// into a boolean value. Returns an error if input is invalid.
func ParseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "on", "enable", "enabled", "true", "1", "yes", "y":
		return true, nil
	case "off", "disable", "disabled", "false", "0", "no", "n":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean string: %q", s)
	}
}

// IntToStr converts any signed integer type to string.
func IntToStr[T ~int | ~int8 | ~int16 | ~int32 | ~int64](v T) string {
	return strconv.FormatInt(int64(v), 10)
}
