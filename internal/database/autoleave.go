/*
 * ● YukkiMusic
 * ○ A high-performance engine for streaming music in Telegram voicechats.
 * ________________________________________________________________________________________
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
 * ________________________________________________________________________________________
 */

package database

func AutoLeave() (bool, error) {
	state, err := getBotState()
	if err != nil {
		return false, err
	}
	return state.AutoLeave, nil
}

func SetAutoLeave(value bool) error {
	return modifyBotState(func(s *BotState) bool {
		if s.AutoLeave == value {
			return false
		}
		s.AutoLeave = value
		return true
	})
}
