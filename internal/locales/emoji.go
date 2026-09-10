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
	"regexp"
	"strings"
)

// Emoji tokens that can be used inside any locale template as {emoji:<emoji>}.
// A token whose mapped value is empty expands to the bare emoji character;
// special tokens (premium/custom emoji) carry their Telegram entity id and the
// <tg-emoji> tag is assembled in expandEmojis. The tag is language independent,
// so tokens expand identically for every locale.
var emojiTokens = map[string]string{
	"⏩":    "",
	"⏭️":   "",
	"⏰":    "",
	"⏱":    "",
	"⏱️":   "",
	"⏳":    "",
	"⏸️":   "",
	"⏹️":   "5247149163132493357",
	"▶️":   "5384208892467106958",
	"♻️":   "",
	"⚙️":   "",
	"⚠️":   "",
	"⚡":    "5174818074167083884",
	"⚪":    "",
	"✅":    "",
	"✨":    "",
	"❌":    "",
	"❓":    "5436113877181941026",
	"➕":    "",
	"➡️":   "",
	"⬅️":   "5404737878364277847",
	"🌍":    "5388779214411412749",
	"🌐":    "",
	"🍃":    "",
	"🎙️":   "",
	"🎧":    "",
	"🎵":    "",
	"🎶":    "",
	"🏃‍♂️": "",
	"🏓":    "",
	"👋":    "",
	"👍":    "5368324170671202286",
	"👑":    "5319149831673887746",
	"👤":    "",
	"👥":    "",
	"💡":    "",
	"💫":    "",
	"💬":    "5443038326535759644",
	"💻":    "",
	"📊":    "",
	"📌":    "",
	"📍":    "",
	"📒":    "5226647579126678454",
	"📚":    "",
	"📛":    "",
	"📜":    "",
	"📝":    "",
	"📡":    "",
	"📢":    "5789428375261023681",
	"📥":    "",
	"📦":    "",
	"📭":    "",
	"🔀":    "",
	"🔁":    "",
	"🔇":    "",
	"🔊":    "5800920921766104807",
	"🔍":    "",
	"🔗":    "5280942781861208168",
	"🔧":    "",
	"🔴":    "",
	"🖼️":   "",
	"🗑️":   "",
	"😂":    "",
	"😉":    "",
	"😎":    "",
	"🚀":    "",
	"🚫":    "5240241223632954241",
	"🚦":    "6334531163314456604",
	"🛠":    "5398095118735521227",
	"🛠️":   "",
	"🛡️":   "",
	"🟢":    "",
	"🤖":    "",
	"🧩":    "",
	"🧹":    "",
}

var unknownEmojiRe = regexp.MustCompile(`\{emoji:([^}]+)\}`)

func expandEmojis(s string) string {
	for emojiChar, emojiID := range emojiTokens {
		token := "{emoji:" + emojiChar + "}"
		if emojiID != "" {
			emojiChar = `<tg-emoji emoji-id="` + emojiID + `">` + emojiChar + `</tg-emoji>`
		}
		s = strings.ReplaceAll(s, token, emojiChar)
	}
	s = unknownEmojiRe.ReplaceAllString(s, "$1")
	return s
}
