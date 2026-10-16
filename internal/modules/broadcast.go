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
	"strconv"
	"strings"
	"sync"
	"time"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

const (
	broadcastDefaultDelay = 0.7
	broadcastBatchSize    = 30
	broadcastBatchPause   = 7500 * time.Millisecond
	broadcastProgressTick = 5 * time.Second
)

// Only one broadcast may run at a time; bcastJob holds the active one.
var (
	bcastMu  sync.Mutex
	bcastJob *broadcast
)

type broadcastFlags struct {
	noChat  bool
	noUser  bool
	copy    bool
	pin     bool
	pinLoud bool
	limit   int
	delay   float64
}

type broadcastTarget struct {
	id     int64
	isUser bool
}

// broadcast is a single running broadcast and its live counters.
type broadcast struct {
	client   *td.Client
	source   *td.Message
	progress *td.Message
	flags    broadcastFlags
	content  string

	ctx    context.Context
	cancel context.CancelFunc

	mu          sync.Mutex
	totalChats  int
	totalUsers  int
	doneChats   int
	doneUsers   int
	failedChats []int64
	failedUsers []int64
	errs        strings.Builder
	startedAt   time.Time
	cancelled   bool
}

func init() {
	helpTexts["broadcast"] = `<i>Broadcast a message to all served chats and users.</i>

<u>Usage:</u>
<b>/broadcast [flags] [text] </b> — Broadcast text message.
<b>/broadcast [flags] [reply to message]</b> — Broadcast the replied message.
<b>/broadcast -cancel</b> — Cancel ongoing broadcast.

<blockquote>
<b>📋 Flags:</b>
• <code>--nochat</code> — Exclude groups from broadcast
• <code>--nouser</code> — Exclude users from broadcast
• <code>--copy</code> — Remove forwarded tag, when broadcasting a replied message (copy mode)
• <code>--limit [n]</code> — Limit total messages sent (default: 0 = no limit)
• <code>--delay [seconds]</code> — Delay between messages (default: 0.7s)
• <code>--pin</code> — Pin the message (silent)
• <code>--pinloud</code> — Pin the message (with notification)

• <code>-cancel</code> - Cancel a ongoing broadcast.
</blockquote>
<blockquote>
<b>📌 Examples:</b>
/broadcast -nochat -delay 2 Important announcement
/broadcast -copy -nochat -pin [reply to message]
/broadcast -limit 10 -delay 3 Limited broadcast
</blockquote>
<b>⚠️ Notes:</b>
• Only the <b>owner</b> can use this command
• After every 30 messages, there's an automatic 7.5s pause
• You can cancel ongoing broadcasts using the inline button or <code>/broadcast -cancel</code>
• Only one broadcast can run at a time`

	helpTexts["gcast"] = helpTexts["broadcast"]
	helpTexts["bcast"] = helpTexts["broadcast"]
}

func broadcastHandler(c *td.Client, m *td.Message) error {
	if !checkOwner(c, m) {
		return nil
	}
	chatID := m.ChatID()

	if strings.Contains(strings.ToLower(m.Text()), "-cancel") {
		return handleBroadcastCancel(c, m)
	}
	if broadcastRunning() {
		_, err := m.ReplyText(c, F(chatID, "broadcast_already_running"), nil)
		return err
	}

	flags, content, err := parseBroadcastCommand(m)
	if err != nil {
		_, err := m.ReplyText(c, F(chatID, "broadcast_parse_failed", locales.Arg{
			"error": html.EscapeString(err.Error()),
		}), nil)
		return err
	}
	if content == "" && m.ReplyToMessageID() == 0 {
		_, err := m.ReplyText(c, F(chatID, "broadcast_no_content", locales.Arg{
			"cmd": getCommand(m),
		}), nil)
		return err
	}

	targets, err := loadBroadcastTargets(c, m, flags)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		_, err := m.ReplyText(c, F(chatID, "broadcast_no_targets"), nil)
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	job := &broadcast{
		client:      c,
		source:      m,
		flags:       flags,
		content:     content,
		ctx:         ctx,
		cancel:      cancel,
		startedAt:   time.Now(),
		failedChats: make([]int64, 0),
		failedUsers: make([]int64, 0),
	}
	for _, t := range targets {
		if t.isUser {
			job.totalUsers++
		} else {
			job.totalChats++
		}
	}

	if !reserveBroadcast(job) {
		cancel()
		_, err := m.ReplyText(c, F(chatID, "broadcast_already_running"), nil)
		return err
	}

	job.progress, err = m.ReplyText(c, F(chatID, "broadcast_initializing"),
		&td.SendMessageOpts{ReplyMarkup: core.GetBroadcastCancelKeyboard(chatID)})
	if err != nil {
		job.release()
		job.cancel()
		utils.Log.Errorf("Failed to send broadcast progress message: %v", err)
		return nil
	}

	go job.run(targets)
	return nil
}

func parseBroadcastCommand(m *td.Message) (broadcastFlags, string, error) {
	flags := broadcastFlags{delay: broadcastDefaultDelay}

	text := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(m.Text()), getCommand(m)))
	lines := strings.Split(text, "\n")
	words := strings.Fields(lines[0])

	var content []string
	for i := 0; i < len(words); i++ {
		switch w := strings.ToLower(words[i]); w {
		case "-nochat", "--nochat":
			flags.noChat = true
		case "-nouser", "--nouser":
			flags.noUser = true
		case "-copy", "--copy":
			flags.copy = true
		case "-pin", "--pin":
			flags.pin = true
		case "-pinloud", "--pinloud":
			flags.pinLoud = true
		case "-cancel", "--cancel":
			// handled before parsing
		case "-limit", "--limit", "-delay", "--delay":
			if i+1 >= len(words) {
				return flags, "", fmt.Errorf("missing value for %s", w)
			}
			val := words[i+1]
			if w == "-limit" || w == "--limit" {
				limit, err := strconv.Atoi(val)
				if err != nil || limit < 0 {
					return flags, "", fmt.Errorf("invalid limit value: %s", val)
				}
				flags.limit = limit
			} else {
				delay, err := strconv.ParseFloat(val, 64)
				if err != nil || delay < 0 {
					return flags, "", fmt.Errorf("invalid delay value: %s", val)
				}
				flags.delay = delay
			}
			i++
		default:
			content = append(content, words[i])
		}
	}

	body := strings.Join(content, " ")
	if len(lines) > 1 {
		if body != "" {
			body += "\n"
		}
		body += strings.Join(lines[1:], "\n")
	}
	return flags, strings.TrimSpace(body), nil
}

// loadBroadcastTargets collects the recipients (chats first, then users),
// applies the limit and reports any lookup failure to the caller's chat.
func loadBroadcastTargets(c *td.Client, m *td.Message, flags broadcastFlags) ([]broadcastTarget, error) {
	var targets []broadcastTarget

	if !flags.noChat {
		chats, err := database.ServedChats()
		if err != nil {
			m.ReplyText(c, F(m.ChatID(), "broadcast_fetch_chats_failed", locales.Arg{
				"error": html.EscapeString(err.Error()),
			}), nil)
			return nil, err
		}
		for _, id := range chats {
			targets = append(targets, broadcastTarget{id: id})
		}
	}

	if !flags.noUser {
		users, err := database.ServedUsers()
		if err != nil {
			m.ReplyText(c, F(m.ChatID(), "broadcast_fetch_users_failed", locales.Arg{
				"error": html.EscapeString(err.Error()),
			}), nil)
			return nil, err
		}
		for _, id := range users {
			targets = append(targets, broadcastTarget{id: id, isUser: true})
		}
	}

	if flags.limit > 0 && len(targets) > flags.limit {
		targets = targets[:flags.limit]
	}
	return targets, nil
}

func reserveBroadcast(job *broadcast) bool {
	bcastMu.Lock()
	defer bcastMu.Unlock()
	if bcastJob != nil {
		return false
	}
	bcastJob = job
	return true
}

func (b *broadcast) release() {
	bcastMu.Lock()
	if bcastJob == b {
		bcastJob = nil
	}
	bcastMu.Unlock()
}

func broadcastRunning() bool {
	bcastMu.Lock()
	defer bcastMu.Unlock()
	return bcastJob != nil
}

// stopBroadcast cancels the running broadcast so its summary shows cancelled.
func stopBroadcast() bool {
	bcastMu.Lock()
	job := bcastJob
	bcastJob = nil
	bcastMu.Unlock()
	if job == nil {
		return false
	}
	job.mu.Lock()
	job.cancelled = true
	job.mu.Unlock()
	job.cancel()
	return true
}

func (b *broadcast) run(targets []broadcastTarget) {
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		b.pumpProgress(done)
	}()

	defer func() {
		if r := recover(); r != nil {
			utils.Log.Errorf("Broadcast panic recovered: %v", r)
		}
		close(done)
		wg.Wait()
		b.finish()
		b.release()
	}()

	for i, t := range targets {
		if b.ctx.Err() != nil {
			return
		}
		b.deliver(t)
		if !b.pause(i + 1) {
			return
		}
	}
}

func (b *broadcast) deliver(t broadcastTarget) {
	var (
		sent *td.Message
		err  error
	)

	if b.source.ReplyToMessageID() > 0 {
		res, ferr := b.client.ForwardMessages(
			t.id,
			b.source.ChatID(),
			[]int64{b.source.ReplyToMessageID()},
			&td.ForwardMessagesOpts{SendCopy: b.flags.copy},
		)
		if ferr != nil {
			err = ferr
		} else if len(res.Messages) > 0 {
			sent = &res.Messages[0]
		}
	} else {
		sent, err = b.client.SendMessage(t.id, b.content, nil)
	}

	if err != nil {
		if !skipBroadcastError(err) {
			utils.Log.Errorf("Broadcast failed for %d: %v", t.id, err)
		}
	} else if sent != nil && (b.flags.pin || b.flags.pinLoud) {
		if perr := b.client.PinChatMessage(
			t.id,
			sent.Id,
			&td.PinChatMessageOpts{DisableNotification: !b.flags.pinLoud},
		); perr != nil {
			utils.Log.Errorf("Pin failed for %d: %v", t.id, perr)
		}
	}

	b.record(t, err)
}

func (b *broadcast) record(t broadcastTarget, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err != nil {
		if t.isUser {
			b.failedUsers = append(b.failedUsers, t.id)
		} else {
			b.failedChats = append(b.failedChats, t.id)
		}
		fmt.Fprintf(&b.errs, "[%d] - [%v]\n", t.id, err)
		return
	}
	if t.isUser {
		b.doneUsers++
	} else {
		b.doneChats++
	}
}

// skipBroadcastError reports errors that are expected for removed/blocked targets.
func skipBroadcastError(err error) bool {
	return strings.Contains(err.Error(), "USER_IS_BLOCKED") ||
		strings.Contains(err.Error(), "CHAT_WRITE_FORBIDDEN") ||
		strings.Contains(err.Error(), "USER_IS_DEACTIVATED")
}

// pause sleeps the configured delay, pausing longer after every 30 sends.
func (b *broadcast) pause(sent int) bool {
	d := time.Duration(b.flags.delay * float64(time.Second))
	if sent%broadcastBatchSize == 0 {
		d = broadcastBatchPause
	}
	select {
	case <-b.ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

func (b *broadcast) pumpProgress(done <-chan struct{}) {
	ticker := time.NewTicker(broadcastProgressTick)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			b.mu.Lock()
			text := b.format(false)
			b.mu.Unlock()
			_, _ = b.progress.EditText(b.client, text, &td.EditMessageTextOpts{
				ReplyMarkup: core.GetBroadcastCancelKeyboard(b.progress.ChatID()),
			})
		}
	}
}

func (b *broadcast) finish() {
	b.mu.Lock()
	text := b.format(true)
	errs := b.errs.String()
	b.mu.Unlock()

	_, _ = b.progress.EditText(b.client, text, nil)

	if errs == "" {
		return
	}

	const file = "broadcast_errors.txt"
	if err := os.WriteFile(file, []byte(errs), 0o600); err != nil {
		utils.Log.Errorf("Failed to write broadcast errors: %v", err)
		return
	}
	defer os.Remove(file)

	_, _ = b.progress.ReplyDocument(b.client, &td.InputFileLocal{Path: file}, nil)
}

func (b *broadcast) format(final bool) string {
	var sb strings.Builder
	chatID := b.progress.ChatID()

	if !final {
		sb.WriteString(F(chatID, "broadcast_progress_header") + "\n\n")
	} else if b.cancelled {
		sb.WriteString(F(chatID, "broadcast_cancel_done") + "\n\n")
	}

	sb.WriteString(broadcastProgressLine(chatID, "broadcast_total_chats", b.doneChats, b.totalChats) + "\n")
	sb.WriteString(broadcastProgressLine(chatID, "broadcast_total_users", b.doneUsers, b.totalUsers) + "\n\n")

	if len(b.failedChats) > 0 {
		sb.WriteString(F(chatID, "broadcast_failed_chats", locales.Arg{"count": len(b.failedChats)}) + "\n")
	}
	if len(b.failedUsers) > 0 {
		sb.WriteString(F(chatID, "broadcast_failed_users", locales.Arg{"count": len(b.failedUsers)}) + "\n")
	}
	if len(b.failedChats) > 0 || len(b.failedUsers) > 0 {
		sb.WriteString("\n")
	}

	elapsed := time.Since(b.startedAt)
	done := b.doneChats + b.doneUsers
	total := b.totalChats + b.totalUsers

	sb.WriteString(F(chatID, "broadcast_delay", locales.Arg{"delay": fmt.Sprintf("%.1f", b.flags.delay)}) + "\n")
	sb.WriteString(F(chatID, "broadcast_elapsed", locales.Arg{"elapsed": utils.FormatDuration(int(elapsed.Seconds()))}) + "\n")

	if !final && done > 0 && elapsed.Seconds() > 0 && done < total {
		rate := float64(done) / elapsed.Seconds()
		eta := time.Duration(float64(total-done) / rate * float64(time.Second))
		sb.WriteString("\n" + F(chatID, "broadcast_eta", locales.Arg{"eta": utils.FormatDuration(int(eta.Seconds()))}))
	}

	if final {
		failed := len(b.failedChats) + len(b.failedUsers)
		sent := done - failed

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

	return sb.String()
}

func broadcastProgressLine(chatID int64, key string, done, total int) string {
	progress := 0.0
	if total > 0 {
		progress = float64(done) / float64(total) * 100
	}
	return F(chatID, key, locales.Arg{
		"done":     done,
		"total":    total,
		"progress": fmt.Sprintf("%.1f", progress),
	})
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
	if cb.SenderUserId != config.OwnerID {
		cb.Answer(c, 0, true, F(cb.ChatId, "broadcast_cancel_owner_only"), "")
		return nil
	}
	if !stopBroadcast() {
		cb.Answer(c, 0, true, F(cb.ChatId, "broadcast_cancel_none_running"), "")
		return nil
	}
	cb.Answer(c, 0, true, F(cb.ChatId, "broadcast_cancel_done"), "")
	cb.EditMessageText(c, F(cb.ChatId, "broadcast_cancel_done"), nil)
	return nil
}
