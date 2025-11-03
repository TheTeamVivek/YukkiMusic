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
package modules

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/utils"
)

func queueHandler(m *telegram.NewMessage) error {
	return handleQueue(m, false)
}

func cqueueHandler(m *telegram.NewMessage) error {
	return handleQueue(m, true)
}

func handleQueue(m *telegram.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.EndGroup
	}
	if !r.IsActiveChat() || r.Track == nil {
		m.Reply("⚠️ <b>No active playback.</b>\nNothing is queued right now.")
		return telegram.EndGroup
	}
	mystic, err := m.Reply("⏳ <b>Fetching queue...</b>")
	if err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("🎶 <b>Current Queue</b>\n\n")
	b.WriteString("▶️ <b>Now Playing:</b>\n")
	b.WriteString(fmt.Sprintf("🎧 <a href=\"%s\">%s</a> – %s [%s]\n\n",
		r.Track.URL, html.EscapeString(utils.ShortTitle(r.Track.Title, 35)),
		r.Track.BY,
		formatDuration(r.Track.Duration),
	))
	if len(r.Queue) > 0 {
		b.WriteString("⏳ <b>Up Next:</b>\n\n")
		for i, track := range r.Queue {
			if i >= 10 {
				remaining := len(r.Queue) - 10
				b.WriteString(fmt.Sprintf("\n… and %d more", remaining))
				break
			}
			b.WriteString(fmt.Sprintf("%d. 🎵 <a href=\"%s\">%s</a> – %s [%s]\n",
				i+1,
				track.URL,
				html.EscapeString(utils.ShortTitle(track.Title, 35)),
				track.BY,
				formatDuration(track.Duration),
			))
		}
	} else {
		b.WriteString("📭 <i>No more songs in queue.</i>")
	}
	utils.EOR(mystic, b.String())
	return telegram.EndGroup
}

func removeHandler(m *telegram.NewMessage) error {
	return handleRemove(m, false)
}

func cremoveHandler(m *telegram.NewMessage) error {
	return handleRemove(m, true)
}

func handleRemove(m *telegram.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.EndGroup
	}
	if r.Track == nil {
		m.Reply("⚠️ No active playback or queue found.")
		return telegram.EndGroup
	}
	if len(r.Queue) == 0 {
		m.Reply("⚠️ The queue is already empty.")
		return telegram.EndGroup
	}
	args := strings.Fields(m.Text())
	if len(args) < 2 {
		m.Reply(fmt.Sprintf("⚠️ Please provide the index of the track to remove.\nUsage: <code>%s [index]</code>", getCommand(m)))
		return telegram.EndGroup
	}
	index, err := strconv.Atoi(args[1])
	if err != nil {
		m.Reply("⚠️ Invalid index: must be a number.")
		return telegram.EndGroup
	}
	if index <= 0 {
		m.Reply("⚠️ Index must be greater than 0.")
		return telegram.EndGroup
	}
	queueLen := len(r.Queue)
	if index > queueLen {
		m.Reply(fmt.Sprintf("⚠️ Invalid index. Queue has only %d tracks.", queueLen))
		return telegram.EndGroup
	}
	r.RemoveFromQueue(index - 1)
	m.Reply(fmt.Sprintf("✅ Removed track at position %d from the queue by %s.", index, utils.MentionHTML(m.Sender)))
	return telegram.EndGroup
}

func clearHandler(m *telegram.NewMessage) error {
	return handleClear(m, false)
}

func cclearHandler(m *telegram.NewMessage) error {
	return handleClear(m, true)
}

func handleClear(m *telegram.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.EndGroup
	}
	if r.Track == nil {
		m.Reply("⚠️ There is no active playback or queue to clear.")
		return telegram.EndGroup
	}
	if len(r.Queue) == 0 {
		m.Reply("⚠️ The queue is already empty.")
		return telegram.EndGroup
	}
	r.RemoveFromQueue(-1)
	m.Reply(fmt.Sprintf("✅ The queue has been cleared by %s.", utils.MentionHTML(m.Sender)))
	return telegram.EndGroup
}

func moveHandler(m *telegram.NewMessage) error {
	return handleMove(m, false)
}

func cmoveHandler(m *telegram.NewMessage) error {
	return handleMove(m, true)
}

func handleMove(m *telegram.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.EndGroup
	}
	if r.Track == nil {
		m.Reply("⚠️ No active playback or queue found.")
		return telegram.EndGroup
	}
	if len(r.Queue) == 0 {
		m.Reply("⚠️ The queue is empty.")
		return telegram.EndGroup
	}
	args := strings.Fields(m.Text())
	if len(args) < 3 {
		m.Reply(fmt.Sprintf("⚠️ Usage: <code>%s [from] [to]</code>", getCommand(m)))
		return telegram.EndGroup
	}
	from, err1 := strconv.Atoi(args[1])
	to, err2 := strconv.Atoi(args[2])
	if err1 != nil || err2 != nil {
		m.Reply("⚠️ Invalid numbers. Example: <code>/move 2 1</code>")
		return telegram.EndGroup
	}
	if from <= 0 || to <= 0 {
		m.Reply("⚠️ Indexes must be greater than 0.")
		return telegram.EndGroup
	}
	queueLen := len(r.Queue)
	if from > queueLen || to > queueLen {
		m.Reply(fmt.Sprintf("⚠️ Queue has only %d tracks.", queueLen))
		return telegram.EndGroup
	}
	r.MoveInQueue(from-1, to-1)
	m.Reply(fmt.Sprintf("✅ Moved track from position %d to %d by %s.", from, to, utils.MentionHTML(m.Sender)))
	return telegram.EndGroup
}
