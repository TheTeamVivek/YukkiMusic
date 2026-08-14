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

package core

import (
	"fmt"

	"yukkimusic/internal/database"
	"yukkimusic/internal/logger"

	"github.com/amarnathcjd/gogram/telegram"

	"yukkimusic/ubot"
)

type Assistant struct {
	Index  int
	Client *telegram.Client
	Self   *telegram.UserObj
	Ntg    *ubot.Context
}

type AssistantManager struct {
	list []*Assistant
}

func (m *AssistantManager) Count() int {
	if m == nil {
		return 0
	}
	return len(m.list)
}

func (m *AssistantManager) Get(idx int) (*Assistant, error) {
	if m == nil {
		return nil, fmt.Errorf("assistant manager not initialized")
	}
	if idx < 1 || idx > len(m.list) {
		return nil, fmt.Errorf("assistant index out of range: %d", idx)
	}
	return m.list[idx-1], nil
}

func (m *AssistantManager) First() (*Assistant, error) {
	return m.Get(1)
}

func (m *AssistantManager) ForEach(fn func(*Assistant)) {
	if m == nil {
		return
	}
	for _, a := range m.list {
		fn(a)
	}
}

func (m *AssistantManager) WithAssistant(chatID int64, fn func(*Assistant)) {
	if m == nil {
		return
	}

	ass, err := m.ForChat(chatID)
	if err != nil {
		logger.Errorf(
			"Failed to get assistant for chat %d, Error: %v",
			chatID,
			err,
		)
		return
	}

	fn(ass)
}

func (m *AssistantManager) ForChat(chatID int64) (*Assistant, error) {
	if m == nil || len(m.list) == 0 {
		return nil, fmt.Errorf("no assistants available")
	}

	idx, err := database.GetAssistant(chatID)
	if err != nil {
		return nil, err
	}

	return m.Get(idx)
}
