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

package modules

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/platforms"
	"main/internal/utils"
)

func cancelHandler(cb *tg.CallbackQuery) error {
	chatID := cb.ChannelID()
	opt := &tg.CallbackOptions{Alert: true}

	if !checkAdminOrAuth(cb, chatID) {
		return tg.ErrEndGroup
	}

	if cancel, ok := downloadCancels[chatID]; ok {
		cancel()
		delete(downloadCancels, chatID)
		cb.Answer(F(chatID, "download_cancelled"), opt)
	} else {
		cb.Answer(F(chatID, "no_download_to_cancel"), opt)
	}
	return tg.ErrEndGroup
}

func closeHandler(cb *tg.CallbackQuery) error {
	cb.Answer("")
	cb.Delete()
	return tg.ErrEndGroup
}

func emptyCBHandler(cb *tg.CallbackQuery) error {
	cb.Answer("")
	return tg.ErrEndGroup
}

func roomHandle(cb *tg.CallbackQuery) error {
	opt := &tg.CallbackOptions{Alert: true}
	chatID := cb.ChannelID()

	parts := strings.SplitN(cb.DataString(), ":", 3)
	if len(parts) != 3 || parts[0] != "room" {
		gologging.WarnF("Invalid room callback payload: %s", cb.DataString())
		cb.Answer(F(chatID, "invalid_request"), opt)
		cb.Delete()
		return tg.ErrEndGroup
	}
	roomID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		gologging.WarnF("Invalid roomID in callback: %s", parts[1])
		cb.Answer(F(chatID, "invalid_request"), opt)
		cb.Delete()
		return tg.ErrEndGroup
	}
	action := parts[2]

	r, ok := core.GetRoom(roomID, nil, false)
	if !ok || !r.IsActiveChat() {
		cb.Answer(F(chatID, "room_not_active_cb"), opt)
		cb.Edit(F(chatID, "room_no_active"))
		return tg.ErrEndGroup
	}

	if !checkAdminOrAuth(cb, chatID) {
		return tg.ErrEndGroup
	}

	key := fmt.Sprintf("room:%d:%d", cb.Sender.ID, chatID)
	if remaining := utils.GetFlood(key); remaining > 0 {
		cb.Answer(F(chatID, "flood_seconds", locales.Arg{
			"duration": int(remaining.Seconds()),
		}), opt)
		return tg.ErrEndGroup
	}
	utils.SetFlood(key, 5*time.Second)

	switch {
	case strings.HasPrefix(action, "seek"):
		return handleSeekAction(cb, r, action, opt)
	case action == "pause":
		return handlePauseAction(cb, r)
	case action == "resume":
		return handleResumeAction(cb, r)
	case action == "replay":
		return handleReplayAction(cb, r)
	case action == "skip":
		return handleSkipAction(cb, r)
	case action == "stop":
		return handleStopAction(cb, r)
	case action == "mute":
		return handleMuteAction(cb, r)
	case action == "unmute":
		return handleUnmuteAction(cb, r)
	default:
		gologging.WarnF("Unknown callback action: %s", action)
		cb.Answer(F(chatID, "unknown_action"), opt)
	}

	return tg.ErrEndGroup
}

func checkAdminOrAuth(cb *tg.CallbackQuery, chatID int64) bool {
	if canUseAdminCommand(cb.Client, chatID, cb.SenderID) {
		return true
	}

	opt := &tg.CallbackOptions{Alert: true}
	mode, err := database.GetAdminMode(chatID)
	if err == nil && mode == database.AdminModeAdminsOnly {
		cb.Answer(F(chatID, "only_admin_cb"), opt)
	} else {
		cb.Answer(F(chatID, "only_admin_or_auth_cb"), opt)
	}
	return false
}

func handlePauseAction(cb *tg.CallbackQuery, r *core.RoomState) error {
	opt := &tg.CallbackOptions{Alert: true}
	chatID := cb.ChannelID()
	gologging.InfoF("Callback → pause, chatID=%d", chatID)

	if r.IsPaused() {
		remaining := r.RemainingResumeDuration()
		msg := utils.IfElse(
			remaining > 0,
			F(chatID, "room_already_paused_auto", locales.Arg{
				"duration": utils.FormatDuration(int(remaining.Seconds())),
			}),
			F(chatID, "room_already_paused"),
		)
		cb.Answer(msg, opt)
		return tg.ErrEndGroup
	}

	if _, err := r.Pause(); err != nil {
		gologging.ErrorF("Pause failed: %v", err)
		cb.Answer(F(chatID, "room_pause_failed", locales.Arg{
			"error": err.Error(),
		}), opt)
		return tg.ErrEndGroup
	}

	if r.IsMuted() {
		r.Unmute()
	}

	cb.Answer(F(chatID, "cb_pause_success", locales.Arg{
		"position": utils.FormatDuration(r.Position()),
	}), opt)
	updatePlaybackMessage(cb, r, "paused")
	return tg.ErrEndGroup
}

func handleResumeAction(cb *tg.CallbackQuery, r *core.RoomState) error {
	opt := &tg.CallbackOptions{Alert: true}
	chatID := cb.ChannelID()
	gologging.InfoF("Callback → resume, chatID=%d", chatID)

	if !r.IsPaused() {
		cb.Answer(F(chatID, "cb_already_playing"), opt)
		return tg.ErrEndGroup
	}

	if _, err := r.Resume(); err != nil {
		gologging.ErrorF("Resume failed: %v", err)
		cb.Answer(F(chatID, "cb_resume_failed"), opt)
		return tg.ErrEndGroup
	}

	cb.Answer(F(chatID, "cb_resume_success", locales.Arg{
		"position": utils.FormatDuration(r.Position()),
	}), opt)
	updatePlaybackMessage(cb, r, "playing")
	return tg.ErrEndGroup
}

func handleReplayAction(cb *tg.CallbackQuery, r *core.RoomState) error {
	opt := &tg.CallbackOptions{Alert: true}
	chatID := cb.ChannelID()
	gologging.InfoF("Callback → replay, chatID=%d", chatID)

	statusMsg, err := cb.Respond(F(chatID, "cb_replaying"))
	if err != nil {
		gologging.ErrorF("Failed to send replay status: %v", err)
		return tg.ErrEndGroup
	}

	if err := r.Replay(); err != nil {
		gologging.ErrorF("Replay failed: %v", err)
		utils.EOR(statusMsg, F(chatID, "replay_failed", locales.Arg{
			"error": err.Error(),
		}))
		cb.Answer(F(chatID, "cb_replay_failed"), opt)
		return tg.ErrEndGroup
	}

	track := r.Track()
	msgText := F(chatID, "stream_now_playing", locales.Arg{
		"url":      track.URL,
		"title":    utils.EscapeHTML(utils.ShortTitle(track.Title, 25)),
		"duration": utils.FormatDuration(track.Duration),
		"by":       track.Requester,
	})

	sendOpt := &tg.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: core.GetPlayMarkup(chatID, r, false),
	}
	if track.Artwork != "" && shouldShowThumb(chatID) {
		sendOpt.Media = utils.CleanURL(track.Artwork)
	}

	statusMsg, _ = utils.EOR(statusMsg, msgText, sendOpt)
	r.SetStatusMsg(statusMsg)

	cb.Answer(F(chatID, "cb_replay_success"), opt)
	if _, err := cb.Edit(F(chatID, "cb_replay_edited", locales.Arg{
		"user": utils.MentionHTML(cb.Sender),
	})); err != nil {
		gologging.ErrorF("Edit error: %v", err)
	}
	return tg.ErrEndGroup
}

func handleSkipAction(cb *tg.CallbackQuery, r *core.RoomState) error {
	opt := &tg.CallbackOptions{Alert: true}
	chatID := cb.ChannelID()
	gologging.InfoF("Callback → skip, chatID=%d", chatID)

	if len(r.Queue()) == 0 {
		scheduleOldPlayingMessage(r)
		core.DeleteRoom(r.ID)
		if _, err := cb.Edit(F(chatID, "skip_stopped", locales.Arg{
			"user": utils.MentionHTML(cb.Sender),
		})); err != nil {
			gologging.ErrorF("Edit error: %v", err)
		}
		cb.Answer(F(chatID, "cb_skip_queue_empty"), opt)
		return tg.ErrEndGroup
	}

	r.SetLoop(0)
	t := r.NextTrack()

	statusMsg, err := cb.Respond(F(chatID, "stream_downloading_next"))
	if err != nil {
		gologging.ErrorF("Failed to send status message: %v", err)
	}

	path, err := platforms.Download(context.Background(), t, statusMsg)
	if err != nil {
		gologging.ErrorF("Download failed for %s: %v", t.URL, err)
		utils.EOR(statusMsg, F(chatID, "stream_download_fail", locales.Arg{
			"error": err.Error(),
		}))
		cb.Answer(F(chatID, "cb_skip_download_failed"), opt)
		scheduleOldPlayingMessage(r)
		core.DeleteRoom(r.ID)
		return tg.ErrEndGroup
	}

	if err := r.Play(t, path); err != nil {
		gologging.ErrorF("Play error: %v", err)
		utils.EOR(statusMsg, F(chatID, "stream_play_fail"))
		cb.Answer(F(chatID, "cb_skip_play_failed"), opt)
		scheduleOldPlayingMessage(r)
		core.DeleteRoom(r.ID)
		return tg.ErrEndGroup
	}

	cb.Answer(F(chatID, "cb_skip_success"), opt)
	cb.Delete()

	msgText := F(chatID, "stream_now_playing", locales.Arg{
		"url":      t.URL,
		"title":    utils.EscapeHTML(utils.ShortTitle(t.Title, 25)),
		"duration": utils.FormatDuration(t.Duration),
		"by":       t.Requester,
	})

	sendOpt := &tg.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: core.GetPlayMarkup(chatID, r, false),
	}
	if t.Artwork != "" && shouldShowThumb(chatID) {
		sendOpt.Media = utils.CleanURL(t.Artwork)
	}

	statusMsg, err = utils.EOR(statusMsg, msgText, sendOpt)
	if err != nil {
		cb.Respond(F(chatID, "cb_skip_edited", locales.Arg{
			"user": utils.MentionHTML(cb.Sender),
		}))
		return tg.ErrEndGroup
	}

	r.SetStatusMsg(statusMsg)
	statusMsg.Reply(F(chatID, "cb_skip_edited", locales.Arg{
		"user": utils.MentionHTML(cb.Sender),
	}))
	return tg.ErrEndGroup
}

func handleStopAction(cb *tg.CallbackQuery, r *core.RoomState) error {
	opt := &tg.CallbackOptions{Alert: true}
	chatID := cb.ChannelID()
	gologging.InfoF("Callback → stop, chatID=%d", chatID)

	scheduleOldPlayingMessage(r)
	core.DeleteRoom(r.ID)

	cb.Answer(F(chatID, "cb_stop_success"), opt)
	if _, err := cb.Edit(F(chatID, "stopped", locales.Arg{
		"user": utils.MentionHTML(cb.Sender),
	})); err != nil {
		gologging.ErrorF("Edit error: %v", err)
	}
	return tg.ErrEndGroup
}

func handleMuteAction(cb *tg.CallbackQuery, r *core.RoomState) error {
	opt := &tg.CallbackOptions{Alert: true}
	chatID := cb.ChannelID()

	if r.IsMuted() {
		remaining := r.RemainingUnmuteDuration()
		msg := utils.IfElse(
			remaining > 0,
			F(chatID, "mute_already_muted_with_time", locales.Arg{
				"duration": utils.FormatDuration(int(remaining.Seconds())),
			}),
			F(chatID, "mute_already_muted"),
		)
		cb.Answer(msg, opt)
		return tg.ErrEndGroup
	}

	if _, err := r.Mute(); err != nil {
		cb.Answer(F(chatID, "mute_failed", locales.Arg{
			"error": err.Error(),
		}), opt)
		return tg.ErrEndGroup
	}

	cb.Answer(F(chatID, "cb_mute_success"), opt)
	updatePlaybackMessage(cb, r, "muted")
	return tg.ErrEndGroup
}

func handleUnmuteAction(cb *tg.CallbackQuery, r *core.RoomState) error {
	opt := &tg.CallbackOptions{Alert: true}
	chatID := cb.ChannelID()

	if !r.IsMuted() {
		cb.Answer(F(chatID, "unmute_already"), opt)
		return tg.ErrEndGroup
	}

	if _, err := r.Unmute(); err != nil {
		cb.Answer(F(chatID, "unmute_failed", locales.Arg{
			"error": err.Error(),
		}), opt)
		return tg.ErrEndGroup
	}

	cb.Answer(F(chatID, "cb_unmute_success"), opt)
	updatePlaybackMessage(cb, r, "playing")
	return tg.ErrEndGroup
}

func handleSeekAction(
	cb *tg.CallbackQuery,
	r *core.RoomState,
	action string,
	opt *tg.CallbackOptions,
) error {
	chatID := cb.ChannelID()

	parts := strings.SplitN(action, "_", 2)
	if len(parts) != 2 {
		cb.Answer(F(chatID, "invalid_request"), opt)
		return tg.ErrEndGroup
	}

	// action is either "seek_<N>" or "seekback_<N>"
	// Strip the direction prefix to get the numeric suffix.
	numStr := parts[1]
	isBackward := strings.HasPrefix(action, "seekback_")

	seconds, err := strconv.Atoi(numStr)
	if err != nil {
		cb.Answer(F(chatID, "invalid_request"), opt)
		return tg.ErrEndGroup
	}

	if isBackward {
		if r.Position() <= seconds {
			r.Seek(-int(r.Position()))
		} else {
			r.Seek(-seconds)
		}
		cb.Answer(F(chatID, "cb_seekback_success", locales.Arg{"seconds": seconds}), opt)
		cb.Reply(F(chatID, "cb_seekback_edited", locales.Arg{
			"seconds": seconds,
			"user":    utils.MentionHTML(cb.Sender),
		}))
	} else {
		if (r.Track().Duration - r.Position()) <= seconds {
			cb.Answer(F(chatID, "cb_seek_near_end", locales.Arg{"seconds": seconds}), opt)
			return tg.ErrEndGroup
		}
		r.Seek(seconds)
		cb.Answer(F(chatID, "cb_seek_success", locales.Arg{"seconds": seconds}), opt)
		cb.Reply(F(chatID, "cb_seek_edited", locales.Arg{
			"seconds": seconds,
			"user":    utils.MentionHTML(cb.Sender),
		}))
	}

	return tg.ErrEndGroup
}

func updatePlaybackMessage(cb *tg.CallbackQuery, r *core.RoomState, state string) {
	track := r.Track()
	if track == nil {
		return
	}

	chatID := cb.ChannelID()
	safeTitle := utils.EscapeHTML(utils.ShortTitle(track.Title, 25))
	mention := utils.MentionHTML(cb.Sender)

	var msgText string
	switch state {
	case "paused":
		msgText = F(chatID, "cb_pause_message", locales.Arg{
			"url":      track.URL,
			"title":    safeTitle,
			"position": utils.FormatDuration(r.Position()),
			"duration": utils.FormatDuration(track.Duration),
			"user":     mention,
		})
	case "playing":
		msgText = F(chatID, "cb_resume_message", locales.Arg{
			"url":      track.URL,
			"title":    safeTitle,
			"duration": utils.FormatDuration(track.Duration),
			"user":     mention,
		})
	case "muted":
		msgText = F(chatID, "cb_mute_message", locales.Arg{
			"url":   track.URL,
			"title": safeTitle,
			"user":  mention,
		})
	}

	if _, err := cb.Edit(msgText, &tg.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: core.GetPlayMarkup(chatID, r, false),
	}); err != nil {
		gologging.ErrorF("Edit error: %v", err)
	}
}
