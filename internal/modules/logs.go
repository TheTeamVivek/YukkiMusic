/*
  - This file is part of YukkiMusic.
    *

  - YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
  - Copyright (C) 2025 TheTeamVivek
    *
  - This program is free software: you can redistribute it and/or modify
  - it under the terms of the GNU General Public License as published by
  - the Free Software Foundation, either version 3 of the License, or
  - (at your option) any later version.
    *
  - This program is distributed in the hope that it will be useful,
  - but WITHOUT ANY WARRANTY; without even the implied warranty of
  - MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
  - GNU General Public License for more details.
    *
  - You should have received a copy of the GNU General Public License
  - along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package modules

import (
	"bufio"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/locales"
)

const (
	maxMessageLength = 2000
	defaultLineCount = 50
)

var logClearMutex sync.Mutex

func init() {
	helpTexts["/logs"] = `<i>Download or view bot logs.</i>

<u>Usage:</u>
<b>/logs</b> — Send current log file or show content
<b>/logs [n]</b> — Get last N lines from recent logs
<b>/logs -old [n]</b> — Get first N lines from oldest logs
<b>/logs -clear</b> — Clear current log file

<b>📋 Examples:</b>
• <code>/logs</code> — Get full current log
• <code>/logs 50</code> — Last 50 lines from recent
• <code>/logs 100</code> — Last 100 lines from recent
• <code>/logs -old 50</code> — First 50 lines from oldest
• <code>/logs -clear</code> — Delete current log

<b>🔒 Restrictions:</b>
• <b>Sudo users only</b>

<b>⚠️ Notes:</b>
• If content < 2000 chars, shows as code preview
• Otherwise sends as file
• N can be any positive number (default: 50)`
}

func logsHandler(m *tg.NewMessage) error {
	chatID := m.ChannelID()
	text := strings.TrimSpace(m.Text())
	parts := strings.Fields(text)

	logFile := config.LogFileName

	if len(parts) > 1 && (parts[1] == "-clear" || parts[1] == "--clear") {
		return handleLogsClear(m, logFile)
	}

	if len(parts) > 1 && (parts[1] == "-old" || parts[1] == "--old") {
		lines := defaultLineCount
		if len(parts) > 2 {
			if n, err := strconv.Atoi(parts[2]); err == nil && n > 0 {
				lines = n
			} else if err != nil {
				m.Reply(F(chatID, "logs_invalid_number", locales.Arg{
					"value": parts[2],
				}))
				return tg.ErrEndGroup
			}
		}
		return handleLogsOld(m, logFile, lines)
	}

	if len(parts) > 1 {
		if n, err := strconv.Atoi(parts[1]); err == nil && n > 0 {
			return handleLogsRecent(m, logFile, n)
		} else if err != nil {
			m.Reply(F(chatID, "logs_invalid_flag", locales.Arg{
				"flag": parts[1],
			}))
			return tg.ErrEndGroup
		}
	}

	return handleLogsDefault(m, logFile)
}

func handleLogsDefault(m *tg.NewMessage, logFile string) error {
	chatID := m.ChannelID()

	info, err := os.Stat(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			m.Reply(F(chatID, "logs_empty"))
			return tg.ErrEndGroup
		}
		m.Reply(F(chatID, "logs_file_error", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	if info.Size() > 50*1024*1024 {
		return sendLogFile(m, logFile, info, chatID)
	}

	content, err := os.ReadFile(logFile)
	if err != nil {
		m.Reply(F(chatID, "logs_file_error", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	contentStr := strings.TrimSpace(string(content))

	if contentStr == "" {
		m.Reply(F(chatID, "logs_empty"))
		return tg.ErrEndGroup
	}

	if len(contentStr) < maxMessageLength-100 {
		escapedContent := html.EscapeString(contentStr)
		msg := F(
			chatID,
			"logs_preview_header",
		) + "\n\n<pre>" + escapedContent + "</pre>"
		m.Reply(msg)
		return tg.ErrEndGroup
	}

	return sendLogFile(m, logFile, info, chatID)
}

func sendLogFile(
	m *tg.NewMessage,
	logFile string,
	info os.FileInfo,
	chatID int64,
) error {
	mystic, _ := m.Reply(F(chatID, "logs_uploading"))

	fileSizeMB := float64(info.Size()) / 1024 / 1024
	caption := F(chatID, "logs_file_caption", locales.Arg{
		"filename": filepath.Base(logFile),
		"size_mb":  fmt.Sprintf("%.2f", fileSizeMB),
		"modified": info.ModTime().Format("2006-01-02 15:04:05"),
	})

	_, err := m.Client.SendMedia(chatID, logFile, &tg.MediaOptions{
		Caption: caption,
	})

	if mystic != nil {
		mystic.Delete()
	}

	if err != nil {
		m.Reply(F(chatID, "logs_send_error", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	return tg.ErrEndGroup
}

func handleLogsRecent(m *tg.NewMessage, logFile string, lines int) error {
	chatID := m.ChannelID()

	info, err := os.Stat(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			m.Reply(F(chatID, "logs_empty"))
			return tg.ErrEndGroup
		}
		m.Reply(F(chatID, "logs_file_error", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	if info.Size() == 0 {
		m.Reply(F(chatID, "logs_empty"))
		return tg.ErrEndGroup
	}

	selectedLines, err := readLastLines(logFile, lines)
	if err != nil {
		m.Reply(F(chatID, "logs_file_error", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	if len(selectedLines) == 0 {
		m.Reply(F(chatID, "logs_empty"))
		return tg.ErrEndGroup
	}

	output := strings.Join(selectedLines, "\n")

	if len(output) < maxMessageLength-100 {
		escapedOutput := html.EscapeString(output)
		msg := F(chatID, "logs_lines_header", locales.Arg{
			"count": len(selectedLines),
		}) + "\n\n<pre>" + escapedOutput + "</pre>"
		m.Reply(msg)
		return tg.ErrEndGroup
	}

	return sendLogLines(m, selectedLines, chatID, "logs_lines_caption")
}

func handleLogsOld(m *tg.NewMessage, logFile string, lines int) error {
	chatID := m.ChannelID()

	info, err := os.Stat(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			m.Reply(F(chatID, "logs_empty"))
			return tg.ErrEndGroup
		}
		m.Reply(F(chatID, "logs_file_error", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	if info.Size() == 0 {
		m.Reply(F(chatID, "logs_empty"))
		return tg.ErrEndGroup
	}

	selectedLines, err := readFirstLines(logFile, lines)
	if err != nil {
		m.Reply(F(chatID, "logs_file_error", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	if len(selectedLines) == 0 {
		m.Reply(F(chatID, "logs_empty"))
		return tg.ErrEndGroup
	}

	output := strings.Join(selectedLines, "\n")

	if len(output) < maxMessageLength-100 {
		escapedOutput := html.EscapeString(output)
		msg := F(chatID, "logs_old_header", locales.Arg{
			"count": len(selectedLines),
		}) + "\n\n<pre>" + escapedOutput + "</pre>"
		m.Reply(msg)
		return tg.ErrEndGroup
	}

	tmpFile := filepath.Join(
		os.TempDir(),
		fmt.Sprintf("logs_old_%d.txt", time.Now().Unix()),
	)
	if err := os.WriteFile(tmpFile, []byte(output), 0o644); err != nil {
		m.Reply(F(chatID, "logs_send_error", locales.Arg{"error": err.Error()}))
		return tg.ErrEndGroup
	}
	defer os.Remove(tmpFile)

	caption := F(chatID, "logs_old_caption", locales.Arg{
		"count": len(selectedLines),
	})

	_, err = m.ReplyMedia(tmpFile, &tg.MediaOptions{
		Caption: caption,
	})
	if err != nil {
		m.Reply(F(chatID, "logs_send_error", locales.Arg{
			"error": err.Error(),
		}))
	}

	return tg.ErrEndGroup
}

func handleLogsClear(m *tg.NewMessage, logFile string) error {
	chatID := m.ChannelID()

	logClearMutex.Lock()
	defer logClearMutex.Unlock()

	info, err := os.Stat(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			m.Reply(F(chatID, "logs_empty"))
			return tg.ErrEndGroup
		}
		m.Reply(F(chatID, "logs_file_error", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	fileSizeMB := float64(info.Size()) / 1024 / 1024

	if err := os.Truncate(logFile, 0); err != nil {
		m.Reply(F(chatID, "logs_clear_failed", locales.Arg{
			"error": err.Error(),
		}))
		gologging.ErrorF("Failed to clear log file: %v", err)
		return tg.ErrEndGroup
	}

	gologging.InfoF(
		"Log file cleared by user %d (was %.2f MB)",
		m.SenderID(),
		fileSizeMB,
	)

	m.Reply(F(chatID, "logs_cleared", locales.Arg{
		"size_mb": fmt.Sprintf("%.2f", fileSizeMB),
	}))

	return tg.ErrEndGroup
}

func readLastLines(filePath string, n int) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var allLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			allLines = append(allLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	start := len(allLines) - n
	if start < 0 {
		start = 0
	}

	return allLines[start:], nil
}

func readFirstLines(filePath string, n int) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var lines []string
	count := 0
	for scanner.Scan() && count < n {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
			count++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func sendLogLines(
	m *tg.NewMessage,
	lines []string,
	chatID int64,
	captionKey string,
) error {
	output := strings.Join(lines, "\n")

	tmpFile := filepath.Join(
		os.TempDir(),
		fmt.Sprintf("logs_%d.txt", time.Now().Unix()),
	)
	if err := os.WriteFile(tmpFile, []byte(output), 0o644); err != nil {
		m.Reply(F(chatID, "logs_send_error", locales.Arg{"error": err.Error()}))
		return tg.ErrEndGroup
	}
	defer os.Remove(tmpFile)

	caption := F(chatID, captionKey, locales.Arg{
		"count": len(lines),
	})

	_, err := m.Client.SendMedia(chatID, tmpFile, &tg.MediaOptions{
		Caption: caption,
	})
	if err != nil {
		m.Reply(F(chatID, "logs_send_error", locales.Arg{
			"error": err.Error(),
		}))
	}

	return tg.ErrEndGroup
}
