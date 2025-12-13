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
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/ubot"
)

var (
	Bot   *telegram.Client
	BUser *telegram.UserObj

	Assistants         *AssistantManager
	AssistantIndexFunc func(chatID int64, assistantCount int) (int, error) // AssistantIndexFunc = database.GetAssistantIndex
)

func Init(apiID int32, apiHash, token string, sessions []string, sessionType string, loggerID int64) func() {
	if len(sessions) == 0 {
		gologging.Fatal("No STRING_SESSIONS provided for assistant client.")
	}

	gologging.Info("Starting bot client...")
	Bot = initBotClient(apiID, apiHash, token)
	BUser = getSelfOrFatal(Bot, "bot")

	gologging.Info("Starting assistant clients...")

	assistants := make([]*Assistant, 0, len(sessions))

	for i, sess := range sessions {
		gologging.InfoF("Initializing assistant[%d]...", i)

		client := initAssistantClient(apiID, apiHash, sess, sessionType, i)
		user := getSelfOrFatal(client, fmt.Sprintf("assistant[%d]", i))
		ctx := ubot.NewContext(client)

		client.SetCommandPrefixes(".")

		assistants = append(assistants, &Assistant{
			Index:  i,
			Client: client,
			User:   user,
			Ntg:    ctx,
		})

		if loggerID != 0 {
			_, _ = client.SendMessage(loggerID, fmt.Sprintf("Assistant %d Started", i+1))
		}

		gologging.InfoF("assistant[%d] ready: %s", i, user.FirstName)
	}

	Bot.SetCommandPrefixes("/")

	Assistants = &AssistantManager{
		list:       assistants,
		indexCache: make(map[int64]int),
	}
	gologging.Info("All assistants initialized successfully.")

	return func() {
		gologging.Info("Shutting down assistant contexts...")
		for _, a := range Assistants.list {
			a.Ntg.Close()
		}

		gologging.Info("Stopping bot...")
		Bot.Stop()

		gologging.Info("Stopping assistants...")
		for _, a := range Assistants.list {
			a.Client.Stop()
		}

		gologging.Info("Shutdown complete.")
	}
}

func initBotClient(apiID int32, apiHash, token string) *telegram.Client {
	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:     apiID,
		AppHash:   apiHash,
		Logger:    telegram.WrapSimpleLogger(GetTgLogger("gogram", telegram.LogError)),
		LogLevel:  telegram.LogError,
		ParseMode: "HTML",
		Session:   "bot.session",
	})
	if err != nil {
		gologging.Fatal("❌ Failed to create bot: " + err.Error())
	}

	if err := client.LoginBot(token); err != nil {
		if strings.Contains(err.Error(), "ACCESS_TOKEN_EXPIRED") {
			gologging.Fatal("❌ Bot token has been revoked or expired.")
		} else {
			gologging.Fatal("❌ Failed to start the bot: " + err.Error())
		}
	}
	return client
}

func initAssistantClient(apiID int32, apiHash, session, sessionType string, idx int) *telegram.Client {
	var stringSession string

	switch strings.ToLower(sessionType) {
	case "pyrogram", "pyro":
		sess, err := decodePyrogramSessionString(session)
		if err != nil {
			gologging.Fatal("Failed to decode Pyrogram session: " + err.Error())
		}
		stringSession = sess.Encode()

	case "telethon":
		sess, err := decodeTelethonSessionString(session)
		if err != nil {
			gologging.Fatal("Failed to decode Telethon session: " + err.Error())
		}
		stringSession = sess.Encode()

	case "gogram":
		stringSession = session

	default:
		gologging.Fatal("Invalid SESSION_TYPE: " + sessionType)
	}

	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:         apiID,
		AppHash:       apiHash,
		LogLevel:      telegram.LogError,
		ParseMode:     "HTML",
		StringSession: stringSession,
		Session:       fmt.Sprintf("ass%d.session", idx),
	})
	if err != nil {
		gologging.Fatal("Failed to create assistant: " + err.Error())
	}

	return client
}

func getSelfOrFatal(c *telegram.Client, label string) *telegram.UserObj {
	me, err := c.GetMe()
	if err != nil {
		gologging.Fatal("❌ Failed to GetMe for " + label + ": " + err.Error())
	}
	gologging.Info("Logged in as " + label + ": " + me.FirstName)
	return me
}

func decodePyrogramSessionString(encodedString string) (*telegram.Session, error) {
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
		return nil, fmt.Errorf("unexpected data length: got %d, want %d", len(packedData), expectedSize)
	}

	return &telegram.Session{
		Hostname: telegram.ResolveDataCenterIP(int(uint8(packedData[0])), packedData[5] != 0, false),
		AppID:    int32(uint32(packedData[1])<<24 | uint32(packedData[2])<<16 | uint32(packedData[3])<<8 | uint32(packedData[4])),
		Key:      packedData[6 : 6+authKeySize],
	}, nil
}

func decodeTelethonSessionString(sessionString string) (*telegram.Session, error) {
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
