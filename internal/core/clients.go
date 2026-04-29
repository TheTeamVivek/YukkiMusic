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
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/ubot"
)

var (
	Bot *telegram.Client

	Assistants            *AssistantManager
	GetAssistantIndexFunc func(chatID int64, assistantCount int) (int, error) // GetAssistantIndexFunc = database.AssistantIndex
)

// Init initializes the bot and assistant clients.
// It returns a shutdown function and an error if initialization fails.
func Init() (func(), error) {
	gologging.Info("Starting bot client...")
	if err := initBot(); err != nil {
		return nil, fmt.Errorf("bot initialization: %w", err)
	}

	gologging.Info("Starting assistant clients...")
	if err := initAssistants(); err != nil {
		return nil, fmt.Errorf("assistants initialization: %w", err)
	}

	Bot.SetCommandPrefixes("/")

	shutdown := func() {
		gologging.Info("Stopping bot...")
		Bot.Stop()

		gologging.Info("Shutting down assistants...")
		Assistants.ForEach(func(a *Assistant) {
			a.Ntg.Close()
			a.Client.Stop()
		})

		gologging.Info("Shutdown complete.")
	}

	return shutdown, nil
}

func initBot() error {
	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:   config.APIID,
		AppHash: config.APIHash,
		Logger: telegram.WrapSimpleLogger(
			GetTgLogger("gogram", telegram.LogError),
		),
		LogLevel:     telegram.LogError,
		ParseMode:    "HTML",
		Session:      "bot.session",
		FloodHandler: handleFlood,
	})
	if err != nil {
		return fmt.Errorf("failed to create bot client: %w", err)
	}

	if err := client.LoginBot(config.Token); err != nil {
		if strings.Contains(err.Error(), "ACCESS_TOKEN_EXPIRED") {
			return fmt.Errorf("bot token has been revoked or expired")
		}
		return fmt.Errorf("failed to start the bot: %w", err)
	}

	user, err := client.GetMe()
	if err != nil {
		return fmt.Errorf("failed to fetch bot identity: %w", err)
	}

                if config.LoggerID != 0 {
                        _, _ = client.SendMessage(
                                config.LoggerID,
                                "Bot Started",
                        )
                }

	gologging.InfoF("Bot started as @%s", user.Username)

	Bot = client
	return nil
}

func initAssistants() error {
	assistantList := make([]*Assistant, 0, len(config.StringSessions))

	for i, sessionStr := range config.StringSessions {
		gologging.InfoF("Initializing assistant[%d]...", i)

		assistant, err := initAssistant(sessionStr, i)
		if err != nil {
			return fmt.Errorf("failed to initialize assistant[%d]: %w", i, err)
		}

		assistantList = append(assistantList, assistant)

		if config.LoggerID != 0 {
			_, _ = assistant.Client.SendMessage(
				config.LoggerID,
				fmt.Sprintf("Assistant %d Started", i+1),
			)
		}

		m , _ := assistant.Client.SendMessage(Bot.Me().Username, "/start")
if m != nil {
_,_ = m.Delete()
}
		assistant.Client.JoinChannel("TheTeamVivek")

		if assistant.Self.Username != "" {
			gologging.InfoF(
				"Assistant[%d] started as @%s",
				i,
				assistant.Self.Username,
			)
		} else {
			gologging.InfoF("Assistant[%d] started as %s", i, assistant.Self.FirstName)
		}
	}

	Assistants = &AssistantManager{
		list:       assistantList,
		indexCache: make(map[int64]int),
	}
	return nil
}

func initAssistant(
	sessionStr string,
	index int,
) (*Assistant, error) {
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
		Session:       fmt.Sprintf("ass_%d.session", index),
	})
	if err != nil {
		return nil, fmt.Errorf("creating client: %w", err)
	}

	user, err := client.GetMe()
	if err != nil {
		return nil, fmt.Errorf("fetching identity: %w", err)
	}

	client.SetCommandPrefixes(".")

	return &Assistant{
		Index:  index,
		Client: client,
		Self:   user,
		Ntg:    ubot.NewContext(client),
	}, nil
}

func handleFlood(err error) bool {
	wait := telegram.GetFloodWait(err)
	if wait <= 0 {
		return false
	}

	if wait > 10 {
		gologging.WarnF("Flood wait too long, skipping sleep %d seconds", wait)
		return false
	}

	gologging.WarnF("Flood wait detected, sleeping %d seconds", wait)
	time.Sleep(time.Duration(wait) * time.Second)
	return true
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
	// SESSION_STRING_FORMAT: Big-endian, uint8, uint32, bool, 256-byte array, uint64, bool
	const (
		dcIDSize     = 1 // uint8
		apiIDSize    = 4 // uint32
		testModeSize = 1 // bool (uint8)
		authKeySize  = 256
		userIDSize   = 8 // uint64
		isBotSize    = 1 // bool (uint8)
	)

	// Add padding to the base64 string if necessary
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

	// ">B{}sH256s"
	offset := 1

	// IP Address (4 or 16 bytes based on IPv4 or IPv6)
	ipData := data[offset : offset+ipLen]
	ip := net.IP(ipData)
	ipAddress := ip.String()
	offset += ipLen

	// Port (2 bytes, Big Endian)
	port := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	// Auth Key (256 bytes)
	var authKey [256]byte
	copy(authKey[:], data[offset:offset+256])

	return &telegram.Session{
		Hostname: ipAddress + ":" + fmt.Sprint(port),
		Key:      authKey[:],
	}, nil
}
