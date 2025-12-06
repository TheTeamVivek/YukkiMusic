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
	"fmt"
	"os"

	"github.com/Laky-64/gologging"
	"resty.dev/v3"

	"main/internal/utils"
)

func normalizeVideo(path string, speed float64) (int, int, int, string) {
	if speed <= 0 {
		speed = 1.0
	}

	w, h := utils.GetVideoDimensions(path)
	if w <= 0 || h <= 0 {
		w = 1280
		h = 720
	}

	maxW := 1280
	maxH := 720

	if w > maxW {
		h = h * maxW / w
		w = maxW
	}
	if h > maxH {
		w = w * maxH / h
		h = maxH
	}

	// Ensure even values (required by x264/yuv420p)
	if w%2 != 0 {
		w--
	}
	if h%2 != 0 {
		h--
	}

	fps := 30
	videoSpeed := 1.0 / speed
	filter := fmt.Sprintf("setpts=%.4f*PTS,scale=%d:%d", videoSpeed, w, h)

	return w, h, fps, filter
}

func downloadThumb(id, url string) string {
	thumbPath := "cache/thumb_" + id + ".jpg"

	if _, err := os.Stat(thumbPath); err == nil {
		return thumbPath
	}

	if err := os.MkdirAll("cache", 0o755); err != nil {
		gologging.Error("mkdir error:" + err.Error())
		return url
	}
	client := resty.New()

	defer client.Close()
	resp, err := client.R().
		SetOutputFileName(thumbPath).
		Get(url)
	if err != nil {
		gologging.Error("thumb download failed:" + err.Error())
		return url
	}

	if resp.IsError() {
		gologiing.Error("thumb HTTP error:" + resp.Status())
		return url
	}

	return thumbPath
}
