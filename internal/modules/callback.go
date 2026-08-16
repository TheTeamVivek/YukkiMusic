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
	"math"
	"strconv"
	"strings"
	"time"

	"yukkimusic/internal/logger"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/core"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/platforms"
	"yukkimusic/internal/utils"
)

func cancelHandler(c *td.Client, u *td.UpdateNewCallbackQuery) error {
	chatID := u.ChatId

	if !checkAdminOrAuth(c, u, chatID) {
		return nil
	}

	if downloads.cancel(chatID) {
		u.Answer(c, 0, true, F(chatID, "download_cancelled"), "")
	} else {
		u.Answer(c, 0, true, F(chatID, "no_download_to_cancel"), "")
	}
	return nil
}

func closeHandler(c *td.Client, u *td.UpdateNewCallbackQuery) error {
	u.Answer(c, 0, false, "", "")
	cbDelete(c, u)
	return nil
}

func emptyCBHandler(c *td.Client, u *td.UpdateNewCallbackQuery) error {
	u.Answer(c, 0, false, "", "")
	return nil
}

func roomHandle(c *td.Client, u *td.UpdateNewCallbackQuery) error {
	chatID := u.ChatId

	parts := strings.SplitN(u.DataString(), ":", 3)
	if len(parts) != 3 || parts[0] != "room" {
		logger.Warnf("Invalid room callback payload: %s", u.DataString())
		u.Answer(c, 0, true, F(chatID, "invalid_request"), "")
		cbDelete(c, u)
		return nil
	}
	roomID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		logger.Warnf("Invalid roomID in callback: %s", parts[1])
		u.Answer(c, 0, true, F(chatID, "invalid_request"), "")
		cbDelete(c, u)
		return nil
	}
	action := parts[2]

	r, ok := core.RoomFor(roomID)
	if !ok || !r.Active() {
		u.Answer(c, 0, true, F(chatID, "room_not_active_cb"), "")
		u.EditMessageText(c, F(chatID, "room_no_active"), nil)
		return nil
	}

	if !checkAdminOrAuth(c, u, chatID) {
		return nil
	}

	key := fmt.Sprintf("room:%d:%d", u.SenderUserId, chatID)
	if remaining := utils.GetFlood(key); remaining > 0 {
		u.Answer(c, 0, true, F(chatID, "flood_seconds", locales.Arg{
			"duration": int(math.Ceil(remaining.Seconds())),
		}), "")
		return nil
	}
	utils.SetFlood(key, 3*time.Second)

	switch {
	case action == "pause":
		return handlePauseAction(c, u, r)
	case action == "resume":
		return handleResumeAction(c, u, r)
	case action == "replay":
		return handleReplayAction(c, u, r)
	case action == "skip":
		return handleSkipAction(c, u, r)
	case action == "stop":
		return handleStopAction(c, u, r)
	case action == "mute":
		return handleMuteAction(c, u, r)
	case action == "unmute":
		return handleUnmuteAction(c, u, r)
	default:
		logger.Warnf("Unknown callback action: %s", action)
		u.Answer(c, 0, true, F(chatID, "unknown_action"), "")
	}

	return nil
}

func checkAdminOrAuth(c *td.Client, u *td.UpdateNewCallbackQuery, chatID int64) bool {
	if canUseAdminCommand(c, chatID, u.SenderUserId) {
		return true
	}

	mode, err := database.GetAdminMode(chatID)
	if err == nil && mode == database.AdminModeAdminsOnly {
		u.Answer(c, 0, true, F(chatID, "only_admin_cb"), "")
	} else {
		u.Answer(c, 0, true, F(chatID, "only_admin_or_auth_cb"), "")
	}
	return false
}

func cbDelete(c *td.Client, u *td.UpdateNewCallbackQuery) {
	msg, err := u.GetMessage(c)
	if err != nil || msg == nil {
		return
	}
	_ = msg.Delete(c, true)
}

func cbRespond(c *td.Client, u *td.UpdateNewCallbackQuery, text string, opts *td.SendTextMessageOpts) *td.Message {
	msg, err := u.GetMessage(c)
	if err != nil || msg == nil {
		logger.Errorf("cbRespond: failed to get callback message: %v", err)
		return nil
	}
	m, err := msg.ReplyText(c, text, opts)
	if err != nil {
		logger.Errorf("cbRespond: %v", err)
		return nil
	}
	return m
}

// mentionOfSender resolves the callback sender and builds an HTML mention.
func mentionOfSender(c *td.Client, userID int64) string {
	user, _ := c.GetUser(userID)
	return mentionOf(user, userID)
}

func handlePauseAction(c *td.Client, u *td.UpdateNewCallbackQuery, r *core.RoomState) error {
	chatID := u.ChatId
	logger.Infof("Callback → pause, chatID=%d", chatID)

	if r.Paused() {
		u.Answer(c, 0, true, F(chatID, "room_already_paused"), "")
		return nil
	}

	if _, err := r.Pause(); err != nil {
		logger.Errorf("Pause failed: %v", err)
		u.Answer(c, 0, true, F(chatID, "room_pause_failed", locales.Arg{
			"error": err.Error(),
		}), "")
		return nil
	}

	if r.Muted() {
		r.Unmute()
	}

	u.Answer(c, 0, true, F(chatID, "cb_pause_success", locales.Arg{
		"position": utils.FormatDuration(r.Position()),
	}), "")
	updatePlaybackMessage(c, u, r, "paused")
	return nil
}

func handleResumeAction(c *td.Client, u *td.UpdateNewCallbackQuery, r *core.RoomState) error {
	chatID := u.ChatId
	logger.Infof("Callback → resume, chatID=%d", chatID)

	if !r.Paused() {
		u.Answer(c, 0, true, F(chatID, "cb_already_playing"), "")
		return nil
	}

	if _, err := r.Resume(); err != nil {
		logger.Errorf("Resume failed: %v", err)
		u.Answer(c, 0, true, F(chatID, "cb_resume_failed"), "")
		return nil
	}

	u.Answer(c, 0, true, F(chatID, "cb_resume_success", locales.Arg{
		"position": utils.FormatDuration(r.Position()),
	}), "")
	updatePlaybackMessage(c, u, r, "playing")
	return nil
}

func handleReplayAction(c *td.Client, u *td.UpdateNewCallbackQuery, r *core.RoomState) error {
	chatID := u.ChatId
	logger.Infof("Callback → replay, chatID=%d", chatID)

	statusMsg := cbRespond(c, u, F(chatID, "cb_replaying"), nil)
	if statusMsg == nil {
		logger.Errorf("Failed to send replay status")
		return nil
	}

	if err := r.Replay(); err != nil {
		logger.Errorf("Replay failed: %v", err)
		utils.EOR(c, statusMsg, F(chatID, "replay_failed", locales.Arg{
			"error": err.Error(),
		}), nil)
		u.Answer(c, 0, true, F(chatID, "cb_replay_failed"), "")
		return nil
	}

	statusMsg = sendNowPlaying(c, statusMsg, chatID, r, r.Track())
	r.SetStatusMsg(statusMsg)

	u.Answer(c, 0, true, F(chatID, "cb_replay_success"), "")
	if _, err := u.EditMessageText(c, F(chatID, "cb_replay_edited", locales.Arg{
		"user": mentionOfSender(c, u.SenderUserId),
	}), &td.EditTextMessageOpts{ParseMode: "HTML"}); err != nil {
		logger.Errorf("Edit error: %v", err)
	}
	return nil
}

func handleSkipAction(c *td.Client, u *td.UpdateNewCallbackQuery, r *core.RoomState) error {
	chatID := u.ChatId
	logger.Infof("Callback → skip, chatID=%d", chatID)

	if len(r.Queue()) == 0 {
		scheduleOldPlayingMessage(r)
		core.DropRoom(r.ID)
		if _, err := u.EditMessageText(c, F(chatID, "skip_stopped", locales.Arg{
			"user": mentionOfSender(c, u.SenderUserId),
		}), &td.EditTextMessageOpts{ParseMode: "HTML"}); err != nil {
			logger.Errorf("Edit error: %v", err)
		}
		u.Answer(c, 0, true, F(chatID, "cb_skip_queue_empty"), "")
		return nil
	}

	r.SetLoop(0)
	t := r.NextTrack()

	statusMsg := cbRespond(c, u, F(chatID, "stream_downloading_next"), nil)
	if statusMsg == nil {
		logger.Errorf("Failed to send status message")
	}

	path, err := platforms.Download(context.Background(), t, statusMsg)
	if err != nil {
		logger.Errorf("Download failed for %s: %v", t.URL, err)
		utils.EOR(c, statusMsg, F(chatID, "stream_download_fail", locales.Arg{
			"error": err.Error(),
		}), nil)
		u.Answer(c, 0, true, F(chatID, "cb_skip_download_failed"), "")
		scheduleOldPlayingMessage(r)
		core.DropRoom(r.ID)
		return nil
	}

	if err := r.Play(t, path); err != nil {
		logger.Errorf("Play error: %v", err)
		utils.EOR(c, statusMsg, F(chatID, "stream_play_fail"), nil)
		u.Answer(c, 0, true, F(chatID, "cb_skip_play_failed"), "")
		scheduleOldPlayingMessage(r)
		core.DropRoom(r.ID)
		return nil
	}

	u.Answer(c, 0, true, F(chatID, "cb_skip_success"), "")
	cbDelete(c, u)

	statusMsg = sendNowPlaying(c, statusMsg, chatID, r, t)
	r.SetStatusMsg(statusMsg)

_, err = statusMsg.ReplyText(c, F(chatID, "cb_skip_edited", locales.Arg{
		"user": mentionOfSender(c, u.SenderUserId),
	}), &td.SendTextMessageOpts{ParseMode: "HTML"})

	return err
}

func handleStopAction(c *td.Client, u *td.UpdateNewCallbackQuery, r *core.RoomState) error {
	chatID := u.ChatId
	logger.Infof("Callback → stop, chatID=%d", chatID)

	scheduleOldPlayingMessage(r)
	core.DropRoom(r.ID)

	u.Answer(c, 0, true, F(chatID, "cb_stop_success"), "")
	if _, err := u.EditMessageText(c, F(chatID, "stopped", locales.Arg{
		"user": mentionOfSender(c, u.SenderUserId),
	}), &td.EditTextMessageOpts{ParseMode: "HTML"}); err != nil {
		logger.Errorf("Edit error: %v", err)
	}
	return nil
}

func handleMuteAction(c *td.Client, u *td.UpdateNewCallbackQuery, r *core.RoomState) error {
	chatID := u.ChatId

	if r.Muted() {
		u.Answer(c, 0, true, F(chatID, "mute_already_muted"), "")
		return nil
	}

	if _, err := r.Mute(); err != nil {
		u.Answer(c, 0, true, F(chatID, "mute_failed", locales.Arg{
			"error": err.Error(),
		}), "")
		return nil
	}

	u.Answer(c, 0, true, F(chatID, "cb_mute_success"), "")
	updatePlaybackMessage(c, u, r, "muted")
	return nil
}

func handleUnmuteAction(c *td.Client, u *td.UpdateNewCallbackQuery, r *core.RoomState) error {
	chatID := u.ChatId

	if !r.Muted() {
		u.Answer(c, 0, true, F(chatID, "unmute_already"), "")
		return nil
	}

	if _, err := r.Unmute(); err != nil {
		u.Answer(c, 0, true, F(chatID, "unmute_failed", locales.Arg{
			"error": err.Error(),
		}), "")
		return nil
	}

	u.Answer(c, 0, true, F(chatID, "cb_unmute_success"), "")
	updatePlaybackMessage(c, u, r, "playing")
	return nil
}

func updatePlaybackMessage(c *td.Client, u *td.UpdateNewCallbackQuery, r *core.RoomState, state string) {
	track := r.Track()
	if track == nil {
		return
	}

	chatID := u.ChatId
	safeTitle := utils.EscapeHTML(utils.ShortTitle(track.Title, 25))
	mention := mentionOfSender(c, u.SenderUserId)

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

	markup := core.GetPlayMarkup(chatID, r, false)

	msg, err := u.GetMessage(c)
	if err != nil || msg == nil {
		return
	}

	switch {
	case isPhotoMessage(msg):
		_, err = u.EditMessageCaption(c, msgText, &td.EditCaptionOpts{
			ParseMode:   "HTML",
			ReplyMarkup: markup,
		})
	default:
		_, err = u.EditMessageText(c, msgText, &td.EditTextMessageOpts{
			ParseMode:             "HTML",
			ReplyMarkup:           markup,
			DisableWebPagePreview: true,
		})
	}
	if err != nil {
		logger.Errorf("Edit error: %v", err)
	}
}
