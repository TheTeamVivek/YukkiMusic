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
	remoteID string             // Telegram remote file id, for updateFile routing
	fileID   int32              // TDLib file id, set once updateFile events arrive
	started  time.Time
	lastEdit time.Time
	lastSize int64
	total    int64
}

// Downloads are keyed by chat ID (for the cancel button) and by Telegram remote
// file ID (because TDLib updateFile events only carry file ids, not chats).
var (
	downloadMu     sync.RWMutex
	downloadByChat = make(map[int64]*downloadSession)
	downloadByFile = make(map[string]int64) // telegram remote file id -> chat id
)

func init() {
	platforms.OnDownloadStart = attachDownload
}

// beginDownload registers a new download session for a chat.
func beginDownload(chatID int64, cancel context.CancelFunc, msg *td.Message) {
	downloadMu.Lock()
	defer downloadMu.Unlock()
	downloadByChat[chatID] = &downloadSession{
		cancel:  cancel,
		msg:     msg,
		started: time.Now(),
	}
}

// attachDownload links a Telegram remote file id to the chat that is downloading
// it. It runs right before a TDLib download starts, via platforms.OnDownloadStart.
func attachDownload(fileID string, msg *td.Message) {
	if fileID == "" || msg == nil {
		return
	}

	downloadMu.Lock()
	defer downloadMu.Unlock()

	s, ok := downloadByChat[msg.ChatId]
	if !ok {
		s = &downloadSession{msg: msg, started: time.Now()}
		downloadByChat[msg.ChatId] = s
	}
	if s.remoteID != "" && s.remoteID != fileID {
		delete(downloadByFile, s.remoteID)
	}
	s.remoteID = fileID
	downloadByFile[fileID] = msg.ChatId
}

// cancelDownload stops the download running in a chat. It cancels the Go context
// (yt-dlp/http) and, for Telegram file downloads, also asks TDLib to stop.
// It reports whether a download was in progress for the chat.
func cancelDownload(chatID int64) bool {
	downloadMu.Lock()
	s, ok := downloadByChat[chatID]
	if !ok {
		downloadMu.Unlock()
		return false
	}
	removeDownloadLocked(chatID)
	fileID := s.fileID
	if s.cancel != nil {
		s.cancel()
	}
	downloadMu.Unlock()

	if fileID != 0 {
		if err := core.Bot.CancelDownloadFile(fileID, &td.CancelDownloadFileOpts{}); err != nil {
			utils.Log.Warnf("Failed to cancel download for chat %d: %v", chatID, err)
		}
	}
	return true
}

// finishDownload drops the session once a download completes (success or failure).
func finishDownload(chatID int64) {
	downloadMu.Lock()
	defer downloadMu.Unlock()
	removeDownloadLocked(chatID)
}

func removeDownloadLocked(chatID int64) {
	if s, ok := downloadByChat[chatID]; ok && s.remoteID != "" {
		delete(downloadByFile, s.remoteID)
	}
	delete(downloadByChat, chatID)
}

// downloadUpdateHandler records the TDLib file id (needed for cancellation) and
// refreshes the "Downloading..." progress message. It is registered to run on
// every TDLib file update.
func downloadUpdateHandler(c *td.Client, u *td.UpdateFile) error {
	if u == nil || u.File == nil || u.File.Remote == nil || u.File.Local == nil {
		return nil
	}

	downloadMu.Lock()
	chatID, ok := downloadByFile[u.File.Remote.Id]
	if !ok {
		downloadMu.Unlock()
		return nil
	}
	s := downloadByChat[chatID]
	if s == nil {
		downloadMu.Unlock()
		return nil
	}
	s.fileID = u.File.Id

	now := time.Now()
	done := !u.File.Local.IsDownloadingActive
	if now.Sub(s.lastEdit) < downloadEditInterval && !done {
		downloadMu.Unlock()
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
		removeDownloadLocked(chatID)
	}
	downloadMu.Unlock()

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

	if _, err := msg.EditText(c, text, &td.EditMessageTextOpts{
		ParseMode:   td.ParseModeHTML,
		ReplyMarkup: msg.ReplyMarkup,
	}); err != nil {
		// Ignore edit errors; the next update will retry.
		return nil
	}
	return nil
}
