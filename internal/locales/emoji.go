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

// Emoji tokens that can be used inside any locale template as {emoji:<name>}.
// Premium emojis use Telegram's custom emoji tag: the <tg-emoji> entity renders
// the custom emoji when the sender has Telegram Premium, otherwise the emoji
// character inside the tag is shown as a fallback. The tag is language
// independent, so tokens expand identically for every locale.
var emojiTokens = map[string]string{
	"{emoji:thumbs_up}": `<tg-emoji emoji-id="5368324170671202286">👍</tg-emoji>`,
}

func expandEmojis(s string) string {
	for token, emoji := range emojiTokens {
		s = strings.ReplaceAll(s, token, emoji)
	}
	return s
}
