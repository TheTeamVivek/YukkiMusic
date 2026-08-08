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
	"strconv"
	"strings"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

func queueHandler(c *td.Client, m *td.Message) error {
	return handleQueue(c, m, false)
}

func cqueueHandler(c *td.Client, m *td.Message) error {
	return handleQueue(c, m, true)
}

func removeHandler(c *td.Client, m *td.Message) error {
	return handleRemove(c, m, false)
}

func cremoveHandler(c *td.Client, m *td.Message) error {
	return handleRemove(c, m, true)
}

func moveHandler(c *td.Client, m *td.Message) error {
	return handleMove(c, m, false)
}

func cmoveHandler(c *td.Client, m *td.Message) error {
	return handleMove(c, m, true)
}

func clearHandler(c *td.Client, m *td.Message) error {
	return handleClear(c, m, false)
}

func cclearHandler(c *td.Client, m *td.Message) error {
	return handleClear(c, m, true)
}

func handleQueue(c *td.Client, m *td.Message, cplay bool) error {
	if !isSuperGroupTd(c, m) {
		return nil
	}

	if cplay && !filterAuthUsersTd(c, m) {
		return nil
	}

	chatID := m.ChatID()

	r, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		m.ReplyText(c, err.Error(), nil)
		return nil
	}

	t := r.Track()
	if !r.IsActiveChat() || t == nil {
		m.ReplyText(c, F(chatID, "queue_no_active"), nil)
		return nil
	}

	var b strings.Builder

	b.WriteString(F(chatID, "queue_header"))
	b.WriteString("\n\n")

	b.WriteString(F(chatID, "queue_now_playing"))
	b.WriteString("\n")

	fmt.Fprintf(
		&b,
		"🎧 <a href=\"%s\">%s</a> — %s [%s]\n\n",
		t.URL,
		utils.EscapeHTML(utils.ShortTitle(t.Title, 35)),
		t.Requester,
		utils.FormatDuration(t.Duration),
	)

	queue := r.Queue()
	q := len(queue)
	useQuote := q >= 3

	if q > 0 {
		b.WriteString(F(chatID, "queue_up_next"))
		n := "\n"
		if !useQuote {
			n += "\n"
		}
		b.WriteString(n)

		if useQuote {
			b.WriteString("<blockquote>")
		}

		for i, track := range queue {
			if i >= 10 {
				break
			}

			fmt.Fprintf(
				&b,
				"%d. 🎵 <a href=\"%s\">%s</a> — %s [%s]\n",
				i+1,
				track.URL,
				utils.EscapeHTML(utils.ShortTitle(track.Title, 35)),
				track.Requester,
				utils.FormatDuration(track.Duration),
			)
		}

		if useQuote {
			b.WriteString("</blockquote>")
		}

		if q > 10 {
			var full strings.Builder

			full.WriteString(F(chatID, "queue_header"))
			full.WriteString("\n\n")

			full.WriteString(F(chatID, "queue_now_playing"))
			full.WriteString("\n")

			fmt.Fprintf(
				&full,
				"🎧 %s — %s [%s]\n\n",
				t.Title,
				t.Requester,
				utils.FormatDuration(t.Duration),
			)

			full.WriteString(F(chatID, "queue_up_next"))
			full.WriteString("\n\n")

			for i, track := range queue {
				fmt.Fprintf(
					&full,
					"%d. %s — %s [%s]\n",
					i+1,
					track.Title,
					track.Requester,
					utils.FormatDuration(track.Duration),
				)
			}

			link, err := utils.CreatePaste(full.String())
			remaining := q - 10

			if err == nil && link != "" {
				more := fmt.Sprintf("<a href=\"%s\">%d</a>", link, remaining)

				b.WriteString(F(chatID, "queue_more_line", locales.Arg{
					"remaining": more,
				}))
			} else {
				b.WriteString(F(chatID, "queue_more_line", locales.Arg{
					"remaining": remaining,
				}))
			}
		}
	} else {
		b.WriteString(F(chatID, "queue_empty_tail"))
	}

	m.ReplyText(c, b.String(), nil)
	return nil
}

func handleRemove(c *td.Client, m *td.Message, cplay bool) error {
	if !isSuperGroupTd(c, m) {
		return nil
	}

	if !filterAuthUsersTd(c, m) {
		return nil
	}

	chatID := m.ChatID()

	r, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		m.ReplyText(c, err.Error(), nil)
		return nil
	}
	t := r.Track()
	if !r.IsActiveChat() || t == nil {
		m.ReplyText(c, F(chatID, "queue_no_active"), nil)
		return nil
	}

	if len(r.Queue()) == 0 {
		m.ReplyText(c, F(chatID, "queue_empty"), nil)
		return nil
	}

	args := strings.Fields(m.Text())
	if len(args) < 2 {
		m.ReplyText(c, F(chatID, "remove_usage", locales.Arg{
			"cmd": getCommandTd(m),
		}), nil)
		return nil
	}

	index, err := strconv.Atoi(args[1])
	if err != nil {
		m.ReplyText(c, F(chatID, "remove_invalid_index"), nil)
		return nil
	}

	if index <= 0 {
		m.ReplyText(c, F(chatID, "remove_index_too_small"), nil)
		return nil
	}

	total := len(r.Queue())
	if index > total {
		m.ReplyText(c, F(chatID, "remove_index_too_big", locales.Arg{
			"total": total,
		}), nil)
		return nil
	}

	r.RemoveFromQueue(index - 1)

	sender, _ := m.GetUser(c)
	mention := mentionOf(sender, m.SenderID())

	m.ReplyText(c, F(chatID, "remove_success", locales.Arg{
		"index": index,
		"user":  mention,
	}), nil)

	return nil
}

func handleClear(c *td.Client, m *td.Message, cplay bool) error {
	if !isSuperGroupTd(c, m) {
		return nil
	}

	if !filterAuthUsersTd(c, m) {
		return nil
	}

	chatID := m.ChatID()

	r, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		m.ReplyText(c, err.Error(), nil)
		return nil
	}
	t := r.Track()
	if !r.IsActiveChat() || t == nil {
		m.ReplyText(c, F(chatID, "clear_no_active"), nil)
		return nil
	}

	if len(r.Queue()) == 0 {
		m.ReplyText(c, F(chatID, "queue_empty"), nil)
		return nil
	}

	r.SetData("last_queue", r.Queue())
	r.RemoveFromQueue(-1)

	restoreCmd := "restore"
	if cplay {
		restoreCmd = "crestore"
	}

	sender, _ := m.GetUser(c)
	mention := mentionOf(sender, m.SenderID())

	m.ReplyText(c, F(chatID, "clear_success", locales.Arg{
		"user": mention,
		"cmd":  restoreCmd,
	}), nil)

	return nil
}

func handleMove(c *td.Client, m *td.Message, cplay bool) error {
	if !isSuperGroupTd(c, m) {
		return nil
	}

	if !filterAuthUsersTd(c, m) {
		return nil
	}

	chatID := m.ChatID()

	r, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		m.ReplyText(c, err.Error(), nil)
		return nil
	}

	if !r.IsActiveChat() || r.Track() == nil {
		m.ReplyText(c, F(chatID, "queue_no_active"), nil)
		return nil
	}

	if len(r.Queue()) == 0 {
		m.ReplyText(c, F(chatID, "queue_empty"), nil)
		return nil
	}

	args := strings.Fields(m.Text())
	if len(args) < 3 {
		m.ReplyText(c, F(chatID, "move_usage", locales.Arg{
			"cmd": getCommandTd(m),
		}), nil)
		return nil
	}

	from, err1 := strconv.Atoi(args[1])
	to, err2 := strconv.Atoi(args[2])
	if err1 != nil || err2 != nil {
		m.ReplyText(c, F(chatID, "move_invalid_numbers", locales.Arg{
			"cmd": getCommandTd(m),
		}), nil)
		return nil
	}

	if from <= 0 || to <= 0 {
		m.ReplyText(c, F(chatID, "move_invalid_indexes_min"), nil)
		return nil
	}

	queueLen := len(r.Queue())
	if from > queueLen || to > queueLen {
		m.ReplyText(c, F(chatID, "move_invalid_indexes_max", locales.Arg{
			"queue_len": queueLen,
		}), nil)
		return nil
	}

	r.MoveInQueue(from-1, to-1)

	sender, _ := m.GetUser(c)
	mention := mentionOf(sender, m.SenderID())

	m.ReplyText(c, F(chatID, "move_success", locales.Arg{
		"from": from,
		"to":   to,
		"user": mention,
	}), nil)

	return nil
}
