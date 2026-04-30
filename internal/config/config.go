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
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/joho/godotenv"
)

var (
	APIID          int32
	APIHash        string
	Token          string
	LoggerID       int64
	MongoURI       string
	StringSessions []string
	SessionType    string

	DisableColour       bool
	OwnerID             int64
	SpotifyClientID     string
	SpotifyClientSecret string
	FallenAPIURL        string
	FallenAPIKey        string
	DefaultLang         string
	DurationLimit       int
	LeaveOnDemoted      bool
	QueueLimit          int
	SupportChat         string
	SupportChannel      string
	CookiesLink         string
	SetCmds             bool
	MaxAuthUsers        int
	StartImage          string
	PingImage           string
	Port                string
	EnablePprof         bool

	StartTime   time.Time
	LogFileName = "logs.txt"
	LogWriter   io.Writer

	logger  = gologging.GetLogger("config")
	logFile *os.File
)

func Load() (func(), error) {
	godotenv.Load()
	if err := initLogging(); err != nil {
		return nil, fmt.Errorf("config: logging init failed: %w", err)
	}

	loadConfig()

	if err := validateConfig(); err != nil {
		closeLogging()
		return nil, fmt.Errorf("config: validation failed: %w", err)
	}

	return closeLogging, nil
}

func initLogging() error {
	_ = os.Remove(LogFileName)

	file, err := os.OpenFile(
		LogFileName,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0o644,
	)
	if err != nil {
		return fmt.Errorf("open log file %q: %w", LogFileName, err)
	}

	logFile = file
	LogWriter = io.MultiWriter(file, os.Stderr)

	return nil
}

func loadConfig() {
	StartTime = time.Now()

	APIID = int32(getInt64("API_ID", 0))
	APIHash = getString("API_HASH", "")
	Token = getString("TOKEN", getString("BOT_TOKEN", ""))
	LoggerID = getInt64("LOGGER_ID", getInt64("LOG_GROUP_ID", 0))
	MongoURI = getString("MONGO_DB_URI", "")
	SessionType = getString("SESSION_TYPE", "pyrogram")
	StringSessions = getStringSlice(
		"STRING_SESSIONS",
		getStringSlice("STRING_SESSION", nil),
	)

	DisableColour = getBool("DISABLE_COLOUR", false)
	OwnerID = getInt64("OWNER_ID", 0)
	SpotifyClientID = getString("SPOTIFY_CLIENT_ID", "")
	SpotifyClientSecret = getString("SPOTIFY_CLIENT_SECRET", "")
	FallenAPIURL = getString("FALLEN_API_URL", "https://beta.fallenapi.fun")
	FallenAPIKey = getString("FALLEN_API_KEY", "")
	DefaultLang = getString("DEFAULT_LANG", "en")
	DurationLimit = int(getInt64("DURATION_LIMIT", 4200))
	LeaveOnDemoted = getBool("LEAVE_ON_DEMOTED", false)
	QueueLimit = int(getInt64("QUEUE_LIMIT", 24))
	SupportChat = getString("SUPPORT_CHAT", "https://t.me/TheTeamVk")
	SupportChannel = getString("SUPPORT_CHANNEL", "https://t.me/TheTeamVivek")
	CookiesLink = getString("COOKIES_LINK", "")
	SetCmds = getBool("SET_CMDS", false)
	MaxAuthUsers = int(getInt64("MAX_AUTH_USERS", 25))
	StartImage = getString("START_IMG_URL", "")
	PingImage = getString(
		"PING_IMG_URL",
		"https://telegra.ph/file/91533956c91d0fd7c9f20.jpg",
	)
	Port = getString("PORT", "8000")
	EnablePprof = getBool("ENABLE_PPROF", false)
}

func validateConfig() error {
	type check struct {
		ok  bool
		msg string
	}

	required := []check{
		{APIID != 0, "API_ID is required but missing"},
		{APIHash != "", "API_HASH is required but missing"},
		{MongoURI != "", "MONGO_DB_URI is required but missing"},
		{Token != "", "TOKEN (or BOT_TOKEN) is required but missing"},
		{
			len(StringSessions) > 0,
			fmt.Sprintf(
				"STRING_SESSIONS is empty — at least one %s session string is required",
				SessionType,
			),
		},
	}

	for _, c := range required {
		if !c.ok {
			return errors.New(c.msg)
		}
	}

	if SpotifyClientID == "" || SpotifyClientSecret == "" {
		logger.Warn("Spotify credentials not configured — Spotify links won't work")
	}

	return nil
}

func closeLogging() {
	if logFile != nil {
		_ = logFile.Close()
	}
}

func lookupEnv(baseKey string) (string, bool) {
	variants := []string{
		baseKey,
		strings.ToUpper(baseKey),
		strings.ToLower(baseKey),
		strings.ReplaceAll(baseKey, "_", ""),
	}

	for _, key := range variants {
		if val, ok := os.LookupEnv(key); ok {
			if val = strings.TrimSpace(val); val != "" {
				return val, true
			}
		}
	}

	return "", false
}

func getString(key, fallback string) string {
	if val, ok := lookupEnv(key); ok {
		return val
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	val, ok := lookupEnv(key)
	if !ok {
		return fallback
	}

	b, err := strconv.ParseBool(val)
	if err != nil {
		logger.FatalF("config: invalid boolean value for %s=%q: %v", key, val, err)
	}

	return b
}

func getInt64(key string, fallback int64) int64 {
	val, ok := lookupEnv(key)
	if !ok {
		return fallback
	}

	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		logger.FatalF("config: invalid integer value for %s=%q: %v", key, val, err)
	}

	return n
}

func getStringSlice(key string, fallback []string) []string {
	val, ok := lookupEnv(key)
	if !ok {
		return fallback
	}

	parts := strings.Fields(
		strings.NewReplacer(",", " ", ";", " ").Replace(val),
	)

	if len(parts) > 0 {
		return parts
	}

	return fallback
}
