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
        "strconv"
	"strings"
)

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
// Returns empty string if type is unsupported.
func IntToStr(v any) string {
        switch n := v.(type) {
        case int:
                return strconv.Itoa(n)
        case int8, int16, int32, int64:
                return strconv.FormatInt(toInt64(n), 10)
        default:
                return ""
        }
}

func toInt64(v any) int64 {
        switch n := v.(type) {
        case int8:
                return int64(n)
        case int16:
                return int64(n)
        case int32:
                return int64(n)
        case int64:
                return n
        default:
                return 0
        }
}