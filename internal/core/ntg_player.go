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
	"strconv"

	"main/ntgcalls"
	"main/ubot"
)

type NtgPlayer struct {
	Ntg *ubot.Context
}

func (p *NtgPlayer) Play(r *RoomState) error {
	desc := getMediaDescription(r.fpath, r.position, r.speed, r.track.Video)
	return p.Ntg.Play(r.chatID, desc)
}

func (p *NtgPlayer) Pause(r *RoomState) (bool, error) {
	return p.Ntg.Pause(r.chatID)
}

func (p *NtgPlayer) Resume(r *RoomState) (bool, error) {
	return p.Ntg.Resume(r.chatID)
}

func (p *NtgPlayer) Stop(r *RoomState) error {
	return p.Ntg.Stop(r.chatID)
}

func (p *NtgPlayer) Mute(r *RoomState) (bool, error) {
	return p.Ntg.Mute(r.chatID)
}

func (p *NtgPlayer) Unmute(r *RoomState) (bool, error) {
	return p.Ntg.Unmute(r.chatID)
}

func getMediaDescription(url string, pos int, speed float64, isVideo bool) ntgcalls.MediaDescription {
	if speed < 0.5 {
		speed = 0.5
	} else if speed > 4.0 {
		speed = 4.0
	}

	audio := &ntgcalls.AudioDescription{
		MediaSource:  ntgcalls.MediaSourceShell,
		SampleRate:   96000,
		ChannelCount: 2,
	}

	baseCmd := "ffmpeg "
	if pos > 0 {
		baseCmd += "-ss " + strconv.Itoa(pos) + " "
	}
	baseCmd += "-v warning -i \"" + url + "\" "

	// Audio pipeline
	audioCmd := baseCmd
	audioCmd += "-filter:a \"atempo=" + strconv.FormatFloat(speed, 'f', 2, 64) + "\" "
	audioCmd += "-f s16le -ac " + strconv.Itoa(int(audio.ChannelCount)) + " "
	audioCmd += "-ar " + strconv.Itoa(int(audio.SampleRate)) + " "
	audioCmd += "pipe:1"
	audio.Input = audioCmd

	if !isVideo {
		return ntgcalls.MediaDescription{
			Microphone: audio,
		}
	}

	w, h, fps, filter := normalizeVideo(url, speed)

	video := &ntgcalls.VideoDescription{
		MediaSource: ntgcalls.MediaSourceShell,
		Width:       int16(w),
		Height:      int16(h),
		Fps:         uint8(fps),
	}

	// Video ffmpeg command
	videoCmd := baseCmd
	videoCmd += "-filter:v \"" + filter + "\" "
	videoCmd += "-f rawvideo -r " + strconv.Itoa(fps) + " -pix_fmt yuv420p "
	videoCmd += "pipe:1"
	video.Input = videoCmd

	return ntgcalls.MediaDescription{
		Microphone: audio,
		Camera:     video,
	}
}
