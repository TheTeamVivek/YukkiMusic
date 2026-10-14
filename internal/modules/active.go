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
	"slices"
	"strings"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/core"
	"yukkimusic/internal/utils"
)

func init() {
	helpTexts["/active"] = `<i>Show all active voice chat sessions.</i>

<u>Usage:</u>
<b>/active</b> or <b>/ac</b> — List active chats

<b>📊 Information Shown:</b>
• Total active voice chats
• Playback status (playing/paused/muted)
• Now playing track and queue size

<b>🔒 Restrictions:</b>
• <b>Sudo users</b> only

<b>💡 Use Case:</b>
Monitor bot usage and identify issues.`

	keys := []string{"/ac", "/activevc", "/activevoice"}
	for _, k := range keys {
		helpTexts[k] = helpTexts["/active"]
	}
}

func activeHandler(c *td.Client, m *td.Message) error {
	if !checkSudo(c, m) {
		return nil
	}
	chatID := m.ChatID()

	// Only truly active sessions are reported: skip destroyed rooms and
	// rooms that are not actively playing anything.
	states := make(map[int64]core.State)
	for id, r := range core.Rooms.All() {
		if r == nil {
			continue
		}
		snap := r.Snapshot()
		if !snap.Active() {
			continue
		}
		states[id] = snap
	}

	if len(states) == 0 {
		_, err := m.ReplyText(c, "<b>🎵 Active Voice Chats</b>\n\nNo active chat sessions found.", &td.SendMessageOpts{
			DisableWebPagePreview: true,
		})
		return err
	}

	ids := make([]int64, 0, len(states))
	for id := range states {
		ids = append(ids, id)
	}
	slices.Sort(ids)

	var sb strings.Builder
	sb.WriteString("<h3>🎵 Active Voice Chats</h3>")
	sb.WriteString("\n\nThere are currently <b>")
	fmt.Fprintf(&sb, "%d", len(ids))
	sb.WriteString("</b> active voice/video chat(s) running.")

	sb.WriteString("\n\n<details>\n<summary>📊 Click to Show Active Chats</summary>\n")
	sb.WriteString("<table bordered striped>\n")
	sb.WriteString("<thead>\n<tr><th>#</th><th>Chat ID</th><th>Status</th><th>Now Playing Track Info</th><th>Queue</th></tr>\n</thead>\n")
	sb.WriteString("<tbody>\n")

	for i, id := range ids {
		snap := states[id]

		sb.WriteString("<tr>")
		fmt.Fprintf(&sb, "<td><b>%d</b></td>", i+1)
		fmt.Fprintf(&sb, "<td><code>%d</code></td>", id)
		fmt.Fprintf(&sb, "<td>%s</td>", activeStatusCell(snap))
		fmt.Fprintf(&sb, "<td>%s</td>", activeTrackCell(snap))
		fmt.Fprintf(&sb, "<td>%s</td>", activeQueueCell(snap))
		sb.WriteString("</tr>\n")
	}

	sb.WriteString("</tbody>\n</table>\n")
	fmt.Fprintf(&sb, "Total active: %d", len(ids))
	sb.WriteString("</details>")

	rich := &td.InputRichMessage{
		Source: td.RichMessageSourceHtml{
			Text: sb.String(),
		},
	}
	_, err := c.SendRichMessage(chatID, rich, &td.SendMessageOpts{
		DisableWebPagePreview: true,
	})
	return err
}

func activeStatusCell(s core.State) string {
	if s.Muted {
		return "<b>🔇 Muted</b>"
	}
	if s.Paused {
		return "<b>⏸ Paused</b>"
	}
	if t := s.Track; t != nil {
		if t.Video {
			return "<b>▶️ Playing (Video)</b>"
		}
		return "<b>▶️ Playing (Audio)</b>"
	}
	return "<b>⏹ Idle</b>"
}

func activeTrackCell(s core.State) string {
	t := s.Track
	if t == nil {
		return "<i>🔇 No song playing.</i>"
	}
	title := utils.EscapeHTML(utils.ShortTitle(t.Title, 25))
	if t.URL != "" {
		return fmt.Sprintf(
			"<a href=\"%s\">%s</a> <i>(%s)</i>",
			t.URL,
			title,
			utils.FormatDuration(t.Duration),
		)
	}
	return fmt.Sprintf("%s <i>(%s)</i>", title, utils.FormatDuration(t.Duration))
}

func activeQueueCell(s core.State) string {
	n := s.QueueLen
	if n == 0 {
		return "<i>Empty</i>"
	}
	return fmt.Sprintf("<b>%d</b>", n)
}
