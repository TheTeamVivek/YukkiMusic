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

package main

import (
	"os/exec"
)

func checkFFmpegAndFFprobe() {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		gologging.Fatal("❌ ffmpeg not found in PATH. Please install ffmpeg.")
	}

	if _, err := exec.LookPath("ffprobe"); err != nil {
		gologging.Fatal("❌ ffprobe not found in PATH. Please install ffprobe.")
	}
}