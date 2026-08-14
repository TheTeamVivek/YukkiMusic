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
	"sync"
	"time"

	"yukkimusic/internal/logger"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/core"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/platforms"
	"yukkimusic/internal/utils"
)

// downloadEditInterval throttles how often the progress message is edited.
const downloadEditInterval = 2 * time.Second

// downloadSession tracks one in-flight track download for a chat.
// A chat only downloads a single track at a time.
type downloadSession struct {
	cancel   context.CancelFunc // cancels the Go-side download (yt-dlp/http)
	msg      *td.Message        // status message edited while downloading
	started  time.Time
	lastEdit time.Time
	lastSize int64
	total    int64
	fileID   int32 // TDLib file id, set once updateFile events arrive
}

// downloadManager coordinates cancellation and progress for downloads. It is
// keyed by chat ID (for the cancel button) and by the Telegram remote file ID
// (because TDLib updateFile events only carry file ids, not chats).
type downloadManager struct {
	mu     sync.RWMutex
	byChat map[int64]*downloadSession
	byFile map[string]int64 // telegram remote file id -> chat id
}

var downloads = &downloadManager{
	byChat: make(map[int64]*downloadSession),
	byFile: make(map[string]int64),
}

func init() {
	platforms.OnDownloadStart = func(fileID string, msg *td.Message) {
		downloads.attach(fileID, msg)
	}
}

// begin registers a new download session for a chat.
func (dm *downloadManager) begin(chatID int64, cancel context.CancelFunc, msg *td.Message) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.byChat[chatID] = &downloadSession{
		cancel:  cancel,
		msg:     msg,
		started: time.Now(),
	}
}

// attach links a Telegram remote file id to the chat that is downloading it.
// It is called by the platforms package right before a TDLib download starts.
func (dm *downloadManager) attach(fileID string, msg *td.Message) {
	if fileID == "" || msg == nil {
		return
	}
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if _, ok := dm.byChat[msg.ChatId]; !ok {
		dm.byChat[msg.ChatId] = &downloadSession{msg: msg, started: time.Now()}
	}
	dm.byFile[fileID] = msg.ChatId
}

// cancel stops the download running in a chat. It cancels the Go context
// (yt-dlp/http) and, for Telegram file downloads, also asks TDLib to stop.
// It returns false when no download is in progress for the chat.
func (dm *downloadManager) cancel(chatID int64) bool {
	dm.mu.Lock()
	s, ok := dm.byChat[chatID]
	if !ok {
		dm.mu.Unlock()
		return false
	}
	dm.removeLocked(chatID)
	fileID := s.fileID
	if s.cancel != nil {
		s.cancel()
	}
	dm.mu.Unlock()

	if fileID != 0 {
		if err := core.Bot.CancelDownloadFile(fileID, &td.CancelDownloadFileOpts{}); err != nil {
			logger.Warnf("Failed to cancel download for chat %d: %v", chatID, err)
		}
	}
	return true
}

// finish drops the session once a download completes (success or failure).
func (dm *downloadManager) finish(chatID int64) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.removeLocked(chatID)
}

func (dm *downloadManager) removeLocked(chatID int64) {
	delete(dm.byChat, chatID)
	for id, c := range dm.byFile {
		if c == chatID {
			delete(dm.byFile, id)
		}
	}
}

// onFileUpdate records the TDLib file id (needed for cancellation) and
// refreshes the "Downloading..." progress message.
func (dm *downloadManager) onFileUpdate(c *td.Client, u *td.UpdateFile) error {
	if u == nil || u.File == nil || u.File.Remote == nil || u.File.Local == nil {
		return nil
	}

	dm.mu.Lock()
	chatID, ok := dm.byFile[u.File.Remote.Id]
	if !ok {
		dm.mu.Unlock()
		return nil
	}
	s := dm.byChat[chatID]
	if s == nil {
		dm.mu.Unlock()
		return nil
	}
	s.fileID = u.File.Id

	now := time.Now()
	done := !u.File.Local.IsDownloadingActive
	if now.Sub(s.lastEdit) < downloadEditInterval && !done {
		dm.mu.Unlock()
		return nil
	}
	interval := now.Sub(s.lastEdit).Seconds()
	s.lastEdit = now

	total := u.File.Size
	if total <= 0 {
		total = u.File.ExpectedSize
	}
	if total > 0 {
		s.total = total
	}
	current := min(u.File.Local.DownloadedSize, s.total)

	speed := 0.0
	if interval > 0 && current >= s.lastSize {
		speed = float64(current-s.lastSize) / interval
	}
	s.lastSize = current

	percentage := 0.0
	if s.total > 0 {
		percentage = float64(current) / float64(s.total) * 100
	}
	eta := int64(0)
	if speed > 0 && s.total > current {
		eta = int64(float64(s.total-current) / speed)
	}

	msg := s.msg
	started := s.started
	if done {
		dm.removeLocked(chatID)
	}
	dm.mu.Unlock()

	if msg == nil {
		return nil
	}

	elapsed := now.Sub(started).Seconds()

	lang, err := database.Language(msg.ChatId)
	if err != nil {
		lang = "en"
	}

	text := locales.Get(lang, "download_progress", locales.Arg{
		"percentage": fmt.Sprintf("%.1f", percentage),
		"speed":      utils.FormatBytes(int64(speed)) + "/s",
		"eta":        utils.FormatDuration(int(eta)),
		"elapsed":    utils.FormatTime(int(elapsed)),
	})

	if _, err := msg.EditText(c, text, &td.EditTextMessageOpts{
		ParseMode:   td.ParseModeHTML,
		ReplyMarkup: msg.ReplyMarkup,
	}); err != nil {
		// Ignore edit errors; the next update will retry.
		return nil
	}
	return nil
}

// downloadUpdateHandler is registered in handlers.go and drives the progress
// message of the active download.
func downloadUpdateHandler(c *td.Client, u *td.UpdateFile) error {
	return downloads.onFileUpdate(c, u)
}
