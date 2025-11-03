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
	"fmt"
	"strings"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/ubot"
)

var (
	Bot  *telegram.Client // bot client
	UBot *telegram.Client // user client
	Ntg  *ubot.Context    // wrapper client of ntgcalls

	BUser, UbUser *telegram.UserObj
)

func Init(apiID int32, apiHash, token, session string, loggerID int64) func() {
	Bot = initBotClient(apiID, apiHash, token)
	BUser = getSelfOrFatal(Bot, "bot")

	UBot = initAssistantClient(apiID, apiHash, session)
	UbUser = getSelfOrFatal(UBot, "assistant")

	if loggerID != 0 {
		notifyStartup(Bot, UBot, loggerID)
	}

	Ntg = ubot.NewContext(UBot, UbUser)

	return func() {
		Ntg.Close()
		Bot.Stop()
		UBot.Stop()
	}
}

func initBotClient(apiID int32, apiHash, token string) *telegram.Client {
	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:     apiID,
		AppHash:   apiHash,
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

func initAssistantClient(apiID int32, apiHash, session string) *telegram.Client {
	sess, err := decodePyrogramSessionString(session)
	if err != nil {
		gologging.Fatal("❌ Failed to decode Pyrogram session: " + err.Error())
	}

	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:         apiID,
		AppHash:       apiHash,
		LogLevel:      telegram.LogError,
		ParseMode:     "HTML",
		StringSession: sess.Encode(),
		Session:       "ass.session",
	})
	if err != nil {
		gologging.Fatal("❌ Failed to create assistant: " + err.Error())
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

func notifyStartup(bot, ub *telegram.Client, loggerID int64) {
	_, err := bot.SendMessage(loggerID, "🚀 Bot Started...")
	if err != nil {
		gologging.Warn("Failed to send bot startup message: " + err.Error())
	}

	_, err = ub.SendMessage(loggerID, "🚀 Assistant Started...")
	if err != nil {
		gologging.Warn("Failed to send assistant startup message: " + err.Error())
	}
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
