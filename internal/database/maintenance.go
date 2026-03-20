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

func SetMaintenance(enabled bool, reason ...string) error {
	return modifyBotState(func(s *BotState) bool {
		changed := false
		if s.Maintenance.Enabled != enabled {
			s.Maintenance.Enabled = enabled
			changed = true
		}

		newReason := ""
		if enabled && len(reason) > 0 {
			newReason = reason[0]
		}

		if s.Maintenance.Reason != newReason {
			s.Maintenance.Reason = newReason
			changed = true
		}
		return changed
	})
}

func MaintenanceReason() (string, error) {
	state, err := getBotState()
	if err != nil {
		return "", err
	}
	return state.Maintenance.Reason, nil
}

func IsMaintenanceEnabled() (bool, error) {
	state, err := getBotState()
	if err != nil {
		return false, err
	}
	return state.Maintenance.Enabled, nil
}
