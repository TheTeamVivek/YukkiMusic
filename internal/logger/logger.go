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

// Package logger provides a lightweight, standard-library-only logger with
// ANSI colors when the output is a terminal and plain text otherwise.
package logger

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Level represents a log severity. Higher values are more severe.
type Level int32

const (
	// DebugLevel is the least severe level.
	DebugLevel Level = 1 << iota
	// InfoLevel is the info level.
	InfoLevel
	// WarnLevel is the warn level.
	WarnLevel
	// ErrorLevel is the error level.
	ErrorLevel
	// FatalLevel is the most severe level and exits the process after logger.
	FatalLevel
)

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return fmt.Sprintf("LEVEL(%d)", int32(l))
	}
}

// ANSI escape codes. Only used when colored output is enabled.
const (
	ansiReset   = "\x1b[0m"
	ansiDim     = "\x1b[2m"
	ansiCyan    = "\x1b[36m"
	ansiGreen   = "\x1b[32m"
	ansiYellow  = "\x1b[33m"
	ansiRed     = "\x1b[31m"
	ansiBoldRed = "\x1b[1;31m"
)

var (
	registryMu   sync.Mutex
	namedLoggers = map[string]*Logger{}

	defaultLogger = newLogger("")
)

// Logger writes formatted log lines to an io.Writer.
type callSite struct {
	file string
	fn   string
	line int
}

type Logger struct {
	mu         sync.Mutex
	level      Level
	w          io.Writer
	name       string
	timeFormat string
	colored    bool
	colorSet   bool
}

func newLogger(name string) *Logger {
	return &Logger{
		level:      InfoLevel,
		w:          os.Stderr,
		name:       name,
		timeFormat: "2006-01-02 15:04:05",
		colored:    colorEnabled(),
	}
}

// GetLogger returns a named logger, creating it on first use.
func GetLogger(name string) *Logger {
	registryMu.Lock()
	defer registryMu.Unlock()
	if l, ok := namedLoggers[name]; ok {
		return l
	}
	l := newLogger(name)
	l.SetLevel(defaultLogger.GetLevel())
	namedLoggers[name] = l
	return l
}

// SetLevel sets the minimum level of the default logger.
func SetLevel(level Level) {
	defaultLogger.SetLevel(level)
}

// GetLevel returns the current level of the default logger.
func GetLevel() Level {
	return defaultLogger.GetLevel()
}

// SetOutput sets the output writer of the default logger.
func SetOutput(w io.Writer) {
	defaultLogger.SetOutput(w)
}

// SetColored forces color on or off for the default logger.
func SetColored(enabled bool) {
	defaultLogger.SetColored(enabled)
}

func Debug(message ...any) {
	defaultLogger.log(DebugLevel, caller(2), fmt.Sprint(message...))
}

func Debugf(format string, args ...any) {
	defaultLogger.log(DebugLevel, caller(2), fmt.Sprintf(format, args...))
}

func Info(message ...any) {
	defaultLogger.log(InfoLevel, caller(2), fmt.Sprint(message...))
}

func Infof(format string, args ...any) {
	defaultLogger.log(InfoLevel, caller(2), fmt.Sprintf(format, args...))
}

func Warn(message ...any) {
	defaultLogger.log(WarnLevel, caller(2), fmt.Sprint(message...))
}

func Warnf(format string, args ...any) {
	defaultLogger.log(WarnLevel, caller(2), fmt.Sprintf(format, args...))
}

func Error(message ...any) {
	defaultLogger.log(ErrorLevel, caller(2), fmt.Sprint(message...))
}

func Errorf(format string, args ...any) {
	defaultLogger.log(ErrorLevel, caller(2), fmt.Sprintf(format, args...))
}

func Fatal(message ...any) {
	defaultLogger.log(FatalLevel, caller(2), fmt.Sprint(message...))
}

func Fatalf(format string, args ...any) {
	defaultLogger.log(FatalLevel, caller(2), fmt.Sprintf(format, args...))
}

// SetLevel sets the minimum level that will be logged.
func (ctx *Logger) SetLevel(level Level) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.level = level
}

// GetLevel returns the current minimum level.
func (ctx *Logger) GetLevel() Level {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return ctx.level
}

// SetOutput sets the output writer. A nil writer defaults to os.Stderr.
func (ctx *Logger) SetOutput(w io.Writer) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if w == nil {
		w = os.Stderr
	}
	ctx.w = w
	if !ctx.colorSet {
		ctx.colored = isTerminal(w) && !colorDisabled()
	}
}

// SetColored forces color on or off regardless of terminal detection.
func (ctx *Logger) SetColored(enabled bool) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.colored = enabled
	ctx.colorSet = true
}

// GetOutput returns the current output writer.
func (ctx *Logger) GetOutput() any {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return ctx.w
}

func (ctx *Logger) Debug(message ...any) {
	ctx.log(DebugLevel, caller(2), fmt.Sprint(message...))
}

func (ctx *Logger) Debugf(format string, args ...any) {
	ctx.log(DebugLevel, caller(2), fmt.Sprintf(format, args...))
}

func (ctx *Logger) Info(message ...any) {
	ctx.log(InfoLevel, caller(2), fmt.Sprint(message...))
}

func (ctx *Logger) Infof(format string, args ...any) {
	ctx.log(InfoLevel, caller(2), fmt.Sprintf(format, args...))
}

func (ctx *Logger) Warn(message ...any) {
	ctx.log(WarnLevel, caller(2), fmt.Sprint(message...))
}

func (ctx *Logger) Warnf(format string, args ...any) {
	ctx.log(WarnLevel, caller(2), fmt.Sprintf(format, args...))
}

func (ctx *Logger) Error(message ...any) {
	ctx.log(ErrorLevel, caller(2), fmt.Sprint(message...))
}

func (ctx *Logger) Errorf(format string, args ...any) {
	ctx.log(ErrorLevel, caller(2), fmt.Sprintf(format, args...))
}

func (ctx *Logger) Fatal(message ...any) {
	ctx.log(FatalLevel, caller(2), fmt.Sprint(message...))
}

func (ctx *Logger) Fatalf(format string, args ...any) {
	ctx.log(FatalLevel, caller(2), fmt.Sprintf(format, args...))
}

func (ctx *Logger) log(level Level, cs callSite, msg string) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	if level < DebugLevel || level > FatalLevel {
		return
	}
	if ctx.level > level {
		return
	}
	if ctx.w == io.Discard {
		if level == FatalLevel {
			os.Exit(1)
		}
		return
	}

	var b strings.Builder
	ctx.formatLine(&b, level, cs.file, cs.fn, cs.line, msg)
	if level == FatalLevel {
		ctx.formatStack(&b, 4)
	}
	b.WriteByte('\n')
	_, _ = io.WriteString(ctx.w, b.String())

	if level == FatalLevel {
		os.Exit(1)
	}
}

func (ctx *Logger) formatLine(b *strings.Builder, level Level, file string, fn string, line int, msg string) {
	timestamp := time.Now().Format(ctx.timeFormat)
	if ctx.colored {
		fmt.Fprintf(b, "%s%s%s", ansiDim, timestamp, ansiReset)
	} else {
		b.WriteString(timestamp)
	}

	b.WriteString(" [")
	levelTag := fmt.Sprintf("%-5s", level.String())
	if ctx.colored {
		fmt.Fprintf(b, "%s%s%s", levelColor(level), levelTag, ansiReset)
	} else {
		b.WriteString(levelTag)
	}
	b.WriteByte(']')

	if ctx.name != "" {
		b.WriteString(" [" + ctx.name + "]")
	}

	b.WriteByte(' ')
	b.WriteString(msg)

	if file != "" {
		callerStr := fmt.Sprintf("(%s:%d", file, line)
		if level == DebugLevel && fn != "" {
			callerStr += " " + fn
		}
		callerStr += ")"
		if ctx.colored {
			callerStr = ansiDim + callerStr + ansiReset
		}
		b.WriteByte(' ')
		b.WriteString(callerStr)
	}
}

func (ctx *Logger) formatStack(b *strings.Builder, skip int) {
	pcs := make([]uintptr, 8)
	n := runtime.Callers(skip, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		fr, more := frames.Next()
		fmt.Fprintf(b, "\n  at %s (%s:%d)", trimFunc(fr.Function), path.Base(fr.File), fr.Line)
		if !more {
			break
		}
	}
}

func caller(skip int) callSite {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return callSite{}
	}
	var fn string
	if f := runtime.FuncForPC(pc); f != nil {
		fn = trimFunc(f.Name())
	}
	return callSite{file: path.Base(file), fn: fn, line: line}
}

func trimFunc(name string) string {
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	return name
}

func levelColor(level Level) string {
	switch level {
	case DebugLevel:
		return ansiCyan
	case InfoLevel:
		return ansiGreen
	case WarnLevel:
		return ansiYellow
	case ErrorLevel:
		return ansiRed
	case FatalLevel:
		return ansiBoldRed
	default:
		return ""
	}
}

func colorEnabled() bool {
	return isTerminal(os.Stderr) && !colorDisabled()
}

func colorDisabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return true
	}
	if strings.EqualFold(os.Getenv("TERM"), "dumb") {
		return true
	}
	if v, _ := strconv.ParseBool(os.Getenv("DISABLE_COLOUR")); v {
		return true
	}
	return false
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
