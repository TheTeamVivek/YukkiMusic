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

package utils

import (
	"io"
	"log/slog"
	"os"
	"strings"

	gotdlogger "github.com/AshokShau/gotdbot/logger"
)

var (
	logFile *os.File

	// Log is the shared application logger. Every message is written to both
	// stderr and logs.txt.
	//
	// The minimum level is set via the LOG_LEVEL env var
	// (debug, info, warn, error); defaults to info.
	Log = gotdlogger.New(
		gotdlogger.WithOutput(logOutput()),
		gotdlogger.WithLevel(logLevel()),
	)
)

func logLevel() slog.Level {
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func logOutput() io.Writer {
	f, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return os.Stderr
	}
	logFile = f
	return io.MultiWriter(os.Stderr, f)
}

// CloseLog closes the log file, if it was opened.
func CloseLog() {
	if logFile != nil {
		_ = logFile.Close()
	}
}
