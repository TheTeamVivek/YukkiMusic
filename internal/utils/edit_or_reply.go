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
	"fmt"
	"runtime"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/logger"
)

func callerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown:0"
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// EOR edits a message, falling back to delete + reply on edit failure.
func EOR(
	c *td.Client,
	msg *td.Message,
	text string,
	opts *td.EditTextMessageOpts,
) (*td.Message, error) {
	if msg == nil {
		logger.Error("[EOR] nil msg at " + callerInfo(2))
		return nil, nil
	}

	m, err := msg.EditText(c, text, opts)
	if err != nil {
		_ = msg.Delete(c, true)
		replyOpts := &td.SendTextMessageOpts{}
		if opts != nil {
			replyOpts.ParseMode = opts.ParseMode
			replyOpts.ReplyMarkup = opts.ReplyMarkup
		}
		m, err = msg.ReplyText(c, text, replyOpts)
	}

	if err != nil {
		logger.Error(
			"[EOR] " + err.Error() + " | called from " + callerInfo(2),
		)
	}
	return m, err
}

// ToInt64s converts a slice of int32 message IDs to a slice of int64.
func ToInt64s(ids []int32) []int64 {
	out := make([]int64, len(ids))
	for i, id := range ids {
		out[i] = int64(id)
	}
	return out
}
