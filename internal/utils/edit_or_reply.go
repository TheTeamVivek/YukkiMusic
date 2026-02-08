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
package utils

import (
	"fmt"
	"runtime"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"
)

func EOR(
	msg *telegram.NewMessage,
	text string,
	opts ...*telegram.SendOptions,
) (m *telegram.NewMessage, err error) {
	if msg == nil {
		gologging.Error("[EOR] nil msg at " + callerInfo(2))
		return nil, nil
	}

	m, err = msg.Edit(text, opts...)
	if err != nil {
		msg.Delete()
		m, err = msg.Respond(text, opts...)
	}

	if err != nil {
		gologging.Error(
			"[EOR] " + err.Error() +
				" | called from " + callerInfo(2),
		)
	}
	return m, err
}

func callerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown:0"
	}
	return fmt.Sprintf("%s:%d", file, line)
}
