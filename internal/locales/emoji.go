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

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"yukkimusic/config"
)

// Emoji tokens are written in locale templates as {emoji:X}, where X is one of:
//
//	▶️         an emoji listed in emojiTokens (premium id if it has one)
//	5234…|▶️   a custom emoji id with a fallback character
//	5234…      a custom emoji id with no fallback
//
// In messages a token becomes <tg-emoji> with the fallback inside, and Telegram
// shows the fallback when the bot cannot send premium emoji. In buttons only one
// custom emoji can be shown as the icon, see buttonEmoji.
var emojiTokens = map[string]int64{
	"⏹️": 5247149163132493357,
	"▶️": 5384208892467106958,
	"⚙️": 5398095118735521227,
	"⚠️": 5420323339723881652,
	"⚡":  5174818074167083884,
	"❓":  5436113877181941026,
	"⬅️": 5404737878364277847,
	"🌍": 5388779214411412749,
	"🎵": 5463107823946717464,
	"👍": 5368324170671202286,
	"👑": 5319149831673887746,
	"💬": 5443038326535759644,
	"📒": 5226647579126678454,
	"📜": 6323096332579899122,
	"📢": 5789428375261023681,
	"🔊": 5800920921766104807,
	"🔗": 5280942781861208168,
	"🚫": 5240241223632954241,
	"🚦": 6334531163314456604,
	"🛠": 5398095118735521227,
}

var emojiRe = regexp.MustCompile(`\{emoji:[^}]+\}`)

func parseToken(tok string) (int64, string) {
	c := tok[len("{emoji:") : len(tok)-1]
	if idStr, char, ok := strings.Cut(c, "|"); ok {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		return id, char
	}
	if id, err := strconv.ParseInt(c, 10, 64); err == nil {
		return id, ""
	}
	return emojiTokens[c], c
}

func plainEmoji(tok string) string {
	_, char := parseToken(tok)
	return char
}

func expandEmojis(s string) string {
	return emojiRe.ReplaceAllStringFunc(s, func(tok string) string {
		id, char := parseToken(tok)
		if id == 0 {
			return char
		}
		return fmt.Sprintf(`<tg-emoji emoji-id="%d">%s</tg-emoji>`, id, char)
	})
}

// buttonEmoji returns the button text and its icon id. Buttons have no
// fallback and take a single custom emoji, so with premium the leftmost token
// that has an id becomes the icon and every token is removed from the text.
// Without premium, or without any id, tokens become plain emoji.
func buttonEmoji(s string) (string, int64) {
	var icon int64
	if config.HasPremium {
		for _, tok := range emojiRe.FindAllString(s, -1) {
			if icon, _ = parseToken(tok); icon != 0 {
				break
			}
		}
	}
	if icon != 0 {
		return strings.TrimSpace(emojiRe.ReplaceAllString(s, "")), icon
	}
	return strings.TrimSpace(emojiRe.ReplaceAllStringFunc(s, plainEmoji)), 0
}
