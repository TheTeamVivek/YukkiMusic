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

package core

import (
	"fmt"
	"sync"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/ubot"
)

type Assistant struct {
	Index  int
	Client *telegram.Client
	User   *telegram.UserObj
	Ntg    *ubot.Context
}

type AssistantManager struct {
	list       []*Assistant
	cacheMu    sync.RWMutex
	indexCache map[int64]int // chatID -> assistantIndex (1-based)
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
		gologging.ErrorF(
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
	if AssistantIndexFunc == nil {
		return nil, fmt.Errorf("AssistantIndexFunc is not set")
	}

	m.cacheMu.RLock()
	if idx, ok := m.indexCache[chatID]; ok {
		m.cacheMu.RUnlock()
		return m.Get(idx)
	}
	m.cacheMu.RUnlock()

	idx1, err := AssistantIndexFunc(chatID, len(m.list))
	if err != nil {
		return nil, err
	}

	m.cacheMu.Lock()
	if m.indexCache == nil {
		m.indexCache = make(map[int64]int)
	}
	m.indexCache[chatID] = idx1
	m.cacheMu.Unlock()

	return m.Get(idx1)
}
