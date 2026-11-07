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
	"context"
	"errors"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	td "github.com/AshokShau/gotdbot"

	state "yukkimusic/internal/core/models"
)

var errUnsafeURL = errors.New("invalid or unsafe url")

// OnDownloadStart is set by the modules package and runs right before a
// TDLib download starts.
var OnDownloadStart func(fileID string, statusMsg *td.Message)

var (
	audioExts = []string{".mp3", ".wav", ".flac", ".ogg", ".m4a", ".aac", ".opus"}
	videoExts = []string{".mp4", ".mkv", ".webm", ".avi", ".mov"}
)

func isCancelled(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if tdErr, ok := errors.AsType[*td.Error](err); ok {
		return strings.Contains(strings.ToLower(tdErr.Message), "cancel")
	}
	return strings.Contains(strings.ToLower(err.Error()), "cancel")
}

// canceled tells a user-cancelled download apart from a failed one.
func canceled(ctx context.Context, err error) bool {
	return ctx.Err() != nil || isCancelled(err)
}

func trackFile(track *state.Track, ext string) string {
	kind := "audio"
	if track.Video {
		kind = "video"
	}
	return filepath.Join("downloads", kind+"_"+track.ID+ext)
}

func trackFiles(track *state.Track) []string {
	files, _ := filepath.Glob(trackFile(track, ".*"))
	return files
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Size() > 0
}

func findFile(track *state.Track) string {
	for _, f := range trackFiles(track) {
		if !strings.HasSuffix(f, ".part") && fileExists(f) {
			return f
		}
	}
	return ""
}

func removeFiles(track *state.Track) {
	for _, f := range trackFiles(track) {
		_ = os.Remove(f)
	}
}

func sanitizeAPIError(err error, apiKey string) error {
	if err == nil || apiKey == "" {
		return err
	}
	return errors.New(strings.ReplaceAll(err.Error(), apiKey, "***REDACTED***"))
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
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// messageMedia reports whether msg itself carries video or audio.
func messageMedia(msg *td.Message) (video, audio bool) {
	if msg == nil {
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
		mime := strings.ToLower(content.Document.MimeType)
		ext := strings.ToLower(filepath.Ext(content.Document.FileName))
		switch {
		case strings.HasPrefix(mime, "audio/"), slices.Contains(audioExts, ext):
			return false, true
		case strings.HasPrefix(mime, "video/"), slices.Contains(videoExts, ext):
			return true, false
		}
	}
	return false, false
}

// playableMedia walks up the reply chain from m to the first audio or video message.
func playableMedia(c *td.Client, m *td.Message) (msg *td.Message, isVideo, isAudio bool) {
	for curr := m; curr != nil; {
		if v, a := messageMedia(curr); v || a {
			return curr, v, a
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
	return nil, false, false
}

func sanitizeMediaURL(raw string) (string, error) {
	u := strings.TrimSpace(raw)
	if u == "" || strings.ContainsFunc(u, func(r rune) bool {
		return unicode.IsControl(r) || unicode.IsSpace(r)
	}) {
		return "", errUnsafeURL
	}

	parsed, err := url.Parse(u)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", errUnsafeURL
	}

	host := parsed.Hostname()
	if host == "" || parsed.User != nil || strings.EqualFold(host, "localhost") {
		return "", errUnsafeURL
	}

	if ip := net.ParseIP(host); ip != nil &&
		(ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
			ip.IsPrivate() || ip.IsUnspecified()) {
		return "", errUnsafeURL
	}

	return parsed.String(), nil
}
