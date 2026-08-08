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
	"unicode/utf16"

	td "github.com/AshokShau/gotdbot"
)

func ExtractURLs(c *td.Client, m *td.Message) ([]string, error) {
	if m == nil {
		return nil, fmt.Errorf("invalid message")
	}

	urls := make([]string, 0, len(m.Entities()))
	urls = append(urls, collectURLs(m.Text(), m.Entities())...)

	if m.ReplyToMessageID() <= 0 {
		return finalizeURLs(urls)
	}

	r, err := m.GetRepliedMessage(c)
	if err != nil {
		if len(urls) > 0 {
			return urls, fmt.Errorf("failed to fetch reply message: %w", err)
		}
		return nil, fmt.Errorf("failed to fetch reply message: %w", err)
	}

	urls = append(urls, collectURLs(r.Text(), r.Entities())...)
	return finalizeURLs(urls)
}

// --- Sub Functions ---

func collectURLs(text string, entities []td.TextEntity) []string {
	urls := make([]string, 0, len(entities))

	for _, ent := range entities {
		switch e := ent.Type.(type) {
		case *td.TextEntityTypeUrl:
			if seg, ok := sliceUTF16(text, ent.Offset, ent.Length); ok {
				urls = append(urls, seg)
			}
		case *td.TextEntityTypeTextUrl:
			if e.Url != "" {
				urls = append(urls, e.Url)
			}
		}
	}
	return urls
}

// sliceUTF16 slices the given string using UTF-16 code unit offsets,
// which is what Telegram text entities use.
func sliceUTF16(s string, offset, length int32) (string, bool) {
	if offset < 0 || length < 0 {
		return "", false
	}
	u16 := utf16.Encode([]rune(s))
	start := int(offset)
	end := start + int(length)
	if end > len(u16) {
		return "", false
	}
	return string(utf16.Decode(u16[start:end])), true
}

func finalizeURLs(urls []string) ([]string, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("no URLs found")
	}
	return urls, nil
}
