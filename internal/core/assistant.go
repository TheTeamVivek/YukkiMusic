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
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	"yukkimusic/internal/database"
	"yukkimusic/internal/utils"

	"github.com/amarnathcjd/gogram/telegram"

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

func resolveSession(session string) (string, error) {
	switch strings.ToLower(config.SessionType) {
	case "pyrogram", "pyro":
		sess, err := decodePyrogramSessionString(session)
		if err != nil {
			return "", fmt.Errorf("decoding Pyrogram session: %w", err)
		}
		return sess.Encode(), nil

	case "telethon":
		sess, err := decodeTelethonSessionString(session)
		if err != nil {
			return "", fmt.Errorf("decoding Telethon session: %w", err)
		}
		return sess.Encode(), nil

	case "gogram":
		return session, nil

	default:
		return "", fmt.Errorf("invalid SESSION_TYPE: %s", config.SessionType)
	}
}

func decodePyrogramSessionString(
	encodedString string,
) (*telegram.Session, error) {
	const (
		dcIDSize     = 1
		apiIDSize    = 4
		testModeSize = 1
		authKeySize  = 256
		userIDSize   = 8
		isBotSize    = 1
	)

	for len(encodedString)%4 != 0 {
		encodedString += "="
	}

	packedData, err := base64.URLEncoding.DecodeString(encodedString)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 string: %w", err)
	}

	expectedSize := dcIDSize + apiIDSize + testModeSize + authKeySize + userIDSize + isBotSize
	if len(packedData) != expectedSize {
		return nil, fmt.Errorf(
			"unexpected data length: got %d, want %d",
			len(packedData),
			expectedSize,
		)
	}

	return &telegram.Session{
		Hostname: telegram.ResolveDC(
			int(uint8(packedData[0])),
			packedData[5] != 0,
			false,
		),
		AppID: int32(
			uint32(
				packedData[1],
			)<<24 | uint32(
				packedData[2],
			)<<16 | uint32(
				packedData[3],
			)<<8 | uint32(
				packedData[4],
			),
		),
		Key: packedData[6 : 6+authKeySize],
	}, nil
}

func decodeTelethonSessionString(
	sessionString string,
) (*telegram.Session, error) {
	data, err := base64.URLEncoding.DecodeString(sessionString[1:])
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %v", err)
	}

	ipLen := 4
	if len(data) == 352 {
		ipLen = 16
	}

	expectedLen := 1 + ipLen + 2 + 256
	if len(data) != expectedLen {
		return nil, fmt.Errorf("invalid session string length")
	}

	offset := 1

	ipData := data[offset : offset+ipLen]
	ip := net.IP(ipData)
	ipAddress := ip.String()
	offset += ipLen

	port := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	var authKey [256]byte
	copy(authKey[:], data[offset:offset+256])

	return &telegram.Session{
		Hostname: ipAddress + ":" + fmt.Sprint(port),
		Key:      authKey[:],
	}, nil
}
