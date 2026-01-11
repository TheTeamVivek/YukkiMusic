/*
  - This file is part of YukkiMusic.
    *

  - YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
  - Copyright (C) 2025 TheTeamVivek
    *
  - This program is free software: you can redistribute it and/or modify
  - it under the terms of the GNU General Public License as published by
  - the Free Software Foundation, either version 3 of the License, or
  - (at your option) any later version.
    *
  - This program is distributed in the hope that it will be useful,
  - but WITHOUT ANY WARRANTY; without even the implied warranty of
  - MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
  - GNU General Public License for more details.
    *
  - You should have received a copy of the GNU General Public License
  - along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package modules

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/queue"] = `<i>Display the current playback queue.</i>

<u>Usage:</u>
<b>/queue</b> — Show queue

<b>📋 Display Format:</b>
• Now Playing - Current track with position
• Up Next - Next 10 tracks in queue
• Track info: Title, requester, duration

<b>⚙️ Features:</b>
• Real-time queue status
• Requester attribution
• Duration display
• Queue length indicator

<b>💡 Related Commands:</b>
• <code>/position</code> - Current track position only
• <code>/remove</code> - Remove specific track
• <code>/clear</code> - Clear all tracks
• <code>/move</code> - Reorder tracks`

	helpTexts["/remove"] = `<i>Remove a specific track from the queue.</i>

<u>Usage:</u>
<b>/remove [index]</b> — Remove track at position

<b>⚙️ Behavior:</b>
• Index starts from 1 (first track in queue)
• Cannot remove currently playing track
• Queue positions update automatically

<b>🔒 Restrictions:</b>
• Only <b>chat admins</b> or <b>authorized users</b> can use this

<b>💡 Examples:</b>
<code>/remove 1</code> — Remove first track in queue
<code>/remove 5</code> — Remove 5th track

<b>⚠️ Notes:</b>
• Use <code>/queue</code> to see track indices
• Invalid index shows error with queue length
• Use <code>/clear</code> to remove all tracks`

	helpTexts["/clear"] = `<i>Clear all tracks from the queue.</i>

<u>Usage:</u>
<b>/clear</b> — Remove all queued tracks

<b>⚙️ Behavior:</b>
• Removes all tracks from queue
• Current playing track continues
• Queue becomes empty after current track ends

<b>🔒 Restrictions:</b>
• Only <b>chat admins</b> or <b>authorized users</b> can use this

<b>⚠️ Warning:</b>
This action cannot be undone. Use <code>/remove</code> for selective removal.`

	helpTexts["/move"] = `<i>Reorder tracks in the queue.</i>

<u>Usage:</u>
<b>/move [from] [to]</b> — Move track from position to position

<b>⚙️ Behavior:</b>
• Moves track at index 'from' to index 'to'
• Other tracks shift positions accordingly
• Indices start from 1

<b>🔒 Restrictions:</b>
• Only <b>chat admins</b> or <b>authorized users</b> can use this

<b>💡 Examples:</b>
<code>/move 3 1</code> — Move 3rd track to 1st position
<code>/move 1 5</code> — Move 1st track to 5th position

<b>⚠️ Notes:</b>
• Both positions must be valid queue indices
• Use <code>/queue</code> to see current order
• Cannot move currently playing track`
}

func queueHandler(m *tg.NewMessage) error {
	return handleQueue(m, false)
}

func cqueueHandler(m *tg.NewMessage) error {
	return handleQueue(m, true)
}

func removeHandler(m *tg.NewMessage) error {
	return handleRemove(m, false)
}

func cremoveHandler(m *tg.NewMessage) error {
	return handleRemove(m, true)
}

func moveHandler(m *tg.NewMessage) error {
	return handleMove(m, false)
}

func cmoveHandler(m *tg.NewMessage) error {
	return handleMove(m, true)
}

func clearHandler(m *tg.NewMessage) error {
	return handleClear(m, false)
}

func cclearHandler(m *tg.NewMessage) error {
	return handleClear(m, true)
}

func handleQueue(m *tg.NewMessage, cplay bool) error {
	chatID := m.ChannelID()

	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return tg.ErrEndGroup
	}
	t := r.Track()

	if !r.IsActiveChat() || t == nil {
		m.Reply(F(chatID, "queue_no_active"))
		return tg.ErrEndGroup
	}

	var b strings.Builder

	b.WriteString(F(chatID, "queue_header"))
	b.WriteString("\n\n")

	// Now Playing
	b.WriteString(F(chatID, "queue_now_playing"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(
		"🎧 <a href=\"%s\">%s</a> — %s [%s]\n\n",
		t.URL,
		html.EscapeString(utils.ShortTitle(t.Title, 35)),
		t.Requester,
		formatDuration(t.Duration),
	))

	// Up Next
	if len(r.Queue()) > 0 {
		b.WriteString(F(chatID, "queue_up_next"))
		b.WriteString("\n\n")

		for i, track := range r.Queue() {
			if i >= 10 {
				b.WriteString(F(chatID, "queue_more_line", locales.Arg{
					"remaining": len(r.Queue()) - 10,
				}))
				break
			}

			b.WriteString(fmt.Sprintf(
				"%d. 🎵 <a href=\"%s\">%s</a> — %s [%s]\n",
				i+1,
				track.URL,
				html.EscapeString(utils.ShortTitle(track.Title, 35)),
				track.Requester,
				formatDuration(track.Duration),
			))
		}
	} else {
		b.WriteString(F(chatID, "queue_empty_tail"))
	}

	m.Reply(b.String())
	return tg.ErrEndGroup
}

func handleRemove(m *tg.NewMessage, cplay bool) error {
	chatID := m.ChannelID()

	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return tg.ErrEndGroup
	}
	t := r.Track()
	if !r.IsActiveChat() || t == nil {
		m.Reply(F(chatID, "queue_no_active"))
		return tg.ErrEndGroup
	}

	if len(r.Queue()) == 0 {
		m.Reply(F(chatID, "queue_empty"))
		return tg.ErrEndGroup
	}

	args := strings.Fields(m.Text())
	if len(args) < 2 {
		m.Reply(F(chatID, "remove_usage", locales.Arg{
			"cmd": getCommand(m),
		}))
		return tg.ErrEndGroup
	}

	index, err := strconv.Atoi(args[1])
	if err != nil {
		m.Reply(F(chatID, "remove_invalid_index"))
		return tg.ErrEndGroup
	}

	if index <= 0 {
		m.Reply(F(chatID, "remove_index_too_small"))
		return tg.ErrEndGroup
	}

	total := len(r.Queue())
	if index > total {
		m.Reply(F(chatID, "remove_index_too_big", locales.Arg{
			"total": total,
		}))
		return tg.ErrEndGroup
	}

	r.RemoveFromQueue(index - 1)

	m.Reply(F(chatID, "remove_success", locales.Arg{
		"index": index,
		"user":  utils.MentionHTML(m.Sender),
	}))

	return tg.ErrEndGroup
}

func handleClear(m *tg.NewMessage, cplay bool) error {
	chatID := m.ChannelID()

	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return tg.ErrEndGroup
	}
	t := r.Track()
	if !r.IsActiveChat() || t == nil {
		m.Reply(F(chatID, "clear_no_active"))
		return tg.ErrEndGroup
	}

	if len(r.Queue()) == 0 {
		m.Reply(F(chatID, "queue_empty"))
		return tg.ErrEndGroup
	}

	r.RemoveFromQueue(-1)

	m.Reply(F(chatID, "clear_success", locales.Arg{
		"user": utils.MentionHTML(m.Sender),
	}))

	return tg.ErrEndGroup
}

func handleMove(m *tg.NewMessage, cplay bool) error {
	chatID := m.ChannelID()

	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return tg.ErrEndGroup
	}

	if !r.IsActiveChat() || r.Track() == nil {
		m.Reply(F(chatID, "queue_no_active"))
		return tg.ErrEndGroup
	}

	if len(r.Queue()) == 0 {
		m.Reply(F(chatID, "queue_empty"))
		return tg.ErrEndGroup
	}

	args := strings.Fields(m.Text())
	if len(args) < 3 {
		m.Reply(F(chatID, "move_usage", locales.Arg{
			"cmd": getCommand(m),
		}))
		return tg.ErrEndGroup
	}

	from, err1 := strconv.Atoi(args[1])
	to, err2 := strconv.Atoi(args[2])
	if err1 != nil || err2 != nil {
		m.Reply(F(chatID, "move_invalid_numbers", locales.Arg{
			"cmd": getCommand(m),
		}))
		return tg.ErrEndGroup
	}

	if from <= 0 || to <= 0 {
		m.Reply(F(chatID, "move_invalid_indexes_min"))
		return tg.ErrEndGroup
	}

	queueLen := len(r.Queue())
	if from > queueLen || to > queueLen {
		m.Reply(F(chatID, "move_invalid_indexes_max", locales.Arg{
			"queue_len": queueLen,
		}))
		return tg.ErrEndGroup
	}

	r.MoveInQueue(from-1, to-1)

	m.Reply(F(chatID, "move_success", locales.Arg{
		"from": from,
		"to":   to,
		"user": utils.MentionHTML(m.Sender),
	}))

	return tg.ErrEndGroup
}
