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

package core

import (
	"io"

	"github.com/amarnathcjd/gogram/telegram"

	"yukkimusic/config"
	"yukkimusic/internal/logger"
)

type TgLogger struct {
	gl  *logger.Logger
	lvl telegram.LogLevel
}

func GetTgLogger(name string, lvl telegram.LogLevel) *TgLogger {
	l := &TgLogger{
		gl:  logger.GetLogger(name),
		lvl: lvl,
	}
	l.SetLevel(lvl)
	l.SetOutput(config.LogWriter)
	return l
}

func (l *TgLogger) Debug(msg any, a ...any) {
	if l.lvl <= telegram.DebugLevel {
		l.gl.Debugf("%v %v", msg, a)
	}
}

func (l *TgLogger) Info(msg any, a ...any) {
	if l.lvl <= telegram.InfoLevel {
		l.gl.Infof("%v %v", msg, a)
	}
}

func (l *TgLogger) Warn(msg any, a ...any) {
	if l.lvl <= telegram.WarnLevel {
		l.gl.Warnf("%v %v", msg, a)
	}
}

func (l *TgLogger) Error(msg any, a ...any) {
	if l.lvl <= telegram.ErrorLevel {
		l.gl.Errorf("%v %v", msg, a)
	}
}

func (l *TgLogger) SetLevel(v telegram.LogLevel) {
	l.lvl = v
	switch v {
	case telegram.TraceLevel, telegram.DebugLevel:
		l.gl.SetLevel(logger.DebugLevel)
	case telegram.InfoLevel:
		l.gl.SetLevel(logger.InfoLevel)
	case telegram.WarnLevel:
		l.gl.SetLevel(logger.WarnLevel)
	case telegram.ErrorLevel, telegram.PanicLevel:
		l.gl.SetLevel(logger.ErrorLevel)
	case telegram.FatalLevel:
		l.gl.SetLevel(logger.FatalLevel)
	default:
		l.gl.SetLevel(logger.InfoLevel)
	}
}

func (l *TgLogger) GetLevel() telegram.LogLevel {
	return l.lvl
}

func (l *TgLogger) SetOutput(w any) {
	if ww, ok := w.(io.Writer); ok {
		l.gl.SetOutput(ww)
	}
}

func (l *TgLogger) GetOutput() any {
	return l.gl
}

func (l *TgLogger) SetTimestampFormat(s string) {}
