/*
 * ● YukkiMusic
 * ○ A high-performance engine for streaming music in Telegram voicechats.
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
 */
 
 package database

func BlacklistedUsers() ([]int64, error) {
	state, err := getBotState()
	if err != nil {
		return nil, err
	}
	return append([]int64(nil), state.Blacklisted.Users...), nil
}

func BlacklistedChats() ([]int64, error) {
	state, err := getBotState()
	if err != nil {
		return nil, err
	}
	return append([]int64(nil), state.Blacklisted.Chats...), nil
}

func IsBlacklistedUser(userID int64) (bool, error) {
	state, err := getBotState()
	if err != nil {
		return false, err
	}
	return contains(state.Blacklisted.Users, userID), nil
}

func IsBlacklistedChat(chatID int64) (bool, error) {
	state, err := getBotState()
	if err != nil {
		return false, err
	}
	return contains(state.Blacklisted.Chats, chatID), nil
}

func AddBlacklistedUser(userID int64) error {
	return modifyBotState(func(s *BotState) bool {
		var added bool
		s.Blacklisted.Users, added = addUnique(s.Blacklisted.Users, userID)
		return added
	})
}

func RemoveBlacklistedUser(userID int64) error {
	return modifyBotState(func(s *BotState) bool {
		var removed bool
		s.Blacklisted.Users, removed = removeElement(s.Blacklisted.Users, userID)
		return removed
	})
}

func AddBlacklistedChat(chatID int64) error {
	return modifyBotState(func(s *BotState) bool {
		var added bool
		s.Blacklisted.Chats, added = addUnique(s.Blacklisted.Chats, chatID)
		return added
	})
}

func RemoveBlacklistedChat(chatID int64) error {
	return modifyBotState(func(s *BotState) bool {
		var removed bool
		s.Blacklisted.Chats, removed = removeElement(s.Blacklisted.Chats, chatID)
		return removed
	})
}
