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
	l := gologging.GetLogger("Clients")

	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:     apiID,
		AppHash:   apiHash,
		LogLevel:  telegram.LogError,
		ParseMode: "HTML",
		Session:   "bot.session",
	})
	if err != nil {
		l.FatalF("❌ Failed to create bot: %s", err)
		return nil
	}

	if err := client.LoginBot(token); err != nil {
		if strings.Contains(err.Error(), "ACCESS_TOKEN_EXPIRED") {
			l.FatalF("❌ Bot token has been revoked or expired.")
		} else {
			l.FatalF("❌ Failed to start the bot: %s", err)
		}
		return nil
	}

	if me, err := client.GetMe(); err != nil {
		l.FatalF("❌ Failed to GetMe: %s", err)
		return nil
	} else {
		BUser = me
	}

	sess, serr := decodePyrogramSessionString(session)
	if serr != nil {
		l.FatalF("❌ Failed to decode Pyrogram session: %s", serr)
		return nil
	}

	ub, err2 := telegram.NewClient(telegram.ClientConfig{
		AppID:         apiID,
		AppHash:       apiHash,
		LogLevel:      telegram.LogError,
		ParseMode:     "HTML",
		StringSession: sess.Encode(),

		Session: "ass.session",
	})
	if err2 != nil {
		l.FatalF("❌ Failed to create ubot: %s", err2)
		return nil
	}

	if me, err := ub.GetMe(); err != nil {
		l.FatalF("❌ Failed to GetMe: %s", err)
		return nil
	} else {
		UbUser = me
		l.InfoF("Logged in as: %s", me.FirstName)
	}
	if peer, err := client.ResolvePeer(loggerID); err != nil {
		l.WarnF("Failed to get peer ID of logger: %s", err)
	} else if _, err := client.SendMessage(peer, "🚀 Bot Started..."); err != nil {
		l.WarnF("Failed to send startup message: %s", err)
	}

	if peer, err := ub.ResolvePeer(loggerID); err != nil {
		l.WarnF("Failed to get peer ID of logger (assistant): %s", err)
	} else if _, err := ub.SendMessage(peer, "🚀 Assistant Started..."); err != nil {
		l.WarnF("Failed to send assistant startup message: %s", err)
	}

	ub.SendMessage(BUser.Username, "/start")
	Bot = client
	UBot = ub
	Ntg = ubot.NewContext(ub, UbUser)

	return func() {
		if Ntg != nil {
			Ntg.Close()
			Bot.Stop()
			UBot.Stop()
		}
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
