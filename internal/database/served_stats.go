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

func GetServed(user ...bool) ([]int64, error) {
	state, err := getBotState()
	if err != nil {
		return nil, err
	}

	if len(user) > 0 && user[0] {
		return state.Served.Users, nil
	}

	return state.Served.Chats, nil
}

func IsServed(id int64, user ...bool) (bool, error) {
	state, err := getBotState()
	if err != nil {
		return false, err
	}

	if len(user) > 0 && user[0] {
		_, ok := state.servedUsersMap[id]
		return ok, nil
	}

	_, ok := state.servedChatsMap[id]
	return ok, nil
}

func AddServed(id int64, user ...bool) error {
	state, err := getBotState()
	if err != nil {
		return err
	}

	if len(user) > 0 && user[0] {

		if _, ok := state.servedUsersMap[id]; ok {
			return nil
		}

		state.Served.Users = append(state.Served.Users, id)
		state.servedUsersMap[id] = struct{}{}

		return updateBotState(state)
	}

	if _, ok := state.servedChatsMap[id]; ok {
		return nil
	}

	state.Served.Chats = append(state.Served.Chats, id)
	state.servedChatsMap[id] = struct{}{}

	return updateBotState(state)
}

func DeleteServed(id int64, user ...bool) error {
	state, err := getBotState()
	if err != nil {
		return err
	}

	if len(user) > 0 && user[0] {

		if _, ok := state.servedUsersMap[id]; !ok {
			return nil
		}

		delete(state.servedUsersMap, id)

		newSlice := make([]int64, 0, len(state.Served.Users))
		for _, v := range state.Served.Users {
			if v != id {
				newSlice = append(newSlice, v)
			}
		}

		state.Served.Users = newSlice
		return updateBotState(state)
	}

	if _, ok := state.servedChatsMap[id]; !ok {
		return nil
	}

	delete(state.servedChatsMap, id)

	newSlice := make([]int64, 0, len(state.Served.Chats))
	for _, v := range state.Served.Chats {
		if v != id {
			newSlice = append(newSlice, v)
		}
	}

	state.Served.Chats = newSlice

	return updateBotState(state)
}
