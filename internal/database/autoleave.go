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

func GetAutoLeave() (bool, error) {
	state, err := getBotState()
	if err != nil {
		return false, err
	}
	return state.AutoLeave, nil
}

func SetAutoLeave(value bool) error {
	state, err := getBotState()
	if err != nil {
		logger.ErrorF("Failed to get bot state for setting AutoEnd: %v", err)
		return err
	}
	if state.AutoLeave == value {
		return nil
	}

	state.AutoLeave = value
	if err := updateBotState(state); err != nil {
		logger.ErrorF("Failed to update AutoEnd: %v", err)
		return err
	}
	return nil
}
