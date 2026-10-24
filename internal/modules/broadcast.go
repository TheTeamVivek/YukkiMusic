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
	"context"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

const (
	broadcastDelay        = 400 * time.Millisecond // pause between two sends
	broadcastProgressTick = 5 * time.Second        // how often the progress message is edited
	broadcastMaxFloodWait = time.Minute            // longest flood wait that is slept through
)

// activeBroadcast is the running broadcast; only one may run at a time.
var activeBroadcast atomic.Pointer[broadcast]

func init() {
	helpTexts["broadcast"] = `<i>Broadcast a message to all served chats and users.</i>

<u>Usage:</u>
<b>/broadcast [flags] [text]</b> — Broadcast a text message.
<b>/broadcast [flags]</b> (as a reply) — Broadcast the replied message.
<b>/broadcast -cancel</b> — Cancel the ongoing broadcast.

<blockquote>
<b>📋 Flags:</b>
• <code>-nochat</code> — Skip groups
• <code>-nouser</code> — Skip users
• <code>-copy</code> — Send the replied message without the forwarded tag
• <code>-pin</code> — Pin the message silently
• <code>-cancel</code> — Cancel the ongoing broadcast
</blockquote>
<blockquote>
<b>📌 Examples:</b>
/broadcast -nochat Important announcement
/broadcast -copy -pin [reply to message]
</blockquote>
<b>⚠️ Notes:</b>
• Only the <b>owner</b> can use this command
• Flags go first, before the text
• You can cancel with the inline button or <code>/broadcast -cancel</code>
• Only one broadcast can run at a time`

	helpTexts["gcast"] = helpTexts["broadcast"]
	helpTexts["bcast"] = helpTexts["broadcast"]
}

type broadcastFlags struct {
	noChat, noUser, copy, pin, cancel bool
}

type broadcastTarget struct {
	id   int64
	user bool
}

// tally counts the outcome of the sends to one kind of target.
type tally struct{ total, sent, failed int }

func (t tally) line(chatID int64, key string) string {
	done := t.sent + t.failed
	progress := 0.0
	if t.total > 0 {
		progress = float64(done) / float64(t.total) * 100
	}
	return F(chatID, key, locales.Arg{
		"done":     done,
		"total":    t.total,
		"progress": fmt.Sprintf("%.1f", progress),
	})
}

// broadcast is one running broadcast. Only run touches the counters; the
// cancel button just cancels ctx.
type broadcast struct {
	c       *td.Client
	reply   *td.Message // message to forward or copy; nil when text is sent
	text    string
	copy    bool
	pin     bool
	targets []broadcastTarget

	progress     *td.Message
	ctx          context.Context
	cancel       context.CancelFunc
	started      time.Time
	chats, users tally
	errs         strings.Builder // one line per failed target
}

func broadcastHandler(c *td.Client, m *td.Message) error {
	if !checkOwner(c, m) {
		return nil
	}
	chatID := m.ChatID()

	flags, text := parseBroadcast(m)
	if flags.cancel {
		return handleBroadcastCancel(c, m)
	}
	if activeBroadcast.Load() != nil {
		_, err := m.ReplyText(c, F(chatID, "broadcast_already_running"), nil)
		return err
	}

	var reply *td.Message
	if m.ReplyToMessageID() > 0 {
		var err error
		if reply, err = m.GetRepliedMessage(c); err != nil {
			utils.Log.Errorf("broadcast: cannot get the replied message: %v", err)
		}
	}
	if reply == nil && text == "" {
		_, err := m.ReplyText(c, F(chatID, "broadcast_no_content", locales.Arg{"cmd": getCommand(m)}), nil)
		return err
	}

	// Chats first, then users.
	var targets []broadcastTarget
	job := &broadcast{c: c, reply: reply, text: text, copy: flags.copy, pin: flags.pin}
	if !flags.noChat {
		ids, err := database.ServedChats()
		if err != nil {
			return replyBroadcastError(c, m, "broadcast_fetch_chats_failed", err)
		}
		for _, id := range ids {
			targets = append(targets, broadcastTarget{id: id})
		}
		job.chats.total = len(ids)
	}
	if !flags.noUser {
		ids, err := database.ServedUsers()
		if err != nil {
			return replyBroadcastError(c, m, "broadcast_fetch_users_failed", err)
		}
		for _, id := range ids {
			targets = append(targets, broadcastTarget{id: id, user: true})
		}
		job.users.total = len(ids)
	}
	if len(targets) == 0 {
		_, err := m.ReplyText(c, F(chatID, "broadcast_no_targets"), nil)
		return err
	}

	job.targets = targets
	job.started = time.Now()
	job.ctx, job.cancel = context.WithCancel(context.Background())

	if !activeBroadcast.CompareAndSwap(nil, job) {
		job.cancel()
		_, err := m.ReplyText(c, F(chatID, "broadcast_already_running"), nil)
		return err
	}

	progress, err := m.ReplyText(c, F(chatID, "broadcast_initializing"),
		&td.SendMessageOpts{ReplyMarkup: core.GetBroadcastCancelKeyboard(chatID)})
	if err != nil {
		activeBroadcast.CompareAndSwap(job, nil)
		job.cancel()
		utils.Log.Errorf("broadcast: failed to send the progress message: %v", err)
		return nil
	}
	job.progress = progress

	go job.run()
	return nil
}

// parseBroadcast reads the flags at the start of the command and returns them
// with the text that follows. A dash-word that is not a flag starts the text.
func parseBroadcast(m *td.Message) (broadcastFlags, string) {
	var f broadcastFlags
	rest := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(m.Text()), getCommand(m)))

	for strings.HasPrefix(rest, "-") {
		word, tail := rest, ""
		if i := strings.IndexAny(rest, " \t\r\n"); i >= 0 {
			word, tail = rest[:i], rest[i+1:]
		}

		switch strings.ToLower(strings.TrimLeft(word, "-")) {
		case "nochat":
			f.noChat = true
		case "nouser":
			f.noUser = true
		case "copy":
			f.copy = true
		case "pin":
			f.pin = true
		case "cancel":
			f.cancel = true
		default:
			return f, rest
		}
		rest = strings.TrimSpace(tail)
	}
	return f, rest
}

func replyBroadcastError(c *td.Client, m *td.Message, key string, err error) error {
	_, rerr := m.ReplyText(c, F(m.ChatID(), key, locales.Arg{"error": html.EscapeString(err.Error())}), nil)
	if rerr != nil {
		return rerr
	}
	return err
}

func (b *broadcast) run() {
	defer activeBroadcast.CompareAndSwap(b, nil)
	defer func() {
		if r := recover(); r != nil {
			utils.Log.Errorf("broadcast: panic recovered: %v", r)
		}
		b.report(true)
		b.sendErrors()
		b.cancel()
	}()

	lastReport := time.Now()
	for _, t := range b.targets {
		if b.ctx.Err() != nil {
			return
		}

		counts := &b.chats
		if t.user {
			counts = &b.users
		}
		if err := b.send(t); err != nil {
			counts.failed++
			fmt.Fprintf(&b.errs, "[%d] - [%v]\n", t.id, err)
			utils.Log.Debugf("broadcast: failed for %d: %v", t.id, err)
		} else {
			counts.sent++
		}

		if time.Since(lastReport) >= broadcastProgressTick {
			b.report(false)
			lastReport = time.Now()
		}
		if !b.sleep(broadcastDelay) {
			return
		}
	}
}

// send delivers the broadcast to one target, waiting out a short flood wait.
func (b *broadcast) send(t broadcastTarget) error {
	sent, err := b.deliver(t.id)
	if err != nil {
		wait := time.Duration(getFloodWait(err)) * time.Second
		if wait <= 0 || wait > broadcastMaxFloodWait || !b.sleep(wait) {
			return err
		}
		if sent, err = b.deliver(t.id); err != nil {
			return err
		}
	}

	if b.pin {
		if err := b.c.PinChatMessage(t.id, sent.Id, &td.PinChatMessageOpts{DisableNotification: true}); err != nil {
			utils.Log.Debugf("broadcast: pin failed for %d: %v", t.id, err)
		}
	}
	return nil
}

func (b *broadcast) deliver(id int64) (*td.Message, error) {
	switch {
	case b.reply == nil:
		return b.c.SendMessage(id, b.text, nil)
	case b.copy:
		return b.reply.Copy(b.c, id, &td.CopyMessageOpts{})
	default:
		return b.reply.Forward(b.c, id, &td.ForwardMessageOpts{})
	}
}

// sleep waits for d and reports false if the broadcast was cancelled meanwhile.
func (b *broadcast) sleep(d time.Duration) bool {
	select {
	case <-b.ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

// report edits the progress message; the final report drops the cancel button.
func (b *broadcast) report(final bool) {
	chatID := b.progress.ChatID()
	var sb strings.Builder

	switch {
	case !final:
		sb.WriteString(F(chatID, "broadcast_progress_header") + "\n\n")
	case b.ctx.Err() != nil:
		sb.WriteString(F(chatID, "broadcast_cancel_done") + "\n\n")
	}

	if b.chats.total > 0 {
		sb.WriteString(b.chats.line(chatID, "broadcast_total_chats") + "\n")
	}
	if b.users.total > 0 {
		sb.WriteString(b.users.line(chatID, "broadcast_total_users") + "\n")
	}
	sb.WriteString("\n")

	if b.chats.failed > 0 {
		sb.WriteString(F(chatID, "broadcast_failed_chats", locales.Arg{"count": b.chats.failed}) + "\n")
	}
	if b.users.failed > 0 {
		sb.WriteString(F(chatID, "broadcast_failed_users", locales.Arg{"count": b.users.failed}) + "\n")
	}
	sb.WriteString(F(chatID, "broadcast_elapsed", locales.Arg{
		"elapsed": utils.FormatDuration(int(time.Since(b.started).Seconds())),
	}))

	if final {
		sent, total := b.chats.sent+b.users.sent, b.chats.total+b.users.total
		rate := 0.0
		if total > 0 {
			rate = float64(sent) / float64(total) * 100
		}
		sb.WriteString("\n\n" + F(chatID, "broadcast_success_rate", locales.Arg{
			"rate":  fmt.Sprintf("%.1f", rate),
			"sent":  sent,
			"total": total,
		}))
	}

	var opts *td.EditMessageTextOpts
	if !final {
		opts = &td.EditMessageTextOpts{ReplyMarkup: core.GetBroadcastCancelKeyboard(chatID)}
	}
	_, _ = b.progress.EditText(b.c, sb.String(), opts)
}

// sendErrors replies to the progress message with the failures as a file.
func (b *broadcast) sendErrors() {
	if b.errs.Len() == 0 {
		return
	}

	file := filepath.Join(os.TempDir(), "broadcast_errors.txt")
	if err := os.WriteFile(file, []byte(b.errs.String()), 0o600); err != nil {
		utils.Log.Errorf("broadcast: failed to write the error file: %v", err)
		return
	}
	defer os.Remove(file)

	_, _ = b.progress.ReplyDocument(b.c, &td.InputFileLocal{Path: file}, nil)
}

// stopBroadcast cancels the running broadcast, reporting whether there was one.
func stopBroadcast() bool {
	b := activeBroadcast.Load()
	if b == nil {
		return false
	}
	b.cancel()
	return true
}

func handleBroadcastCancel(c *td.Client, m *td.Message) error {
	key := "broadcast_not_running"
	if stopBroadcast() {
		key = "broadcast_cancel_success"
	}
	_, err := m.ReplyText(c, F(m.ChatID(), key), nil)
	return err
}

func broadcastCancelCB(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	key := "broadcast_cancel_done"
	switch {
	case cb.SenderUserId != config.OwnerID:
		key = "broadcast_cancel_owner_only"
	case !stopBroadcast():
		key = "broadcast_cancel_none_running"
	}
	cb.Answer(c, F(cb.ChatId, key), &td.CallbackAnswerOpts{ShowAlert: true})
	return nil
}
