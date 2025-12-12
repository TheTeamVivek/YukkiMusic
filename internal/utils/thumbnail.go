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

package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"resty.dev/v3"

	state "main/internal/core/models"
)

const (
	CACHE_DIR = "cache"
	W         = 1920
	H         = 1080
	fontPath  = "internal/utils/_font.ttf"
	font2Path = "internal/utils/_font2.ttf"
)

// FormatDuration formats a duration given in milliseconds as MM:SS.
func FormatDuration(d int) string {
	d /= 1000
	m := d / 60
	s := d % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

func trimToWidth(dc *gg.Context, text string, maxWidth float64) string {
	ellipsis := "…"
	if width, _ := dc.MeasureString(text); width <= maxWidth {
		return text
	}

	runes := []rune(text)
	for i := len(runes); i > 0; i-- {
		candidate := string(runes[:i]) + ellipsis
		if width, _ := dc.MeasureString(candidate); width <= maxWidth {
			return candidate
		}
	}

	return ellipsis
}

func GenerateThumbnail(ctx context.Context, track *state.Track, artist string) (string, error) {
	cachePath := filepath.Join(CACHE_DIR, "thumn_"+track.ID+".png")
	if _, err := os.Stat(cachePath); err == nil {
		return cachePath, nil
	}

	if err := os.MkdirAll(CACHE_DIR, 0o755); err != nil {
		return "", err
	}

	// Download thumbnail
	thumbPath := filepath.Join(CACHE_DIR, fmt.Sprintf("thumb_%s.jpg", track.ID))
	defer os.Remove(thumbPath)

	client := resty.New()
	defer client.Close()

	resp, err := client.R().SetContext(ctx).SetOutputFileName(thumbPath).Get(track.Artwork)
	if err != nil {
		return "", fmt.Errorf("failed to download thumbnail: %w", err)
	}

	if resp.IsError() {
		return "", fmt.Errorf("failed to download thumbnail: status %s", resp.Status())
	}

	// Start drawing
	dc := gg.NewContext(W, H)

	// Background
	dc.SetHexColor("#121b21")
	dc.Clear()

	// Wave
	dc.Push()
	dc.SetHexColor("#1c252d")
	dc.DrawEllipse(W/2, 1005, W/2+200, 275)
	dc.Fill()
	dc.Pop()

	// Album art
	albumImg, err := gg.LoadImage(thumbPath)
	if err != nil {
		log.Printf("Failed to load thumbnail image, skipping album art: %v", err)
	} else {
		resizedAlbum := imaging.Resize(albumImg, 650, 650, imaging.Lanczos)
		dc.Push()
		dc.DrawRoundedRectangle(180, 220, 650, 650, 40)
		dc.Clip()
		dc.DrawImage(resizedAlbum, 180, 220)
		dc.ResetClip()
		dc.Pop()
	}

	// Text drawing
	textX := float64(900)
	y1 := float64(330)

	// Playing text
	if err := dc.LoadFontFace(fontPath, 55); err != nil {
		log.Printf("Failed to load font: %v", err)
		return "", fmt.Errorf("load font %s: %w", fontPath, err)
	}
	dc.SetHexColor("#b9c0c7")
	dc.DrawStringAnchored("Playing", textX, y1, 0, 0.2)

	// Title text
	if err := dc.LoadFontFace(font2Path, 100); err != nil {
		log.Printf("Failed to load font2: %v", err)
		return "", fmt.Errorf("load font %s: %w", font2Path, err)
	}
	titleTrimmed := trimToWidth(dc, track.Title, 950)
	dc.SetHexColor("#ffffff")
	dc.DrawStringAnchored(titleTrimmed, textX, y1+90, 0, 0.2)

	// Artist text
	if err := dc.LoadFontFace(font2Path, 65); err != nil {
		log.Printf("Failed to load font2 for artist text: %v", err)
		return "", fmt.Errorf("load font %s: %w", font2Path, err)
	}
	dc.SetHexColor("#cdcdcd")
	dc.DrawStringAnchored(artist, textX, y1+220, 0, 0.2)

	// Duration text
	if err := dc.LoadFontFace(fontPath, 50); err != nil {
		log.Printf("Failed to load font for duration: %v", err)
		return "", fmt.Errorf("load font %s: %w", fontPath, err)
	}
	dc.SetHexColor("#b4b4b4")
	durationStr := fmt.Sprintf("Duration: %s", FormatDuration(track.Duration))
	dc.DrawStringAnchored(durationStr, textX, y1+320, 0, 0.2)

	// Save
	if err := dc.SavePNG(cachePath); err != nil {
		return "", fmt.Errorf("failed to save thumbnail: %w", err)
	}

	return cachePath, nil
}
