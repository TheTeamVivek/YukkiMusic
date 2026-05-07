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

package platforms

import (
	"errors"
	"net/url"
	"strings"
	"unicode"
)

var errUnsafeURL = errors.New("invalid or unsafe url")

func sanitizeMediaURL(raw string) (string, error) {
	u := strings.TrimSpace(raw)
	if u == "" {
		return "", errUnsafeURL
	}

	for _, r := range u {
		if unicode.IsControl(r) {
			return "", errUnsafeURL
		}
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return "", errUnsafeURL
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errUnsafeURL
	}

	if parsed.Host == "" || parsed.User != nil {
		return "", errUnsafeURL
	}

	return parsed.String(), nil
}
