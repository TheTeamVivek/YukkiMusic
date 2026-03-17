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

package database

import "slices"

// GetServedChats returns all chats that have used the bot.
func GetServedChats() ([]int64, error) {
	state, err := getBotState()
	if err != nil {
		return nil, err
	}
	return state.Served.Chats, nil
}

// GetServedUsers returns all users that have used the bot in private.
func GetServedUsers() ([]int64, error) {
	state, err := getBotState()
	if err != nil {
		return nil, err
	}
	return state.Served.Users, nil
}

// IsServedChat checks if the chat is already in the served list.
func IsServedChat(id int64) (bool, error) {
	state, err := getBotState()
	if err != nil {
		return false, err
	}
	_, ok := state.servedChatsMap[id]
	return ok, nil
}

// IsServedUser checks if the user is already in the served list.
func IsServedUser(id int64) (bool, error) {
	state, err := getBotState()
	if err != nil {
		return false, err
	}
	_, ok := state.servedUsersMap[id]
	return ok, nil
}

// AddServedChat adds a chat to the served list.
func AddServedChat(id int64) error {
	state, err := getBotState()
	if err != nil {
		return err
	}

	if _, ok := state.servedChatsMap[id]; ok {
		return nil
	}

	state.Served.Chats = append(state.Served.Chats, id)
	state.servedChatsMap[id] = struct{}{}

	return updateBotState(state)
}

// AddServedUser adds a user to the served list.
func AddServedUser(id int64) error {
	state, err := getBotState()
	if err != nil {
		return err
	}

	if _, ok := state.servedUsersMap[id]; ok {
		return nil
	}

	state.Served.Users = append(state.Served.Users, id)
	state.servedUsersMap[id] = struct{}{}

	return updateBotState(state)
}

// DeleteServedChat removes a chat from the served list.
func DeleteServedChat(id int64) error {
	state, err := getBotState()
	if err != nil {
		return err
	}

	if _, ok := state.servedChatsMap[id]; !ok {
		return nil
	}

	delete(state.servedChatsMap, id)
	state.Served.Chats = slices.DeleteFunc(state.Served.Chats, func(v int64) bool {
		return v == id
	})

	return updateBotState(state)
}

// DeleteServedUser removes a user from the served list.
func DeleteServedUser(id int64) error {
	state, err := getBotState()
	if err != nil {
		return err
	}

	if _, ok := state.servedUsersMap[id]; !ok {
		return nil
	}

	delete(state.servedUsersMap, id)
	state.Served.Users = slices.DeleteFunc(state.Served.Users, func(v int64) bool {
		return v == id
	})

	return updateBotState(state)
}
