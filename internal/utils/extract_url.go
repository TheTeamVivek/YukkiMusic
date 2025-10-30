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

	"github.com/amarnathcjd/gogram/telegram"
)

func ExtractURLs(m *telegram.NewMessage) ([]string, error) {
	if m == nil || m.Message == nil {
		return nil, fmt.Errorf("invalid message")
	}
	capacity := len(m.Message.Entities)
	if m.IsReply() {
		if r, err := m.GetReplyMessage(); err == nil {
			capacity += len(r.Message.Entities)
		}
	}

	urls := make([]string, 0, capacity)

	collect := func(msg *telegram.MessageObj) {
		text := msg.Message
		for _, ent := range msg.Entities {
			switch e := ent.(type) {
			case *telegram.MessageEntityURL:
				if int(e.Offset+e.Length) <= len(text) {
					urls = append(urls, text[e.Offset:e.Offset+e.Length])
				}
			case *telegram.MessageEntityTextURL:
				if e.URL != "" {
					urls = append(urls, e.URL)
				}
			}
		}
	}

	collect(m.Message)

	if m.IsReply() {
		r, err := m.GetReplyMessage()
		if err != nil {
			if len(urls) > 0 {
				return urls, fmt.Errorf("failed to fetch reply message: %w", err)
			}
			return nil, fmt.Errorf("failed to fetch reply message: %w", err)
		}
		collect(r.Message)
	}

	if len(urls) == 0 {
		return nil, fmt.Errorf("no URLs found")
	}

	return urls, nil
}
