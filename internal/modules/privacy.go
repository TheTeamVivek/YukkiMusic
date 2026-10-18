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

package modules

import (
	"fmt"
	"html"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config"
)

func init() {
	helpTexts["/privacy"] = `<i>Show the bot's privacy policy and data handling.</i>`
}

// privacySupport resolves the best support link to show in the policy.
func privacySupport() string {
	s := config.SupportChat
	if s == "" {
		s = config.SupportChannel
	}
	if s == "" {
		s = "https://t.me/TheTeamVk"
	}
	return html.EscapeString(s)
}

func privacyHandler(c *td.Client, m *td.Message) error {
	const repo = "https://github.com/TheTeamVivek/YukkiMusic"

	text := fmt.Sprintf(`<h3>🛡 Privacy Policy &amp; Data Handling</h3>
<p>YukkiMusic is built with privacy in mind. Below is exactly what it stores, why it is stored, and how you stay in control.</p>

<h4>📦 Data We Store</h4>
<p>Only the information required to run the bot — nothing more:</p>
<table bordered striped compact>
<tr><th>Stored</th><th>Purpose</th></tr>
<tr><td><b>User &amp; chat IDs</b></td><td>Identify groups and users so your settings apply</td></tr>
<tr><td><b>Preferences</b></td><td>Language, play mode, clean mode, admin mode, thumbnails</td></tr>
<tr><td><b>Access lists</b></td><td>Authorized users, sudo users, and the blacklist</td></tr>
</table>

<h4>💬 Messages</h4>
<aside>The bot reads <b>only</b> messages that begin with one of its commands (for example <code>/play</code>) or button presses on its own messages. It <b>never</b> reads, stores, or monitors your private chats or ordinary group conversations.</aside>

<h4>🌐 Third-Party Services</h4>
<p>To find and stream the music you request, the bot sends the search query to services such as <b>YouTube</b> and <b>Spotify</b>. Only the query itself is sent — <b>no personal data</b> is shared with these services.</p>

<h4>🔒 Data Sharing</h4>
<p>Your data is <b>never</b> sold, shared, or traded with third parties. It is used strictly to operate and improve the bot's music streaming features.</p>

<h4>🗑 Data Removal</h4>
<p>Access can be cleaned up at any time: the owner can manage authorized users, sudo users, and the blacklist. You may also contact support to request removal of stored data.</p>

<footer>YukkiMusic is an <a href="%s">open-source project</a>. If you have any questions, feel free to join our <a href="%s">Support Chat</a>.</footer>`, repo, privacySupport())

	_, err := m.ReplyRichMessage(c, &td.InputRichMessage{
		Source: &td.RichMessageSourceHtml{Text: text},
	}, nil)
	return err
}
