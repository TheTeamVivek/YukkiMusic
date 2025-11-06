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
package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	logger = gologging.GetLogger("config")

	// To learn more about what each variable does, see README.md
	// Required Vars
	ApiID         = int32(getInt64("API_ID"))
	ApiHash       = getString("API_HASH")
	Token         = getString("TOKEN")
	MongoURI      = getString("MONGO_DB_URI")
	StringSession = getString("STRING_SESSION") // pyro session

	// Optional Vars
	ApiKEY         = getString("FALLEN_API_KEY")
	ApiURL         = getString("FALLEN_API_URL", "https://tgmusic.fallenapi.fun")
	OwnerID        = getInt64("OWNER_ID")
	LoggerID       = getInt64("LOGGER_ID")
	DefaultLang    = getString("DEFAULT_LANG", "en")
	DurationLimit  = int(getInt64("DURATION_LIMIT", 4200)) // in seconds
	QueueLimit     = int(getInt64("QUEUE_LIMIT", 7))
	SupportChat    = getString("SUPPORT_CHAT", "https://t.me/TheTeamVk")
	SupportChannel = getString("SUPPORT_CHANNEL", "https://t.me/TheTeamVivek")
	StartTime      = time.Now()
	CookiesLink    = getString("COOKIES_LINK")
	SetCmds        = getBool("SET_CMDS", false)
	MaxAuthUsers   = int(getInt64("MAX_AUTH_USERS", 25))

	StartImage = getString("START_IMG_URL", "https://raw.githubusercontent.com/Vivekkumar-IN/assets/master/images.png")
	PingImage  = getString("PING_IMG_URL", "https://telegra.ph/file/91533956c91d0fd7c9f20.jpg")
)

func init() {
	if Token == "" {
		Token = getString("BOT_TOKEN")
		if Token == "" {
			logger.Fatal("TOKEN is required but missing! Please set it in .env or environment.")
			return
		}
	}
	if MongoURI == "" {
		logger.Fatal("MONGO_DB_URI is required but missing!")
		return
	}
	if StringSession == "" {
		logger.Fatal("STRING_SESSION is empty — continuing without it.")
		return
	}
	if ApiID == 0 {
		logger.Fatal("API_ID is required but missing!")
		return
	}
	if ApiHash == "" {
		logger.Fatal("API_HASH is required but missing!")
		return
	}
}

func getString(key string, def ...string) string {
	if val, ok := getEnvAny(variants(key)...); ok {
		return val
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

func getBool(key string, def ...bool) bool {
	val, ok := getEnvAny(variants(key)...)
	defaultValue := len(def) > 0 && def[0]

	if ok {
		boolVal, err := strconv.ParseBool(val)
		if err != nil {
			logger.FatalF("Invalid boolean for %s: %v — using default %t", key, err, defaultValue)
			return defaultValue
		}
		return boolVal
	}
	return defaultValue
}

func getInt64(key string, def ...int64) int64 {
	defaultValue := int64(0)
	if len(def) > 0 {
		defaultValue = def[0]
	}

	if val, ok := getEnvAny(variants(key)...); ok {
		num, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			logger.FatalF("Invalid int64 for %s: %v — using default %d", key, err, defaultValue)
			return defaultValue
		}
		return num
	}
	return defaultValue
}

func getEnvAny(keys ...string) (string, bool) {
	for _, key := range keys {
		if val, ok := os.LookupEnv(key); ok {
			val = strings.TrimSpace(val)
			if val != "" {
				return val, true
			}
		}
	}
	return "", false
}

func variants(base string) []string {
	return []string{
		base,
		strings.ToUpper(base),
		strings.ToLower(base),
		strings.ReplaceAll(base, "_", ""),
		cases.Title(language.Und, cases.NoLower).String(strings.ReplaceAll(base, "_", " ")),
	}
}
