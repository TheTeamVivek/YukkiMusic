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

package locales

import "strings"

// Emoji tokens that can be used inside any locale template as {emoji:<emoji>}.
// A token whose mapped value is empty expands to the bare emoji character;
// special tokens (premium/custom emoji) carry their Telegram entity id and the
// <tg-emoji> tag is assembled in expandEmojis. The tag is language independent,
// so tokens expand identically for every locale.
var emojiTokens = map[string]string{
	"{emoji:⏩}":    "",
	"{emoji:⏭️}":   "",
	"{emoji:⏰}":    "",
	"{emoji:⏱}":    "",
	"{emoji:⏱️}":   "",
	"{emoji:⏳}":    "",
	"{emoji:⏸️}":   "",
	"{emoji:⏹️}":   "",
	"{emoji:▶️}":   "",
	"{emoji:♻️}":   "",
	"{emoji:⚙️}":   "",
	"{emoji:⚠️}":   "",
	"{emoji:⚡}":    "",
	"{emoji:⚪}":    "",
	"{emoji:✅}":    "",
	"{emoji:✨}":    "",
	"{emoji:❌}":    "",
	"{emoji:➕}":    "",
	"{emoji:➡️}":   "",
	"{emoji:🌍}":    "",
	"{emoji:🌐}":    "",
	"{emoji:🍃}":    "",
	"{emoji:🎙️}":   "",
	"{emoji:🎧}":    "",
	"{emoji:🎵}":    "",
	"{emoji:🎶}":    "",
	"{emoji:🏃‍♂️}": "",
	"{emoji:🏓}":    "",
	"{emoji:👋}":    "",
	"{emoji:👍}":    "5368324170671202286",
	"{emoji:👑}":    "",
	"{emoji:👤}":    "",
	"{emoji:👥}":    "",
	"{emoji:💡}":    "",
	"{emoji:💫}":    "",
	"{emoji:💬}":    "",
	"{emoji:💻}":    "",
	"{emoji:📊}":    "",
	"{emoji:📌}":    "",
	"{emoji:📍}":    "",
	"{emoji:📚}":    "",
	"{emoji:📛}":    "",
	"{emoji:📜}":    "",
	"{emoji:📝}":    "",
	"{emoji:📡}":    "",
	"{emoji:📥}":    "",
	"{emoji:📦}":    "",
	"{emoji:📭}":    "",
	"{emoji:🔀}":    "",
	"{emoji:🔁}":    "",
	"{emoji:🔇}":    "",
	"{emoji:🔊}":    "",
	"{emoji:🔍}":    "",
	"{emoji:🔗}":    "",
	"{emoji:🔧}":    "",
	"{emoji:🔴}":    "",
	"{emoji:🖼️}":   "",
	"{emoji:🗑️}":   "",
	"{emoji:😂}":    "",
	"{emoji:😉}":    "",
	"{emoji:😎}":    "",
	"{emoji:🚀}":    "",
	"{emoji:🚫}":    "",
	"{emoji:🛠}":    "",
	"{emoji:🛠️}":   "",
	"{emoji:🛡️}":   "",
	"{emoji:🟢}":    "",
	"{emoji:🤖}":    "",
	"{emoji:🧩}":    "",
	"{emoji:🧹}":    "",
}

func expandEmojis(s string) string {
	for token, emojiID := range emojiTokens {
		emoji := strings.TrimSuffix(strings.TrimPrefix(token, "{emoji:"), "}")
		if emojiID != "" {
			emoji = `<tg-emoji emoji-id="` + emojiID + `">` + emoji + `</tg-emoji>`
		}
		s = strings.ReplaceAll(s, token, emoji)
	}
	return s
}
