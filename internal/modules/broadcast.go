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
	"main/internal/database"
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
	if strings.Contains(text, "-cancel") || strings.Contains(text, "--cancel") {
		return handleBroadcastCancel(m)
	}

	// Check if broadcast is already running
	broadcastMu.Lock()
	if broadcastActive {
		broadcastMu.Unlock()
		m.Reply("⚠️ <b>A broadcast is already running.</b>\n\n" +
			"Please wait for it to complete or cancel it first using:\n" +
			"• <code>/broadcast -cancel</code>\n" +
			"• Or click the Cancel button on the progress message")
		return tg.EndGroup
	}
	broadcastActive = true
	broadcastMu.Unlock()

	// Parse flags and content
	flags, content, err := parseBroadcastCommand(m)
	if err != nil {
		broadcastMu.Lock()
		broadcastActive = false
		broadcastMu.Unlock()
		m.Reply(fmt.Sprintf("❌ <b>Failed to parse broadcast command:</b>\n<code>%s</code>",
			html.EscapeString(err.Error())))
		return tg.EndGroup
	}

	if content == "" && !m.IsReply() {
		broadcastMu.Lock()
		broadcastActive = false
		broadcastMu.Unlock()
		m.Reply(fmt.Sprintf("⚠️ <b>No content provided for broadcast.</b>\n\n"+
			"<b>Usage:</b> <code>%s [text or reply to message]</code>\n\n"+
			"<b>Example:</b>\n<code>%s Hello everyone!</code>",
			getCommand(m), getCommand(m)))
		return tg.EndGroup
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
			m.Reply(fmt.Sprintf("❌ <b>Failed to fetch served chats:</b>\n<code>%s</code>",
				html.EscapeString(servedChatErr.Error())))
			return tg.EndGroup
		}
	}

	if !flags.NoUser {
		servedUsers, servedUserErr = database.GetServed(true)
		if servedUserErr != nil {
			broadcastMu.Lock()
			broadcastActive = false
			broadcastMu.Unlock()
			m.Reply(fmt.Sprintf("❌ <b>Failed to fetch served users:</b>\n<code>%s</code>",
				html.EscapeString(servedUserErr.Error())))
			return tg.EndGroup
		}
	}

	// Check if there are any targets
	if len(servedChats) == 0 && len(servedUsers) == 0 {
		broadcastMu.Lock()
		broadcastActive = false
		broadcastMu.Unlock()
		m.Reply("⚠️ <b>No targets found for broadcast.</b>\n\n" +
			"Make sure you haven't excluded all targets with flags.")
		return tg.EndGroup
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

	progressMsg, err := m.Reply("📡 <b>Initializing broadcast...</b>\n\nPlease wait...",
		&tg.SendOptions{
			ReplyMarkup: getBroadcastCancelKeyboard(),
		})
	if err != nil {
		broadcastMu.Lock()
		broadcastActive = false
		broadcastCtx = nil
		broadcastCancel = nil
		broadcastMu.Unlock()
		gologging.ErrorF("Failed to send broadcast progress message: %v", err)
		return tg.EndGroup
	}

	// Start progress updater in goroutine
	go updateBroadcastProgress(broadcastCtx, progressMsg, stats)

	// Start broadcast in goroutine
	go func() {
		defer func() {
			broadcastMu.Lock()
			broadcastActive = false
			broadcastCtx = nil
			broadcastCancel = nil
			broadcastMu.Unlock()
		}()

		startBroadcast(broadcastCtx, m, progressMsg, flags, content, servedChats, servedUsers, stats)
	}()

	return tg.EndGroup
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
			text := formatBroadcastProgress(stats, false)
			stats.mu.Unlock()

			progressMsg.Edit(text, &tg.SendOptions{
				ReplyMarkup: getBroadcastCancelKeyboard(),
			})
		}
	}
}

func finalizeBroadcast(progressMsg *tg.NewMessage, stats *BroadcastStats, cancelled bool) {
	stats.mu.Lock()
	text := formatBroadcastProgress(stats, true)
	stats.mu.Unlock()

	if cancelled {
		text = "🚫 <b>Broadcast Cancelled</b>\n\n" + text
	} else {
		text = "✅ <b>Broadcast Completed</b>\n\n" + text
	}

	progressMsg.Edit(text)
}

func formatBroadcastProgress(stats *BroadcastStats, final bool) string {
	elapsed := time.Since(stats.StartTime)

	var sb strings.Builder

	if !final {
		sb.WriteString("📡 <b>Broadcasting...</b>\n\n")
	}

	chatProgress := 0.0
	if stats.TotalChats > 0 {
		chatProgress = float64(stats.DoneChats) / float64(stats.TotalChats) * 100
	}

	userProgress := 0.0
	if stats.TotalUsers > 0 {
		userProgress = float64(stats.DoneUsers) / float64(stats.TotalUsers) * 100
	}

	sb.WriteString(fmt.Sprintf("📊 <b>Total Chats:</b> %d/%d (%.1f%%)\n",
		stats.DoneChats, stats.TotalChats, chatProgress))
	sb.WriteString(fmt.Sprintf("👥 <b>Total Users:</b> %d/%d (%.1f%%)\n\n",
		stats.DoneUsers, stats.TotalUsers, userProgress))

	if len(stats.FailedChats) > 0 {
		sb.WriteString(fmt.Sprintf("❌ <b>Failed Chats:</b> %d\n", len(stats.FailedChats)))
	}
	if len(stats.FailedUsers) > 0 {
		sb.WriteString(fmt.Sprintf("❌ <b>Failed Users:</b> %d\n", len(stats.FailedUsers)))
	}

	if len(stats.FailedChats) > 0 || len(stats.FailedUsers) > 0 {
		sb.WriteString("\n")
	}

	// Calculate speed and metrics
	totalDone := stats.DoneChats + stats.DoneUsers
	totalTargets := stats.TotalChats + stats.TotalUsers

	avgSpeed := 0.0
	if elapsed.Seconds() > 0 && totalDone > 0 {
		avgSpeed = float64(totalDone) / elapsed.Seconds()
	}

	sb.WriteString(fmt.Sprintf("⏱ <b>Delay:</b> %.1fs\n", stats.Delay))
	sb.WriteString(fmt.Sprintf("⏰ <b>Elapsed:</b> %s\n", formatDuration(int(elapsed.Seconds()))))
	sb.WriteString(fmt.Sprintf("🚀 <b>Avg Speed:</b> %.2f msg/s", avgSpeed))

	// Calculate ETA for non-final broadcasts
	if !final && avgSpeed > 0 && totalDone < totalTargets {
		remaining := totalTargets - totalDone
		etaSeconds := float64(remaining) / avgSpeed
		sb.WriteString(fmt.Sprintf("\n⏳ <b>ETA:</b> %s", formatDuration(int(etaSeconds))))
	}

	if final {
		totalSent := stats.DoneChats + stats.DoneUsers
		totalFailed := len(stats.FailedChats) + len(stats.FailedUsers)

		successRate := 0.0
		if totalTargets > 0 {
			successRate = float64(totalSent-totalFailed) / float64(totalTargets) * 100
		}

		sb.WriteString(fmt.Sprintf("\n\n✨ <b>Success Rate:</b> %.1f%% (%d/%d)",
			successRate, totalSent-totalFailed, totalTargets))
	}

	return sb.String()
}

func getBroadcastCancelKeyboard() tg.ReplyMarkup {
	kb := tg.NewKeyboard()
	kb.AddRow(tg.Button.Data("🚫 Cancel Broadcast", "broadcast:cancel"))
	return kb.Build()
}

func handleBroadcastCancel(m *tg.NewMessage) error {
	broadcastMu.Lock()
	defer broadcastMu.Unlock()

	if !broadcastActive {
		m.Reply("ℹ️ <b>No broadcast is currently running.</b>")
		return tg.EndGroup
	}

	if broadcastCancel != nil {
		broadcastCancel()
	}
	broadcastActive = false
	broadcastCtx = nil
	broadcastCancel = nil

	m.Reply("🚫 <b>Broadcast cancelled successfully.</b>")
	return tg.EndGroup
}

func broadcastCancelCB(cb *tg.CallbackQuery) error {
	if cb.SenderID != config.OwnerID {
		cb.Answer("⚠️ Only the owner can cancel broadcasts.", &tg.CallbackOptions{Alert: true})
		return tg.EndGroup
	}

	broadcastMu.Lock()
	defer broadcastMu.Unlock()

	if !broadcastActive {
		cb.Answer("ℹ️ No broadcast is currently running.", &tg.CallbackOptions{Alert: true})
		return tg.EndGroup
	}

	if broadcastCancel != nil {
		broadcastCancel()
	}

	broadcastActive = false
	broadcastCtx = nil
	broadcastCancel = nil

	cb.Answer("🚫 Broadcast cancelled.", &tg.CallbackOptions{Alert: true})
	return tg.EndGroup
}

func sleepCtx(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}
