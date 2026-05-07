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

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

// BroadcastManager manages the state and lifecycle of a broadcast operation.
type BroadcastManager struct {
	mu     sync.Mutex
	active bool
	cancel context.CancelFunc
	ctx    context.Context
}

var bManager = &BroadcastManager{}

const defaultDelay = 0.7

// TryStart attempts to start a new broadcast. It returns true if successful,
// or false if a broadcast is already running.
func (bm *BroadcastManager) TryStart(
	ctx context.Context,
	cancel context.CancelFunc,
) bool {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	if bm.active {
		return false
	}
	bm.active = true
	bm.ctx = ctx
	bm.cancel = cancel
	return true
}

// Stop cancels the ongoing broadcast and resets the manager's state.
func (bm *BroadcastManager) Stop() {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	if bm.cancel != nil {
		bm.cancel()
	}
	bm.active = false
	bm.ctx = nil
	bm.cancel = nil
}

// IsActive returns whether a broadcast is currently running.
func (bm *BroadcastManager) IsActive() bool {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	return bm.active
}

type BroadcastStats struct {
	TotalChats  int
	TotalUsers  int
	DoneChats   int
	DoneUsers   int
	FailedChats []int64
	FailedUsers []int64
	errorLog    strings.Builder
	Delay       float64
	StartTime   time.Time
	LastUpdate  time.Time
	Finished    bool
	mu          sync.Mutex
}

type BroadcastFlags struct {
	NoChat  bool
	NoUser  bool
	Copy    bool
	Limit   int
	Delay   float64
	Pin     bool
	PinLoud bool
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
• <code>--delay [seconds]</code> — Delay between messages (default: 1.5s)
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

func broadcastHandler(m *tg.NewMessage) error {
	// Check for cancel flag
	text := strings.ToLower(m.Text())
	chatID := m.ChannelID()
	if strings.Contains(text, "-cancel") || strings.Contains(text, "--cancel") {
		return handleBroadcastCancel(m)
	}

	// Check if broadcast is already running
	if bManager.IsActive() {
		m.Reply(F(chatID, "broadcast_already_running"))
		return tg.ErrEndGroup
	}

	// Parse flags and content
	flags, content, err := parseBroadcastCommand(m)
	if err != nil {
		m.Reply(F(chatID, "broadcast_parse_failed", locales.Arg{
			"error": html.EscapeString(err.Error()),
		}))
		return tg.ErrEndGroup
	}

	if content == "" && !m.IsReply() {
		m.Reply(F(chatID, "broadcast_no_content", locales.Arg{
			"cmd": getCommand(m),
		}))
		return tg.ErrEndGroup
	}

	// Get served chats and users
	var servedChats, servedUsers []int64
	var servedChatErr, servedUserErr error

	if !flags.NoChat {
		servedChats, servedChatErr = database.ServedChats()
		if servedChatErr != nil {
			m.Reply(F(chatID, "broadcast_fetch_chats_failed", locales.Arg{
				"error": html.EscapeString(servedChatErr.Error()),
			}))

			return tg.ErrEndGroup
		}
	}

	if !flags.NoUser {
		servedUsers, servedUserErr = database.ServedUsers()
		if servedUserErr != nil {
			m.Reply(F(chatID, "broadcast_fetch_users_failed", locales.Arg{
				"error": html.EscapeString(servedUserErr.Error()),
			}))
			return tg.ErrEndGroup
		}
	}

	// Check if there are any targets
	if len(servedChats) == 0 && len(servedUsers) == 0 {
		m.Reply(F(chatID, "broadcast_no_targets"))
		return tg.ErrEndGroup
	}

	// Apply limit if specified
	totalTargets := len(servedChats) + len(servedUsers)
	if flags.Limit > 0 && totalTargets > flags.Limit {
		if len(servedChats) >= flags.Limit {
			servedChats = servedChats[:flags.Limit]
			servedUsers = nil
		} else {
			remaining := flags.Limit - len(servedChats)
			if len(servedUsers) > remaining {
				servedUsers = servedUsers[:remaining]
			}
		}
	}

	now := time.Now()
	stats := &BroadcastStats{
		TotalChats:  len(servedChats),
		TotalUsers:  len(servedUsers),
		Delay:       flags.Delay,
		StartTime:   now,
		LastUpdate:  now,
		FailedChats: make([]int64, 0),
		FailedUsers: make([]int64, 0),
	}

	ctx, cancel := context.WithCancel(context.Background())
	if !bManager.TryStart(ctx, cancel) {
		m.Reply(F(chatID, "broadcast_already_running"))
		return tg.ErrEndGroup
	}

	progressMsg, err := m.Reply(F(chatID, "broadcast_initializing"),
		&tg.SendOptions{
			ReplyMarkup: core.GetBroadcastCancelKeyboard(chatID),
		})
	if err != nil {
		bManager.Stop()
		gologging.ErrorF("Failed to send broadcast progress message: %v", err)
		return tg.ErrEndGroup
	}

	// Start progress updater in goroutine
	go bManager.updateProgress(ctx, progressMsg, stats)

	// Start broadcast in goroutine
	go bManager.start(
		ctx,
		m,
		progressMsg,
		flags,
		content,
		servedChats,
		servedUsers,
		stats,
	)

	return tg.ErrEndGroup
}

func parseBroadcastCommand(m *tg.NewMessage) (*BroadcastFlags, string, error) {
	flags := &BroadcastFlags{
		Delay: defaultDelay,
	}

	text := strings.TrimSpace(m.Text())
	text = strings.TrimPrefix(text, m.GetCommand())
	text = strings.TrimSpace(text)

	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return flags, "", nil
	}

	words := strings.Fields(lines[0])
	var contentWords []string

	for i := 0; i < len(words); i++ {
		word := strings.ToLower(words[i])
		switch {
		case word == "-nochat" || word == "--nochat":
			flags.NoChat = true
		case word == "-nouser" || word == "--nouser":
			flags.NoUser = true
		case word == "-copy" || word == "--copy":
			flags.Copy = true
		case word == "-pin" || word == "--pin":
			flags.Pin = true
		case word == "-pinloud" || word == "--pinloud":
			flags.PinLoud = true
		case word == "-limit" || word == "--limit":
			if i+1 >= len(words) {
				return nil, "", fmt.Errorf("%s requires a value", words[i])
			}
			limit, err := strconv.Atoi(words[i+1])
			if err != nil || limit < 0 {
				return nil, "", fmt.Errorf(
					"invalid limit value: %s",
					words[i+1],
				)
			}
			flags.Limit = limit
			i++
		case word == "-delay" || word == "--delay":
			if i+1 >= len(words) {
				return nil, "", fmt.Errorf("%s requires a value", words[i])
			}
			delay, err := strconv.ParseFloat(words[i+1], 64)
			if err != nil || delay < 0 {
				return nil, "", fmt.Errorf(
					"invalid delay value: %s",
					words[i+1],
				)
			}
			flags.Delay = delay
			i++
		case word == "-cancel" || word == "--cancel":
			// Handled by the caller
			continue
		default:
			contentWords = append(contentWords, words[i])
		}
	}

	var content strings.Builder
	content.WriteString(strings.Join(contentWords, " "))
	if len(lines) > 1 {
		if content.Len() > 0 {
			content.WriteByte('\n')
		}
		content.WriteString(strings.Join(lines[1:], "\n"))
	}

	return flags, strings.TrimSpace(content.String()), nil
}

func (bm *BroadcastManager) start(
	ctx context.Context,
	m, progressMsg *tg.NewMessage,
	flags *BroadcastFlags,
	content string,
	chats, users []int64,
	stats *BroadcastStats,
) {
	defer bm.Stop()
	defer func() {
		if r := recover(); r != nil {
			gologging.ErrorF("Broadcast panic recovered: %v", r)
			bm.finalize(progressMsg, stats)
		}
	}()

	messagesSent := 0

	// Broadcast to chats
	for _, chatID := range chats {
		select {
		case <-ctx.Done():
			bm.finalize(progressMsg, stats)
			return
		default:
		}

		err := bm.sendMessage(ctx, m, chatID, content, flags)

		stats.mu.Lock()
		stats.DoneChats++
		stats.LastUpdate = time.Now()
		if err != nil {
			stats.FailedChats = append(stats.FailedChats, chatID)
			fmt.Fprintf(&stats.errorLog, "[%d] - [%v]\n", chatID, err)
		}
		stats.mu.Unlock()

		messagesSent++
		if !bm.handleDelay(ctx, messagesSent, flags.Delay) {
			bm.finalize(progressMsg, stats)
			return
		}
	}

	// Broadcast to users
	for _, userID := range users {
		select {
		case <-ctx.Done():
			bm.finalize(progressMsg, stats)
			return
		default:
		}

		err := bm.sendMessage(ctx, m, userID, content, flags)

		stats.mu.Lock()
		stats.DoneUsers++
		stats.LastUpdate = time.Now()
		if err != nil {
			stats.FailedUsers = append(stats.FailedUsers, userID)
			fmt.Fprintf(&stats.errorLog, "[%d] - [%v]\n", userID, err)
		}
		stats.mu.Unlock()

		messagesSent++
		if !bm.handleDelay(ctx, messagesSent, flags.Delay) {
			bm.finalize(progressMsg, stats)
			return
		}
	}

	bm.finalize(progressMsg, stats)
}

func (bm *BroadcastManager) sendMessage(
	ctx context.Context,
	m *tg.NewMessage,
	targetID int64,
	content string,
	flags *BroadcastFlags,
) error {
	var (
		sentMsg *tg.NewMessage
		err     error
	)

	try := func() error {
		if m.IsReply() {
			fOpts := &tg.ForwardOptions{}
			if flags.Copy {
				fOpts.HideAuthor = true
			}
			fMsgs, ferr := m.Client.Forward(
				targetID,
				m.Peer,
				[]int32{m.ReplyID()},
				fOpts,
			)
			if ferr != nil {
				return ferr
			}
			if len(fMsgs) > 0 {
				sentMsg = &fMsgs[0]
			}
			return nil
		}

		sent, ferr := m.Client.SendMessage(targetID, content)
		if ferr != nil {
			return ferr
		}
		sentMsg = sent
		return nil
	}

	maxAttempts := 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err = try()
		if err == nil {
			break
		}

		if wait := tg.GetFloodWait(err); wait > 0 {
			gologging.ErrorF(
				"FloodWait detected (%ds). Retrying (attempt %d).",
				wait,
				attempt,
			)
			if !bm.sleepCtx(ctx, time.Duration(wait)*time.Second) {
				return context.Canceled
			}
			continue
		} else {
			break
		}
	}

	if err != nil {
		if !tg.MatchError(err, "USER_IS_BLOCKED") &&
			!tg.MatchError(err, "CHAT_WRITE_FORBIDDEN") &&
			!tg.MatchError(err, "USER_IS_DEACTIVATED") {
			gologging.ErrorF("Broadcast failed for %d: %v", targetID, err)
		}
		return err
	}

	if sentMsg != nil && (flags.Pin || flags.PinLoud) {
		if _, perr := m.Client.PinMessage(targetID, sentMsg.ID, &tg.PinOptions{Silent: !flags.PinLoud}); perr != nil {
			gologging.ErrorF("Pin failed for %d: %v", targetID, perr)
		}
	}
	return nil
}

func (bm *BroadcastManager) handleDelay(
	ctx context.Context,
	count int,
	baseDelay float64,
) bool {
	if count%30 == 0 {
		return bm.sleepCtx(ctx, 7500*time.Millisecond)
	}
	return bm.sleepCtx(ctx, time.Duration(baseDelay*float64(time.Second)))
}

func (bm *BroadcastManager) updateProgress(
	ctx context.Context,
	progressMsg *tg.NewMessage,
	stats *BroadcastStats,
) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats.mu.Lock()
			if stats.Finished {
				stats.mu.Unlock()
				return
			}
			text := formatBroadcastProgress(
				stats,
				false,
				progressMsg.ChannelID(),
			)
			stats.mu.Unlock()

			_, _ = progressMsg.Edit(text, &tg.SendOptions{
				ReplyMarkup: core.GetBroadcastCancelKeyboard(
					progressMsg.ChannelID(),
				),
			})
		}
	}
}

func (bm *BroadcastManager) finalize(
	progressMsg *tg.NewMessage,
	stats *BroadcastStats,
) {
	stats.mu.Lock()
	stats.Finished = true
	text := formatBroadcastProgress(stats, true, progressMsg.ChannelID())
	errs := stats.errorLog.String()
	stats.mu.Unlock()

	_, _ = progressMsg.Edit(text)

	if errs != "" {
		tmpFile := "broadcast_errors.txt"
		if err := os.WriteFile(tmpFile, []byte(errs), 0o600); err != nil {
			gologging.ErrorF("Failed to write broadcast errors: %v", err)
			return
		}
		defer os.Remove(tmpFile)

		_, _ = progressMsg.ReplyMedia(tmpFile)
	}
}

func formatBroadcastProgress(
	stats *BroadcastStats,
	final bool,
	chatID int64,
) string {
	elapsed := time.Since(stats.StartTime)

	var sb strings.Builder

	if !final {
		sb.WriteString(F(chatID, "broadcast_progress_header") + "\n\n")
	}

	chatProgress := 0.0
	if stats.TotalChats > 0 {
		chatProgress = float64(
			stats.DoneChats,
		) / float64(
			stats.TotalChats,
		) * 100
	}

	userProgress := 0.0
	if stats.TotalUsers > 0 {
		userProgress = float64(
			stats.DoneUsers,
		) / float64(
			stats.TotalUsers,
		) * 100
	}

	sb.WriteString(F(chatID, "broadcast_total_chats", locales.Arg{
		"done":     stats.DoneChats,
		"total":    stats.TotalChats,
		"progress": fmt.Sprintf("%.1f", chatProgress),
	}) + "\n")

	sb.WriteString(F(chatID, "broadcast_total_users", locales.Arg{
		"done":     stats.DoneUsers,
		"total":    stats.TotalUsers,
		"progress": fmt.Sprintf("%.1f", userProgress),
	}) + "\n\n")

	if len(stats.FailedChats) > 0 {
		sb.WriteString(F(chatID, "broadcast_failed_chats", locales.Arg{
			"count": len(stats.FailedChats),
		}) + "\n")
	}
	if len(stats.FailedUsers) > 0 {
		sb.WriteString(F(chatID, "broadcast_failed_users", locales.Arg{
			"count": len(stats.FailedUsers),
		}) + "\n")
	}

	if len(stats.FailedChats) > 0 || len(stats.FailedUsers) > 0 {
		sb.WriteString("\n")
	}

	totalDone := stats.DoneChats + stats.DoneUsers
	totalTargets := stats.TotalChats + stats.TotalUsers

	avgSpeed := 0.0
	if elapsed.Seconds() > 0 && totalDone > 0 {
		avgSpeed = float64(totalDone) / elapsed.Seconds()
	}

	sb.WriteString(F(chatID, "broadcast_delay", locales.Arg{
		"delay": fmt.Sprintf("%.1f", stats.Delay),
	}) + "\n")

	sb.WriteString(F(chatID, "broadcast_elapsed", locales.Arg{
		"elapsed": utils.FormatDuration(int(elapsed.Seconds())),
	}) + "\n")

	if !final && avgSpeed > 0 && totalDone < totalTargets {
		remaining := totalTargets - totalDone
		etaSeconds := float64(remaining) / avgSpeed
		sb.WriteString("\n" + F(chatID, "broadcast_eta", locales.Arg{
			"eta": utils.FormatDuration(int(etaSeconds)),
		}))
	}

	if final {
		totalSent := stats.DoneChats + stats.DoneUsers
		totalFailed := len(stats.FailedChats) + len(stats.FailedUsers)

		successRate := 0.0
		if totalTargets > 0 {
			successRate = float64(
				totalSent-totalFailed,
			) / float64(
				totalTargets,
			) * 100
		}

		sb.WriteString("\n\n" + F(chatID, "broadcast_success_rate", locales.Arg{
			"rate":  fmt.Sprintf("%.1f", successRate),
			"sent":  totalSent - totalFailed,
			"total": totalTargets,
		}))
	}

	return sb.String()
}

func handleBroadcastCancel(m *tg.NewMessage) error {
	if !bManager.IsActive() {
		m.Reply(F(m.ChannelID(), "broadcast_not_running"))
		return tg.ErrEndGroup
	}

	bManager.Stop()

	m.Reply(F(m.ChannelID(), "broadcast_cancel_success"))
	return tg.ErrEndGroup
}

func broadcastCancelCB(cb *tg.CallbackQuery) error {
	if cb.SenderID != config.OwnerID {
		cb.Answer(
			F(cb.ChannelID(), "broadcast_cancel_owner_only"),
			&tg.CallbackOptions{Alert: true},
		)
		return tg.ErrEndGroup
	}

	if !bManager.IsActive() {
		cb.Answer(
			F(cb.ChannelID(), "broadcast_cancel_none_running"),
			&tg.CallbackOptions{Alert: true},
		)
		return tg.ErrEndGroup
	}

	bManager.Stop()

	cb.Answer(
		F(cb.ChannelID(), "broadcast_cancel_done"),
		&tg.CallbackOptions{Alert: true},
	)
	cb.Edit(F(cb.ChannelID(), "broadcast_cancel_done"))
	return tg.ErrEndGroup
}

func (bm *BroadcastManager) sleepCtx(
	ctx context.Context,
	d time.Duration,
) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}
