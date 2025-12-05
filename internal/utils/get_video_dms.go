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
	"encoding/json"
	"os/exec"
	"time"

	"github.com/Laky-64/gologging"
)

const ffprobeTimeout = 7 * time.Second

type ffprobeOutput struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
}

func GetVideoDimensions(filePath string) (int, int) {
	ctx, cancel := context.WithTimeout(context.Background(), ffprobeTimeout)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		filePath,
	)

	out, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		gologging.Error("[getVideoDimensions] ffprobe timed out for " + filePath)
		return 0, 0
	}
	if err != nil {
		gologging.Error("[getVideoDimensions] ffprobe failed for " + filePath + " : " + err.Error())
		return 0, 0
	}

	var probe ffprobeOutput
	if err := json.Unmarshal(out, &probe); err != nil {
		gologging.Error("[getVideoDimensions] failed to parse ffprobe JSON for " + filePath + " : " + err.Error())
		return 0, 0
	}

	for _, s := range probe.Streams {
		if s.CodecType == "video" && s.Width > 0 && s.Height > 0 {
			return s.Width, s.Height
		}
	}

	for _, s := range probe.Streams {
		if s.Width > 0 && s.Height > 0 {
			return s.Width, s.Height
		}
	}

	gologging.Error("[getVideoDimensions] no valid video stream found for " + filePath)
	return 0, 0
}
