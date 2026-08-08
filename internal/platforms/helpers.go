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
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	td "github.com/AshokShau/gotdbot"
	"yukkimusic/internal/logger"

	state "yukkimusic/internal/core/models"
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
		logger.Debug("fileExists: " + path + " not found")
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

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func playableMedia(c *td.Client, m *td.Message) (isVideo, isAudio bool) {
	if m == nil {
		return
	}
	check := func(msg *td.Message) (bool, bool) {
		if msg == nil || msg.Content == nil {
			return false, false
		}
		switch content := msg.Content.(type) {
		case *td.MessageAudio, *td.MessageVoiceNote:
			return false, true
		case *td.MessageVideo, *td.MessageAnimation, *td.MessageVideoNote:
			return true, false
		case *td.MessageDocument:
			if content.Document == nil {
				return false, false
			}
			mt := strings.ToLower(content.Document.MimeType)
			switch {
			case strings.HasPrefix(mt, "audio/"):
				return false, true
			case strings.HasPrefix(mt, "video/"):
				return true, false
			}
		}
		return false, false
	}
	curr := m
	for curr != nil {
		if v, a := check(curr); v || a {
			return v, a
		}
		if curr.ReplyToMessageID() <= 0 {
			break
		}
		next, err := curr.GetRepliedMessage(c)
		if err != nil {
			break
		}
		curr = next
	}
	return false, false
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
