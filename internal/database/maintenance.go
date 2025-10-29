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

// SetMaintenance sets the maintenance mode.
// If enabling, you can provide an optional reason.
// If disabling, it clears any existing reason.
func SetMaintenance(enabled bool, reason ...string) error {
	ctx, cancel := mongoCtx()
	defer cancel()

	state, err := getBotState(ctx)
	if err != nil {
		return err
	}

	if state.Maintenance == enabled {
		// already in desired state, just update reason if enabling
		if enabled && len(reason) > 0 {
			state.MaintReason = reason[0]
			return updateBotState(ctx, state)
		} else if !enabled && state.MaintReason != "" {
			// clear reason when disabling
			state.MaintReason = ""
			return updateBotState(ctx, state)
		}
		return nil
	}

	state.Maintenance = enabled
	if enabled && len(reason) > 0 {
		state.MaintReason = reason[0]
	} else if !enabled {
		state.MaintReason = ""
	}

	return updateBotState(ctx, state)
}

// GetMaintReason retrieves the current maintenance reason
func GetMaintReason() (string, error) {
	ctx, cancel := mongoCtx()
	defer cancel()

	state, err := getBotState(ctx)
	if err != nil {
		return "", err
	}
	return state.MaintReason, nil
}

// IsMaintenance returns whether maintenance mode is enabled
func IsMaintenance() (bool, error) {
	ctx, cancel := mongoCtx()
	defer cancel()

	state, err := getBotState(ctx)
	if err != nil {
		return false, err
	}
	return state.Maintenance, nil
}
