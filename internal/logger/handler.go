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

package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
)

// NewHandler returns a slog.Handler that formats records with the same layout
// as Logger. It lets external libraries (e.g. gotdbot) route their logs
// through this package. A nil writer defaults to os.Stderr.
func NewHandler(w io.Writer, level Level) slog.Handler {
	if w == nil {
		w = os.Stderr
	}
	return &slogHandler{
		level: level,
		w:     w,
		mu:    &sync.Mutex{},
	}
}

type slogHandler struct {
	mu    *sync.Mutex
	level Level
	w     io.Writer
	attrs []slog.Attr
	group string
}

func (h *slogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return minSlogLevel(h.level) <= level
}

func (h *slogHandler) Handle(_ context.Context, r slog.Record) error {
	if !h.Enabled(context.Background(), r.Level) {
		return nil
	}

	var b strings.Builder
	level := levelOf(r.Level)

	timestamp := r.Time.Format("2006-01-02 15:04:05")
	b.WriteString(timestamp)

	b.WriteString(" [")
	levelTag := fmt.Sprintf("%-5s", level.String())
	b.WriteString(levelTag)
	b.WriteByte(']')

	if h.group != "" {
		b.WriteString(" [" + h.group + "]")
	}

	b.WriteByte(' ')
	b.WriteString(r.Message)

	if r.PC != 0 {
		if fr, ok := runtime.CallersFrames([]uintptr{r.PC}).Next(); ok {
			callerStr := fmt.Sprintf("(%s:%d)", path.Base(fr.File), fr.Line)
			b.WriteByte(' ')
			b.WriteString(callerStr)
		}
	}

	r.Attrs(func(a slog.Attr) bool {
		b.WriteString(" " + a.Key + "=" + a.Value.String())
		return true
	})
	for _, a := range h.attrs {
		b.WriteString(" " + a.Key + "=" + a.Value.String())
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	if h.w == io.Discard {
		return nil
	}
	_, err := io.WriteString(h.w, b.String()+"\n")
	return err
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	n := new(slogHandler)
	*n = *h
	n.attrs = append(n.attrs[:len(n.attrs):len(n.attrs)], attrs...)
	return n
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	n := new(slogHandler)
	*n = *h
	if n.group != "" {
		n.group += "." + name
	} else {
		n.group = name
	}
	return n
}

// levelOf maps a slog.Level to the closest Logger level.
func levelOf(l slog.Level) Level {
	switch {
	case l <= slog.LevelDebug:
		return DebugLevel
	case l <= slog.LevelInfo:
		return InfoLevel
	case l <= slog.LevelWarn:
		return WarnLevel
	case l <= slog.LevelError:
		return ErrorLevel
	default:
		return FatalLevel
	}
}

// minSlogLevel maps a Logger level to the minimum accepted slog.Level.
func minSlogLevel(l Level) slog.Level {
	switch l {
	case DebugLevel:
		return slog.LevelDebug
	case WarnLevel:
		return slog.LevelWarn
	case ErrorLevel:
		return slog.LevelError
	case FatalLevel:
		return slog.LevelError + 1
	default:
		return slog.LevelInfo
	}
}
