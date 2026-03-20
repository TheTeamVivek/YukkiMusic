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

func ServedChats() ([]int64, error) {
	state, err := getBotState()
	if err != nil {
		return nil, err
	}
	return state.Served.Chats, nil
}

func ServedUsers() ([]int64, error) {
	state, err := getBotState()
	if err != nil {
		return nil, err
	}
	return state.Served.Users, nil
}

func IsServedChat(id int64) (bool, error) {
	state, err := getBotState()
	if err != nil {
		return false, err
	}
	_, ok := state.servedChatsMap[id]
	return ok, nil
}

func IsServedUser(id int64) (bool, error) {
	state, err := getBotState()
	if err != nil {
		return false, err
	}
	_, ok := state.servedUsersMap[id]
	return ok, nil
}

func AddServedChat(id int64) error {
	return modifyBotState(func(s *BotState) bool {
		var added bool
		s.Served.Chats, added = addUnique(s.Served.Chats, id)
		if added {
			s.servedChatsMap[id] = struct{}{}
		}
		return added
	})
}

func AddServedUser(id int64) error {
	return modifyBotState(func(s *BotState) bool {
		var added bool
		s.Served.Users, added = addUnique(s.Served.Users, id)
		if added {
			s.servedUsersMap[id] = struct{}{}
		}
		return added
	})
}

func RemoveServedChat(id int64) error {
	return modifyBotState(func(s *BotState) bool {
		var removed bool
		s.Served.Chats, removed = removeElement(s.Served.Chats, id)
		if removed {
			delete(s.servedChatsMap, id)
		}
		return removed
	})
}

func RemoveServedUser(id int64) error {
	return modifyBotState(func(s *BotState) bool {
		var removed bool
		s.Served.Users, removed = removeElement(s.Served.Users, id)
		if removed {
			delete(s.servedUsersMap, id)
		}
		return removed
	})
}
