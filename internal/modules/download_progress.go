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
	"time"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/platforms"
	"yukkimusic/internal/utils"
)

const downloadEditInterval = 2 * time.Second

// dlEntry holds the state of one active download.
type dlEntry struct {
	msg      *td.Message
	started  time.Time
	lastEdit time.Time
	lastSize int64
	total    int64
}

// downloadProg tracks active downloads by fileID.
var downloadProg = utils.NewCache[string, *dlEntry](30 * time.Minute)

func init() {
	platforms.OnDownloadStart = downloadProgressStart
}

// downloadProgressStart tracks download progress for fileID and edits msg
// as the file downloads.
func downloadProgressStart(fileID string, msg *td.Message) {
	if fileID == "" || msg == nil {
		return
	}
	downloadProg.Set(fileID, &dlEntry{msg: msg, started: time.Now()})
}

// downloadUpdateHandler is registered in handlers.go and drives the
// "Downloading..." progress edits. Entries clean themselves up once the
// download finishes or is cancelled, so no stop hook is needed.
func downloadUpdateHandler(c *td.Client, u *td.UpdateFile) error {
	if u == nil || u.File == nil || u.File.Remote == nil || u.File.Local == nil {
		return nil
	}

	e, ok := downloadProg.Get(u.File.Remote.Id)
	if !ok {
		return nil
	}

	now := time.Now()
	done := !u.File.Local.IsDownloadingActive
	if now.Sub(e.lastEdit) < downloadEditInterval && !done {
		return nil
	}
	interval := now.Sub(e.lastEdit).Seconds()
	e.lastEdit = now

	total := u.File.Size
	if total <= 0 {
		total = u.File.ExpectedSize
	}
	if total > 0 {
		e.total = total
	}

	current := min(u.File.Local.DownloadedSize, e.total)

	speed := 0.0
	if interval > 0 && current >= e.lastSize {
		speed = float64(current-e.lastSize) / interval
	}
	e.lastSize = current

	percentage := 0.0
	if e.total > 0 {
		percentage = float64(current) / float64(e.total) * 100
	}

	eta := int64(0)
	if speed > 0 && e.total > current {
		eta = int64(float64(e.total-current) / speed)
	}

	elapsed := now.Sub(e.started).Seconds()

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
		// Ignore edit errors; the download keeps going and the next
		// update will retry. The entry is cleaned up once the download
		// finishes or is cancelled.
		return nil
	}

	if done {
		downloadProg.Delete(u.File.Remote.Id)
	}

	return nil
}
