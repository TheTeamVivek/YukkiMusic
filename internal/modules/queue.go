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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/core"
	models "yukkimusic/internal/core/models"
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
	if !isSuperGroup(c, m) {
		return nil
	}

	if cplay && !filterAuthUsers(c, m) {
		return nil
	}

	chatID := m.ChatID()

	r, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		_, err := m.ReplyText(c, err.Error(), nil)
		return err
	}

	t := r.Track()
	if !r.Active() || t == nil {
		_, err := m.ReplyText(c, F(chatID, "queue_no_active"), nil)
		return err
	}

	_, rerr := m.ReplyRichMessage(c, queueRich(r, queueFirstPage), nil)
	return rerr
}

// trackDeleteToken returns a safe callback identifier for a track.
// For short, safe IDs (like YouTube video IDs), it returns track.ID as-is.
// For long IDs (e.g. Telegram files) or IDs with special characters, it returns "tok_<hash>".
func trackDeleteToken(t *models.Track) string {
	if t == nil {
		return ""
	}
	id := t.ID
	if id == "" {
		id = t.URL
	}
	if len(id) > 0 && len(id) <= 20 && !strings.Contains(id, ":") {
		return id
	}
	h := sha256.Sum256([]byte(id))
	return "tok_" + hex.EncodeToString(h[:10])
}

// queueFirstPage is how many up-next tracks the first page shows before the
// "show more" button reveals the whole remaining queue.
const queueFirstPage = 15

// queueRich builds the rich "current queue" card. It renders the first `shown`
// up-next tracks, each with a danger delete button. When tracks remain hidden a
// "more" callback button is appended to reveal the rest.
func queueRich(r *core.Room, shown int) *td.InputRichMessage {
	chatID := r.ControlChatID
	t := r.Track()
	queue := r.Queue()
	q := len(queue)

	if shown < 1 {
		shown = queueFirstPage
	}
	if shown > q {
		shown = q
	}

	var b strings.Builder

	fmt.Fprintf(&b, "<h3>%s</h3>\n", F(chatID, "queue_header"))
	fmt.Fprintf(
		&b,
		"<aside><b>%s</b><br><a href=\"%s\">%s</a><br>\u23f1\ufe0f %s  \u2022  \U0001F464 %s</aside>\n",
		F(chatID, "queue_now_playing"),
		t.URL,
		utils.EscapeHTML(utils.ShortTitle(t.Title, 40)),
		utils.FormatDuration(t.Duration),
		t.Requester,
	)

	if q > 0 {
		fmt.Fprintf(&b, "<p><b>%s</b></p>\n", F(chatID, "queue_up_next", locales.Arg{"count": q}))
		b.WriteString("<table bordered striped compact>\n")
		b.WriteString("<tr><th>#</th><th>\U0001F3B5 Track</th><th>\u23f1\ufe0f</th><th>\U0001F464</th><th>\U0001F5D1\ufe0f</th></tr>\n")

		for i := 0; i < shown; i++ {
			track := queue[i]
			delID := trackDeleteToken(track)
			if delID == "" {
				delID = strconv.Itoa(i + 1)
			}
			fmt.Fprintf(
				&b,
				"<tr><td>%d</td><td><a href=\"%s\">%s</a></td><td>%s</td><td>%s</td>"+
					"<td><tg-button type=\"callback_data\" style=\"danger\" data=\"qdel:%d:%s:%d\">\u2a2f</tg-button></td></tr>\n",
				i+1,
				track.URL,
				utils.EscapeHTML(utils.ShortTitle(track.Title, 40)),
				utils.FormatDuration(track.Duration),
				track.Requester,
				r.ChatID,
				delID,
				shown,
			)
		}
		b.WriteString("</table>\n")

		if more := q - shown; more > 0 {
			fmt.Fprintf(
				&b,
				"<tg-button-row align=\"center\">"+
					"<tg-button type=\"callback_data\" data=\"qmore:%d\">%s</tg-button>"+
					"</tg-button-row>\n",
				r.ChatID,
				F(chatID, "queue_more_button", locales.Arg{"count": more}),
			)
		}
	} else {
		fmt.Fprintf(&b, "<footer>%s</footer>\n", F(chatID, "queue_empty_tail"))
	}

	return &td.InputRichMessage{
		Source: &td.RichMessageSourceHtml{Text: b.String()},
	}
}

func handleRemove(c *td.Client, m *td.Message, cplay bool) error {
	if !isSuperGroup(c, m) {
		return nil
	}

	if !filterAuthUsers(c, m) {
		return nil
	}

	chatID := m.ChatID()

	r, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		_, err := m.ReplyText(c, err.Error(), nil)
		return err
	}
	t := r.Track()
	if !r.Active() || t == nil {
		_, err := m.ReplyText(c, F(chatID, "queue_no_active"), nil)
		return err
	}

	if r.QueueLen() == 0 {
		_, err := m.ReplyText(c, F(chatID, "queue_empty"), nil)
		return err
	}

	args := strings.Fields(m.Text())
	if len(args) < 2 {
		_, err := m.ReplyText(c, F(chatID, "remove_usage", locales.Arg{
			"cmd": getCommand(m),
		}), nil)
		return err
	}

	index, err := strconv.Atoi(args[1])
	if err != nil {
		_, err := m.ReplyText(c, F(chatID, "remove_invalid_index"), nil)
		return err
	}

	if index <= 0 {
		_, err := m.ReplyText(c, F(chatID, "remove_index_too_small"), nil)
		return err
	}

	total := r.QueueLen()
	if index > total {
		_, err := m.ReplyText(c, F(chatID, "remove_index_too_big", locales.Arg{
			"total": total,
		}), nil)
		return err
	}

	r.RemoveTrack(index - 1)

	sender, _ := m.GetUser(c)
	mention := mention(sender, m.SenderID())
	_, rerr := m.ReplyText(c, F(chatID, "remove_success", locales.Arg{
		"index": index,
		"user":  mention,
	}), nil)
	return rerr
}

func handleClear(c *td.Client, m *td.Message, cplay bool) error {
	if !isSuperGroup(c, m) {
		return nil
	}

	if !filterAuthUsers(c, m) {
		return nil
	}

	chatID := m.ChatID()

	r, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		_, err := m.ReplyText(c, err.Error(), nil)
		return err
	}
	t := r.Track()
	if !r.Active() || t == nil {
		_, err := m.ReplyText(c, F(chatID, "clear_no_active"), nil)
		return err
	}

	if r.QueueLen() == 0 {
		_, err := m.ReplyText(c, F(chatID, "queue_empty"), nil)
		return err
	}

	r.ClearQueue()

	sender, _ := m.GetUser(c)
	mention := mention(sender, m.SenderID())
	_, rerr := m.ReplyText(c, F(chatID, "clear_success", locales.Arg{
		"user": mention,
	}), nil)
	return rerr
}

func handleMove(c *td.Client, m *td.Message, cplay bool) error {
	if !isSuperGroup(c, m) {
		return nil
	}

	if !filterAuthUsers(c, m) {
		return nil
	}

	chatID := m.ChatID()

	r, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		_, err := m.ReplyText(c, err.Error(), nil)
		return err
	}

	if !r.Active() || r.Track() == nil {
		_, err := m.ReplyText(c, F(chatID, "queue_no_active"), nil)
		return err
	}

	if r.QueueLen() == 0 {
		_, err := m.ReplyText(c, F(chatID, "queue_empty"), nil)
		return err
	}

	args := strings.Fields(m.Text())
	if len(args) < 3 {
		_, err := m.ReplyText(c, F(chatID, "move_usage", locales.Arg{
			"cmd": getCommand(m),
		}), nil)
		return err
	}

	from, err1 := strconv.Atoi(args[1])
	to, err2 := strconv.Atoi(args[2])
	if err1 != nil || err2 != nil {
		_, err := m.ReplyText(c, F(chatID, "move_invalid_numbers", locales.Arg{
			"cmd": getCommand(m),
		}), nil)
		return err
	}

	if from <= 0 || to <= 0 {
		_, err := m.ReplyText(c, F(chatID, "move_invalid_indexes_min"), nil)
		return err
	}

	queueLen := r.QueueLen()
	if from > queueLen || to > queueLen {
		_, err := m.ReplyText(c, F(chatID, "move_invalid_indexes_max", locales.Arg{
			"queue_len": queueLen,
		}), nil)
		return err
	}

	r.MoveTrack(from-1, to-1)

	sender, _ := m.GetUser(c)
	mention := mention(sender, m.SenderID())
	_, rerr := m.ReplyText(c, F(chatID, "move_success", locales.Arg{
		"from": from,
		"to":   to,
		"user": mention,
	}), nil)
	return rerr
}
