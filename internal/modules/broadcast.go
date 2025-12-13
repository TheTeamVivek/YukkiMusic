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
	"context"
	"fmt"
	"html"
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
)

var (
	broadcastMu     sync.Mutex
	broadcastActive bool
	broadcastCancel context.CancelFunc
	broadcastCtx    context.Context

	defaultDelay = 1.5
)

type BroadcastStats struct {
	TotalChats  int
	TotalUsers  int
	DoneChats   int
	DoneUsers   int
	FailedChats []int64
	FailedUsers []int64
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
	broadcastMu.Lock()
	if broadcastActive {
		broadcastMu.Unlock()
		m.Reply(F(chatID, "broadcast_already_running"))
		return tg.ErrEndGroup
	}
	broadcastActive = true
	broadcastMu.Unlock()

	// Parse flags and content
	flags, content, err := parseBroadcastCommand(m)
	if err != nil {
		broadcastMu.Lock()
		broadcastActive = false
		broadcastMu.Unlock()
		m.Reply(F(chatID, "broadcast_parse_failed", locales.Arg{
			"error": html.EscapeString(err.Error()),
		}))
		return tg.ErrEndGroup
	}

	if content == "" && !m.IsReply() {
		broadcastMu.Lock()
		broadcastActive = false
		broadcastMu.Unlock()
		m.Reply(F(chatID, "broadcast_no_content", locales.Arg{
			"cmd": getCommand(m),
		}))
		return tg.ErrEndGroup
	}

	// Get served chats and users
	var servedChats, servedUsers []int64
	var servedChatErr, servedUserErr error

	if !flags.NoChat {
		servedChats, servedChatErr = database.GetServed()
		if servedChatErr != nil {
			broadcastMu.Lock()
			broadcastActive = false
			broadcastMu.Unlock()
			m.Reply(F(chatID, "broadcast_fetch_chats_failed", locales.Arg{
				"error": html.EscapeString(servedChatErr.Error()),
			}))

			return tg.ErrEndGroup
		}
	}

	if !flags.NoUser {
		servedUsers, servedUserErr = database.GetServed(true)
		if servedUserErr != nil {
			broadcastMu.Lock()
			broadcastActive = false
			broadcastMu.Unlock()
			m.Reply(F(chatID, "broadcast_fetch_users_failed", locales.Arg{
				"error": html.EscapeString(servedUserErr.Error()),
			}))
			return tg.ErrEndGroup
		}
	}

	// Check if there are any targets
	if len(servedChats) == 0 && len(servedUsers) == 0 {
		broadcastMu.Lock()
		broadcastActive = false
		broadcastMu.Unlock()
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

	broadcastCtx, broadcastCancel = context.WithCancel(context.Background())

	progressMsg, err := m.Reply(F(chatID, "broadcast_initializing"),
		&tg.SendOptions{
			ReplyMarkup: core.GetBroadcastCancelKeyboard(chatID),
		})
	if err != nil {
		broadcastMu.Lock()
		broadcastActive = false
		broadcastCtx = nil
		broadcastCancel = nil
		broadcastMu.Unlock()
		gologging.ErrorF("Failed to send broadcast progress message: %v", err)
		return tg.ErrEndGroup
	}

	// Start progress updater in goroutine
	go updateBroadcastProgress(broadcastCtx, progressMsg, stats)

	// Start broadcast in goroutine
	go func() {
		defer func() {
			broadcastMu.Lock()
			if broadcastCancel != nil {
				broadcastCancel()
			}
			broadcastActive = false
			broadcastCtx = nil
			broadcastCancel = nil
			broadcastMu.Unlock()
		}()

		startBroadcast(broadcastCtx, m, progressMsg, flags, content, servedChats, servedUsers, stats)
	}()

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

	firstLine := lines[0]
	words := strings.Fields(firstLine)

	var contentWords []string
	skipNext := false

	for i := 0; i < len(words); i++ {
		if skipNext {
			skipNext = false
			continue
		}

		word := strings.ToLower(words[i])

		switch word {
		case "-nochat", "--nochat":
			flags.NoChat = true
		case "-nouser", "--nouser":
			flags.NoUser = true
		case "-copy", "--copy":
			flags.Copy = true
		case "-pin", "--pin":
			flags.Pin = true
		case "-pinloud", "--pinloud":
			flags.PinLoud = true
		case "-limit", "--limit":
			if i+1 < len(words) {
				limit, err := strconv.Atoi(words[i+1])
				if err != nil {
					return nil, "", fmt.Errorf("invalid limit value: %s", words[i+1])
				}
				if limit < 0 {
					return nil, "", fmt.Errorf("limit must be non-negative")
				}
				flags.Limit = limit
				skipNext = true
			} else {
				return nil, "", fmt.Errorf("-limit requires a value")
			}
		case "-delay", "--delay":
			if i+1 < len(words) {
				delay, err := strconv.ParseFloat(words[i+1], 64)
				if err != nil {
					return nil, "", fmt.Errorf("invalid delay value: %s", words[i+1])
				}
				if delay < 0 {
					return nil, "", fmt.Errorf("delay must be non-negative")
				}
				flags.Delay = delay
				skipNext = true
			} else {
				return nil, "", fmt.Errorf("-delay requires a value")
			}
		case "-cancel", "--cancel":
			// Special case, handled in main handler
			continue
		default:
			// Not a flag, add to content (preserve original case from words slice)
			contentWords = append(contentWords, words[i])
		}
	}

	firstLineContent := strings.Join(contentWords, " ")

	var content string
	if firstLineContent != "" {
		content = firstLineContent
		if len(lines) > 1 {
			content += "\n" + strings.Join(lines[1:], "\n")
		}
	} else if len(lines) > 1 {
		content = strings.Join(lines[1:], "\n")
	}

	content = strings.TrimSpace(content)

	return flags, content, nil
}

func startBroadcast(
	ctx context.Context,
	m, progressMsg *tg.NewMessage,
	flags *BroadcastFlags,
	content string,
	chats, users []int64,
	stats *BroadcastStats,
) {
	defer func() {
		if r := recover(); r != nil {
			gologging.ErrorF("Broadcast panic recovered: %v", r)
			finalizeBroadcast(progressMsg, stats, true)
		}
	}()

	messagesSent := 0

	// Broadcast to chats
	for _, chatID := range chats {
		select {
		case <-ctx.Done():
			finalizeBroadcast(progressMsg, stats, true)
			return
		default:
		}

		success := sendBroadcastMessage(ctx, m, chatID, content, flags)

		stats.mu.Lock()
		stats.DoneChats++
		stats.LastUpdate = time.Now()
		if !success {
			stats.FailedChats = append(stats.FailedChats, chatID)
		}
		stats.mu.Unlock()

		messagesSent++
		if !handleBroadcastDelay(ctx, messagesSent, flags.Delay) {
			finalizeBroadcast(progressMsg, stats, true)
			return
		}
	}

	// Broadcast to users
	for _, userID := range users {
		select {
		case <-ctx.Done():
			finalizeBroadcast(progressMsg, stats, true)
			return
		default:
		}

		success := sendBroadcastMessage(ctx, m, userID, content, flags)

		stats.mu.Lock()
		stats.DoneUsers++
		stats.LastUpdate = time.Now()
		if !success {
			stats.FailedUsers = append(stats.FailedUsers, userID)
		}
		stats.mu.Unlock()

		messagesSent++
		if !handleBroadcastDelay(ctx, messagesSent, flags.Delay) {
			finalizeBroadcast(progressMsg, stats, true)
			return
		}
	}

	finalizeBroadcast(progressMsg, stats, false)
}

func sendBroadcastMessage(ctx context.Context, m *tg.NewMessage, targetID int64, content string, flags *BroadcastFlags) bool {
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
			fMsgs, ferr := m.Client.Forward(targetID, m.Peer, []int32{m.ReplyID()}, fOpts)
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
			gologging.ErrorF("FloodWait detected (%ds). Retrying (attempt %d).", wait, attempt)
			if !sleepCtx(ctx, time.Duration(wait)*time.Second) {
				return false
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
		return false
	}

	if sentMsg != nil && (flags.Pin || flags.PinLoud) {
		if _, perr := m.Client.PinMessage(targetID, sentMsg.ID, &tg.PinOptions{Silent: !flags.PinLoud}); perr != nil {
			gologging.ErrorF("Pin failed for %d: %v", targetID, perr)
		}
	}
	return true
}

func handleBroadcastDelay(ctx context.Context, count int, baseDelay float64) bool {
	if count%30 == 0 {
		return sleepCtx(ctx, 7500*time.Millisecond)
	}
	return sleepCtx(ctx, time.Duration(baseDelay*float64(time.Second)))
}

func updateBroadcastProgress(ctx context.Context, progressMsg *tg.NewMessage, stats *BroadcastStats) {
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
			text := formatBroadcastProgress(stats, false, progressMsg.ChannelID())
			stats.mu.Unlock()

			progressMsg.Edit(text, &tg.SendOptions{
				ReplyMarkup: core.GetBroadcastCancelKeyboard(progressMsg.ChannelID()),
			})
		}
	}
}

func finalizeBroadcast(progressMsg *tg.NewMessage, stats *BroadcastStats, cancelled bool) {
	stats.mu.Lock()
	stats.Finished = true
	text := formatBroadcastProgress(stats, true, progressMsg.ChannelID())
	stats.mu.Unlock()

	if cancelled {
		text = F(progressMsg.ChannelID(), "broadcast_cancelled_header")
	} else {
		text = F(progressMsg.ChannelID(), "broadcast_completed_header")
	}

	progressMsg.Edit(text)
}

func formatBroadcastProgress(stats *BroadcastStats, final bool, chatID int64) string {
	elapsed := time.Since(stats.StartTime)

	var sb strings.Builder

	if !final {
		sb.WriteString(F(chatID, "broadcast_progress_header") + "\n\n")
	}

	chatProgress := 0.0
	if stats.TotalChats > 0 {
		chatProgress = float64(stats.DoneChats) / float64(stats.TotalChats) * 100
	}

	userProgress := 0.0
	if stats.TotalUsers > 0 {
		userProgress = float64(stats.DoneUsers) / float64(stats.TotalUsers) * 100
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
		"elapsed": formatDuration(int(elapsed.Seconds())),
	}) + "\n")

	if !final && avgSpeed > 0 && totalDone < totalTargets {
		remaining := totalTargets - totalDone
		etaSeconds := float64(remaining) / avgSpeed
		sb.WriteString("\n" + F(chatID, "broadcast_eta", locales.Arg{
			"eta": formatDuration(int(etaSeconds)),
		}))
	}

	if final {
		totalSent := stats.DoneChats + stats.DoneUsers
		totalFailed := len(stats.FailedChats) + len(stats.FailedUsers)

		successRate := 0.0
		if totalTargets > 0 {
			successRate = float64(totalSent-totalFailed) / float64(totalTargets) * 100
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
	broadcastMu.Lock()
	defer broadcastMu.Unlock()

	if !broadcastActive {
		m.Reply(F(m.ChannelID(), "broadcast_not_running"))
		return tg.ErrEndGroup
	}

	if broadcastCancel != nil {
		broadcastCancel()
	}
	broadcastActive = false
	broadcastCtx = nil
	broadcastCancel = nil

	m.Reply(F(m.ChannelID(), "broadcast_cancel_success"))
	return tg.ErrEndGroup
}

func broadcastCancelCB(cb *tg.CallbackQuery) error {
	if cb.SenderID != config.OwnerID {
		cb.Answer(F(cb.ChannelID(), "broadcast_cancel_owner_only"), &tg.CallbackOptions{Alert: true})
		return tg.ErrEndGroup
	}

	broadcastMu.Lock()
	defer broadcastMu.Unlock()

	if !broadcastActive {
		cb.Answer(F(cb.ChannelID(), "broadcast_cancel_none_running"), &tg.CallbackOptions{Alert: true})
		return tg.ErrEndGroup
	}

	if broadcastCancel != nil {
		broadcastCancel()
	}

	broadcastActive = false
	broadcastCtx = nil
	broadcastCancel = nil
	cb.Answer(F(cb.ChannelID(), "broadcast_cancel_done"), &tg.CallbackOptions{Alert: true})
	cb.Edit(F(cb.ChannelID(), "broadcast_cancel_done"))
	return tg.ErrEndGroup
}

func sleepCtx(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}
