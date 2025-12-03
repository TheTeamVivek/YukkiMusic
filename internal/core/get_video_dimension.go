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

func getVideoDimensions(filePath string) (int, int) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=s=x:p=0",
		filePath,
	)

	out, err := cmd.Output()
	if err != nil {
		gologging.Error("[getVideoDimensions] Failed to get video dimensions for " + filePath + " : " + err.Error())
		return 0, 0
	}

	dim := strings.Split(strings.TrimSpace(string(out)), "x")
	if len(dim) != 2 {
		gologging.Error("[getVideoDimensions] Invalid video dimensions for " + filePath + " : " + string(out))
		return 0, 0
	}

	width, _ := strconv.Atoi(dim[0])
	height, _ := strconv.Atoi(dim[1])
	return width, height
}