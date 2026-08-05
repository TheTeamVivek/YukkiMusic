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

package main

import (
	"fmt"
	"os/exec"
)

func checkDependencies() error {
	for _, bin := range []string{
		"ffmpeg",
		"ffprobe",
		"yt-dlp",
	} {
		if _, err := exec.LookPath(bin); err != nil {
			return fmt.Errorf(
				"Required dependency %q was not found in your PATH.\n\n"+
					"Please install %s and ensure it is accessible from PATH before starting YukkiMusic.",
				bin,
				bin,
			)
		}
	}
	return nil
}
