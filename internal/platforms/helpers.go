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

package platforms

import (
	"errors"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	state "main/internal/core/models"
)

var errUnsafeURL = errors.New("invalid or unsafe url")

func getPath(track *state.Track, ext string) string {
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	t := "audio"
	if track.Video {
		t = "video"
	}
	return filepath.Join("downloads", t+"_"+track.ID+ext)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		gologging.Debug("fileExists: " + path + " not found")
		return false
	}
	return info.Size() > 0
}

func findFile(track *state.Track) string {
	t := "audio"
	if track.Video {
		t = "video"
	}
	files, err := filepath.Glob(filepath.Join("downloads", t+"_"+track.ID+"*"))
	if err != nil {
		return ""
	}
	for _, f := range files {
		if info, err := os.Stat(f); err == nil && info.Size() > 0 {
			return f
		}
	}
	return ""
}

func findAndRemove(track *state.Track) {
	t := "audio"
	if track.Video {
		t = "video"
	}
	files, err := filepath.Glob(filepath.Join("downloads", t+"_"+track.ID+"*"))
	if err != nil {
		return
	}
	for _, f := range files {
		os.Remove(f)
	}
}

func sanitizeAPIError(err error, apiKey string) error {
	if err == nil || apiKey == "" {
		return err
	}
	return errors.New(strings.ReplaceAll(err.Error(), apiKey, "***REDACTED***"))
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func playableMedia(m *telegram.NewMessage) (isVideo, isAudio bool) {
	if m == nil {
		return
	}
	check := func(msg *telegram.NewMessage) (bool, bool) {
		switch {
		case msg.Audio() != nil, msg.Voice() != nil:
			return false, true
		case msg.Video() != nil:
			return true, false
		case msg.Document() != nil:
			mt := strings.ToLower(msg.Document().MimeType)
			switch {
			case strings.HasPrefix(mt, "audio/"):
				return false, true
			case strings.HasPrefix(mt, "video/"):
				return true, false
			}
		}
		return false, false
	}
	if m.IsReply() {
		rmsg, err := m.GetReplyMessage()
		if err != nil {
			return
		}
		return check(rmsg)
	}
	return check(m)
}

func sanitizeMediaURL(raw string) (string, error) {
	u := strings.TrimSpace(raw)
	if u == "" {
		return "", errUnsafeURL
	}

	for _, r := range u {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return "", errUnsafeURL
		}
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return "", errUnsafeURL
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errUnsafeURL
	}

	host := parsed.Hostname()
	if host == "" || parsed.User != nil {
		return "", errUnsafeURL
	}

	if strings.EqualFold(host, "localhost") {
		return "", errUnsafeURL
	}

	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() ||
			ip.IsLinkLocalMulticast() || ip.IsPrivate() || ip.IsUnspecified() {
			return "", errUnsafeURL
		}
	}

	return parsed.String(), nil
}
