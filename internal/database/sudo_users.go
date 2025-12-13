/*
 * This file is part of YukkiMusic.
 *
 * YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
 * Copyright (C) 2025 TheTeamVivek
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 */
package database

// GetSudoers returns all sudoers.
func GetSudoers() ([]int64, error) {
	state, err := getBotState()
	if err != nil {
		logger.ErrorF("Failed to get current sudoers: %v", err)
		return nil, err
	}
	return state.Sudoers, nil
}

func IsSudoWithoutError(id int64) bool {
	is, _ := IsSudo(id)
	return is
}

// IsSudo checks if the given ID is a sudoer.
func IsSudo(id int64) (bool, error) {
	state, err := getBotState()
	if err != nil {
		logger.ErrorF("Failed to get current sudoers: %v", err)
		return false, err
	}

	for _, v := range state.Sudoers {
		if v == id {
			return true, nil
		}
	}
	return false, nil
}

// AddSudo adds a new sudoer if not already present.
func AddSudo(id int64) error {
	exists, err := IsSudo(id)
	if err != nil {
		logger.ErrorF("Failed to check sudo existence: %v", err)
		return err
	}
	if exists {
		return nil
	}

	state, err := getBotState()
	if err != nil {
		logger.ErrorF("Failed to get current sudoers: %v", err)
		return err
	}

	state.Sudoers = append(state.Sudoers, id)
	if err := updateBotState(state); err != nil {
		logger.ErrorF("Failed to update sudoers: %v", err)
		return err
	}

	return nil
}

// DeleteSudo removes a sudoer by ID.
func DeleteSudo(id int64) error {
	exists, err := IsSudo(id)
	if err != nil {
		logger.ErrorF("Failed to check sudo existence: %v", err)
		return err
	}
	if !exists {
		return nil
	}

	state, err := getBotState()
	if err != nil {
		logger.ErrorF("Failed to get current sudoers: %v", err)
		return err
	}

	newSudoers := make([]int64, 0, len(state.Sudoers))
	for _, v := range state.Sudoers {
		if v != id {
			newSudoers = append(newSudoers, v)
		}
	}
	state.Sudoers = newSudoers

	if err := updateBotState(state); err != nil {
		logger.ErrorF("Failed to update sudoers: %v", err)
		return err
	}

	return nil
}
