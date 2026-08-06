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

package config

import (
	"errors"
	"math/rand/v2"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"yukkimusic/internal/logger"
)

var (
	// Required
	APIID          = int32(getEnvInt64("API_ID"))
	APIHash        = getEnv("API_HASH")
	Token          = getEnv("TOKEN")
	MongoURI       = getEnv("MONGO_DB_URI")
	StringSessions = getEnvStrings("STRING_SESSIONS")

	// Optional
	SessionType         = getEnv("SESSION_TYPE", "pyrogram")
	LoggerID            = getEnvInt64("LOGGER_ID", 0)
	OwnerID             = getEnvInt64("OWNER_ID", 0)
	DisableColour       = getEnvBool("DISABLE_COLOUR", false)
	SpotifyClientID     = getEnv("SPOTIFY_CLIENT_ID")
	SpotifyClientSecret = getEnv("SPOTIFY_CLIENT_SECRET")
	FallenAPIURL        = getEnv("FALLEN_API_URL", "https://beta.fallenapi.fun")
	FallenAPIKey        = getEnv("FALLEN_API_KEY")

	DefaultLang    = getEnv("DEFAULT_LANG", "en")
	DurationLimit  = getEnvInt("DURATION_LIMIT", 4200)
	LeaveOnDemoted = getEnvBool("LEAVE_ON_DEMOTED", false)
	QueueLimit     = getEnvInt("QUEUE_LIMIT", 24)

	SupportChat    = getEnv("SUPPORT_CHAT", "https://t.me/TheTeamVk")
	SupportChannel = getEnv("SUPPORT_CHANNEL", "https://t.me/TheTeamVivek")
	CookiesLink    = getEnv("COOKIES_LINK", "")
	SetCmds        = getEnvBool("SET_CMDS", false)
	MaxAuthUsers   = getEnvInt("MAX_AUTH_USERS", 25)

	StartImages = getEnvStrings("START_IMAGES")

	PingImage = getEnv(
		"PING_IMG_URL",
		"https://telegra.ph/file/91533956c91d0fd7c9f20.jpg",
	)

	Port = getEnv("PORT", "8000")

	StartTime = time.Now()

	logr = logger.GetLogger("config")
)

func Load() error {
	// Legacy compatibility
	if Token == "" {
		Token = getEnv("BOT_TOKEN")
	}

	if LoggerID == 0 {
		LoggerID = getEnvInt64("LOG_GROUP_ID")
	}

	if len(StringSessions) == 0 {
		StringSessions = getEnvStrings("STRING_SESSION")
	}

	if len(StartImages) == 0 {
		if img := getEnv("START_IMG_URL"); img != "" {
			StartImages = []string{img}
		}
	}

	if err := validateConfig(); err != nil {
		return err
	}

	return nil
}

func validateConfig() error {
	checks := []struct {
		ok  bool
		msg string
	}{
		{APIID != 0, "API_ID is required"},
		{APIHash != "", "API_HASH is required"},
		{Token != "", "TOKEN (or BOT_TOKEN) is required"},
		{MongoURI != "", "MONGO_DB_URI is required"},
		{len(StringSessions) > 0, "STRING_SESSIONS (or STRING_SESSION) is required"},
	}

	for _, check := range checks {
		if !check.ok {
			return errors.New(check.msg)
		}
	}

	if SpotifyClientID == "" || SpotifyClientSecret == "" {
		logr.Warn("Spotify credentials not configured; Spotify links will be unavailable")
	}

	return nil
}

func StartImage() string {
	if len(StartImages) == 0 {
		return ""
	}

	return StartImages[rand.IntN(len(StartImages))]
}
