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

package platforms

import (
	"fmt"
	"sync"
	"time"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

const downloadEditInterval = 2 * time.Second

type dlProgressEntry struct {
	client   *td.Client
	msg      *td.Message
	total    int64
	started  time.Time
	lastEdit time.Time
	lastSize int64
}

type downloadProgress struct {
	mu     sync.Mutex
	once   sync.Once
	active map[string]*dlProgressEntry
}

var downloadProg = &downloadProgress{
	active: make(map[string]*dlProgressEntry),
}

// Start begins tracking download progress for the given remote file ID and
// edits statusMsg with progress updates as gotdbot emits updateFile events.
func (dp *downloadProgress) Start(c *td.Client, fileID string, msg *td.Message) {
	if c == nil || fileID == "" || msg == nil {
		return
	}

	dp.once.Do(func() {
		c.OnUpdateFile(dp.handleUpdate, nil)
	})

	dp.mu.Lock()
	defer dp.mu.Unlock()
	dp.active[fileID] = &dlProgressEntry{
		client:  c,
		msg:     msg,
		started: time.Now(),
	}
}

// Stop stops tracking download progress for the given remote file ID.
func (dp *downloadProgress) Stop(fileID string) {
	if fileID == "" {
		return
	}

	dp.mu.Lock()
	defer dp.mu.Unlock()
	delete(dp.active, fileID)
}

func (dp *downloadProgress) handleUpdate(c *td.Client, u *td.UpdateFile) error {
	if u == nil || u.File == nil || u.File.Remote == nil || u.File.Local == nil {
		return nil
	}

	fileID := u.File.Remote.Id
	if fileID == "" {
		return nil
	}

	dp.mu.Lock()
	e, ok := dp.active[fileID]
	dp.mu.Unlock()
	if !ok {
		return nil
	}

	now := time.Now()
	if now.Sub(e.lastEdit) < downloadEditInterval &&
		!u.File.Local.IsDownloadingCompleted {
		return nil
	}

	total := u.File.Size
	if total <= 0 {
		total = u.File.ExpectedSize
	}
	if total > 0 {
		e.total = total
	}

	current := u.File.Local.DownloadedSize
	if e.total > 0 && current > e.total {
		current = e.total
	}

	interval := now.Sub(e.lastEdit).Seconds()
	e.lastEdit = now

	elapsed := now.Sub(e.started).Seconds()
	speed := float64(0)
	if interval > 0 && current >= e.lastSize {
		speed = float64(current-e.lastSize) / interval
	}
	e.lastSize = current

	percentage := 0.0
	if e.total > 0 {
		percentage = (float64(current) / float64(e.total)) * 100
	}

	eta := int64(0)
	if speed > 0 && e.total > current {
		eta = int64(float64(e.total-current) / speed)
	}

	lang, err := database.Language(e.msg.ChatId)
	if err != nil {
		lang = "en"
	}

	text := locales.Get(lang, "download_progress", locales.Arg{
		"percentage": fmt.Sprintf("%.1f", percentage),
		"speed":      utils.FormatBytes(int64(speed)) + "/s",
		"eta":        utils.FormatDuration(int(eta)),
		"elapsed":    utils.FormatTime(int(elapsed)),
	})

	if _, err := e.msg.EditText(c, text, &td.EditTextMessageOpts{
		ParseMode: td.ParseModeHTML,
	}); err != nil {
		dp.Stop(fileID)
	}

	return nil
}
