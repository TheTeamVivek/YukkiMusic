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
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"main/internal/state"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
)

const (
	CACHE_DIR = "cache"
	W         = 1920
	H         = 1080
	fontPath = "assets/font.ttf"
	font2Path = "assets/font2.ttf"

)

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
	for i := len(text); i > 0; i-- {
		newText := text[:i] + ellipsis
		if width, _ := dc.MeasureString(newText); width <= maxWidth {
			return newText
		}
	}
	return ellipsis
}

func GenerateThumbnail(ctx context.Context, track *state.Track, artist string) (string, error) {
	cachePath := filepath.Join(CACHE_DIR, fmt.Sprintf("%s_spotify_style.png", track.ID))
	if _, err := os.Stat(cachePath); err == nil {
		return cachePath, nil
	}

	if err := os.MkdirAll(CACHE_DIR, 0755); err != nil {
		return "", err
	}
	// Download thumbnail
	thumbPath := filepath.Join(CACHE_DIR, fmt.Sprintf("thumb_%s.jpg", track.ID))
	defer os.Remove(thumbPath)

	resp, err := http.Get(track.Artwork)
	if err != nil {
		return "", fmt.Errorf("failed to download thumbnail: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download thumbnail: status %s", resp.Status)
	}

	outFile, err := os.Create(thumbPath)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(outFile, resp.Body)
	outFile.Close()
	if err != nil {
		return "", err
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
		log.Printf("Failed to load font, using default: %v", err)
	}
	dc.SetHexColor("#b9c0c7")
	dc.DrawStringAnchored("Playing", textX, y1, 0, 0.2)

	// Title text
	if err := dc.LoadFontFace(font2Path, 100); err != nil {
		log.Printf("Failed to load font2, using default: %v", err)
	}
	titleTrimmed := trimToWidth(dc, track.Title, 950)
	dc.SetHexColor("#ffffff")
	dc.DrawStringAnchored(titleTrimmed, textX, y1+90, 0, 0.2)

	// Artist text
	if err := dc.LoadFontFace(font2Path, 65); err != nil {
		log.Printf("Failed to load font2, using default: %v", err)
	}
	dc.SetHexColor("#cdcdcd")
	dc.DrawStringAnchored(artist, textX, y1+220, 0, 0.2)

	// Duration text
	if err := dc.LoadFontFace(fontPath, 50); err != nil {
		log.Printf("Failed to load font, using default: %v", err)
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
