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

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
)

func init() {
	helpTexts["/privacy"] = `<i>Show the bot's privacy policy.</i>`
}

func privacyHandler(m *tg.NewMessage) error {
	privacyText := fmt.Sprintf(`<b>🛡 Privacy Policy &amp; Data Handling</b>

Your privacy is important to us. This bot is designed with privacy in mind.

<b>📊 Data We Collect</b>
<blockquote>We only store essential data required for the bot to function:
• <b>User & Chat IDs:</b> To identify groups and manage settings.
• <b>Preferences:</b> Language settings and bot configurations.
• <b>Access Control:</b> Authorized user lists for your group.
• <b>RTMP Config:</b> Only if you use the RTMP streaming feature.</blockquote>

<b>📩 Message Privacy</b>
<blockquote>• The bot <b>only</b> reads messages that start with a command (e.g., <code>/play</code>) or interactions with its own buttons.
• It <b>does not</b> read, store, or monitor your private conversations or general group messages.</blockquote>

<b>🌐 Third-Party Services</b>
<blockquote>• We use external services like <b>YouTube</b> and <b>Spotify</b> solely to search for and stream the music you request.
• No personal data is shared with these services beyond the search query itself.</blockquote>

<b>🤝 Data Sharing</b>
<blockquote>• We <b>never</b> sell, share, or trade your data with third parties.
• All data is used strictly to provide and improve the bot's music streaming features.</blockquote>

<b>✨ Our Commitment</b>
This bot is an <a href="https://github.com/TheTeamVivek/YukkiMusic">open-source project</a> dedicated to providing a high-quality streaming experience while respecting user privacy.

<i>If you have any questions, feel free to join our <a href="%s">Support Chat</a>.</i>`, config.SupportChat)

	_, err := m.Reply(privacyText)
	return err
}
