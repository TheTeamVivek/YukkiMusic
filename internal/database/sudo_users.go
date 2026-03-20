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

func Sudoers() ([]int64, error) {
	state, err := getBotState()
	if err != nil {
		return nil, err
	}
	return state.Sudoers, nil
}

func IsSudoWithoutError(id int64) bool {
	is, _ := IsSudo(id)
	return is
}

func IsSudo(id int64) (bool, error) {
	state, err := getBotState()
	if err != nil {
		return false, err
	}
	return contains(state.Sudoers, id), nil
}

func AddSudo(id int64) error {
	return modifyBotState(func(s *BotState) bool {
		var added bool
		s.Sudoers, added = addUnique(s.Sudoers, id)
		return added
	})
}

func RemoveSudo(id int64) error {
	return modifyBotState(func(s *BotState) bool {
		var removed bool
		s.Sudoers, removed = removeElement(s.Sudoers, id)
		return removed
	})
}
