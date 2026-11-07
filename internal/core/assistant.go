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
	"crypto/sha1"
	"fmt"
	"net"
	"strconv"
	"strings"

	"yukkimusic/internal/database"
	"yukkimusic/internal/utils"

	tgconv "github.com/mtgo-labs/session-converter"
	"github.com/xAmarnath/gogram/telegram"

	"yukkimusic/config"
	"yukkimusic/ubot"
)

type Assistant struct {
	Index  int // 1-based, as stored in the database
	Client *telegram.Client
	Self   *telegram.UserObj
	Ntg    *ubot.Context
}

// Assistants is the shared manager holding every assistant client. It is
// initialized once during startup in clients.go.
var Assistants *ClientManager

type ClientManager struct {
	list []*Assistant
}

// AddClient creates, connects and registers an assistant from the given
// session string. The assistant is appended to the manager and returned.
func (m *ClientManager) AddClient(sessionStr string) (*Assistant, error) {
	stringSession, err := resolveSession(sessionStr)
	if err != nil {
		return nil, fmt.Errorf("resolving session: %w", err)
	}

	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:         config.APIID,
		AppHash:       config.APIHash,
		LogLevel:      telegram.LogError,
		ParseMode:     "HTML",
		StringSession: stringSession,
		MemorySession: true,
	})
	if err != nil {
		return nil, fmt.Errorf("creating client: %w", err)
	}

	user, err := client.GetMe()
	if err != nil {
		return nil, fmt.Errorf("fetching identity: %w", err)
	}

	ntg, err := ubot.NewContext(client)
	if err != nil {
		return nil, fmt.Errorf("creating ubot context: %w", err)
	}

	assistant := &Assistant{
		Index:  len(m.list) + 1,
		Client: client,
		Self:   user,
		Ntg:    ntg,
	}

	m.list = append(m.list, assistant)

	return assistant, nil
}

func (m *ClientManager) Count() int {
	if m == nil {
		return 0
	}
	return len(m.list)
}

func (m *ClientManager) Get(idx int) (*Assistant, error) {
	if m == nil {
		return nil, fmt.Errorf("assistant manager not initialized")
	}
	if idx < 1 || idx > len(m.list) {
		return nil, fmt.Errorf("assistant index out of range: %d", idx)
	}
	return m.list[idx-1], nil
}

func (m *ClientManager) ForEach(fn func(*Assistant)) {
	if m == nil {
		return
	}
	for _, a := range m.list {
		fn(a)
	}
}

func (m *ClientManager) WithAssistant(chatID int64, fn func(*Assistant)) {
	if m == nil {
		return
	}

	ass, err := m.ForChat(chatID)
	if err != nil {
		utils.Log.Errorf(
			"Failed to get assistant for chat %d, Error: %v",
			chatID,
			err,
		)
		return
	}

	fn(ass)
}

func (m *ClientManager) ForChat(chatID int64) (*Assistant, error) {
	if m == nil || len(m.list) == 0 {
		return nil, fmt.Errorf("no assistants available")
	}

	idx, err := database.GetAssistant(chatID)
	if err != nil {
		return nil, err
	}

	return m.Get(idx)
}

// resolveSession auto-detects and converts ANY session string format
// (Telethon, Pyrogram, GramJS, mtcute, MTKruto, gotgproto, mtgo, gogram)
// into a standardized gogram-compatible session string.
func resolveSession(session string) (string, error) {
	trimmed := strings.TrimSpace(session)
	if trimmed == "" {
		return "", fmt.Errorf("empty session string")
	}

	sess, _, err := tgconv.Decode(trimmed)
	if err != nil {
		return "", fmt.Errorf("decoding session string: %w", err)
	}

	sess.AppID = config.APIID
	sess.FillDefaults()

	host := sess.ServerAddress
	port := sess.Port
	if port == 0 {
		port = 443
	}
	if strings.Contains(host, ":") {
		hPart, pPart, err := net.SplitHostPort(host)
		if err == nil {
			host = hPart
			if p, err := strconv.Atoi(pPart); err == nil {
				port = p
			}
		}
	}
	if host == "0.0.0.0" || host == "" {
		host = "149.154.175.50"
	}

	h := sha1.Sum(sess.AuthKey)
	hash := h[12:]

	s := &telegram.Session{
		Hostname: fmt.Sprintf("%s:%d", host, port),
		AppID:    config.APIID,
		Key:      sess.AuthKey,
		Hash:     hash,
	}

	return s.Encode(), nil
}
