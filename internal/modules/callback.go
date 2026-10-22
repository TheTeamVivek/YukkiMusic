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

	if cancelDownload(chatID) {
		u.Answer(c, F(chatID, "download_cancelled"), &td.CallbackAnswerOpts{ShowAlert: true})
	} else {
		u.Answer(c, F(chatID, "no_download_to_cancel"), &td.CallbackAnswerOpts{ShowAlert: true})
	}
	return nil
}

func closeHandler(c *td.Client, u *td.UpdateNewCallbackQuery) error {
	u.Answer(c, "", nil)
	cbDelete(c, u)
	return nil
}

func emptyCBHandler(c *td.Client, u *td.UpdateNewCallbackQuery) error {
	u.Answer(c, "", nil)
	return nil
}

func roomHandle(c *td.Client, u *td.UpdateNewCallbackQuery) error {
	chatID := u.ChatId

	parts := strings.SplitN(u.DataString(), ":", 3)
	if len(parts) != 3 || parts[0] != "room" {
		utils.Log.Warnf("Invalid room callback payload: %s", u.DataString())
		u.Answer(c, F(chatID, "invalid_request"), &td.CallbackAnswerOpts{ShowAlert: true})
		cbDelete(c, u)
		return nil
	}
	roomID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		utils.Log.Warnf("Invalid roomID in callback: %s", parts[1])
		u.Answer(c, F(chatID, "invalid_request"), &td.CallbackAnswerOpts{ShowAlert: true})
		cbDelete(c, u)
		return nil
	}
	action := parts[2]

	r, ok := core.Rooms.Get(roomID)
	if !ok || !r.Active() {
		u.Answer(c, F(chatID, "room_not_active_cb"), &td.CallbackAnswerOpts{ShowAlert: true})
		u.EditMessageText(c, F(chatID, "room_no_active"), nil)
		return nil
	}

	if !checkAdminOrAuth(c, u, chatID) {
		return nil
	}

	key := fmt.Sprintf("room:%d:%d", u.SenderUserId, chatID)
	if remaining := utils.GetFlood(key); remaining > 0 {
		u.Answer(c, F(chatID, "flood_seconds", locales.Arg{
			"duration": int(math.Ceil(remaining.Seconds())),
		}), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}
	utils.SetFlood(key, 3*time.Second)

	switch action {
	case "pause":
		return handlePauseAction(c, u, r)
	case "resume":
		return handleResumeAction(c, u, r)
	case "replay":
		return handleReplayAction(c, u, r)
	case "skip":
		return handleSkipAction(c, u, r)
	case "stop":
		return handleStopAction(c, u, r)
	case "mute":
		return handleMuteAction(c, u, r)
	case "unmute":
		return handleUnmuteAction(c, u, r)
	default:
		utils.Log.Warnf("Unknown callback action: %s", action)
		u.Answer(c, F(chatID, "unknown_action"), &td.CallbackAnswerOpts{ShowAlert: true})
	}

	return nil
}

func queueDeleteHandler(c *td.Client, u *td.UpdateNewCallbackQuery) error {
	chatID := u.ChatId

	parts := strings.Split(u.DataString(), ":")
	if len(parts) != 4 || parts[0] != "qdel" {
		utils.Log.Warnf("Invalid queue delete payload: %s", u.DataString())
		u.Answer(c, F(chatID, "invalid_request"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	roomID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		u.Answer(c, F(chatID, "invalid_request"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	index, err := strconv.Atoi(parts[2])
	if err != nil {
		u.Answer(c, F(chatID, "invalid_request"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	shown, err := strconv.Atoi(parts[3])
	if err != nil {
		u.Answer(c, F(chatID, "invalid_request"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	if !checkAdminOrAuth(c, u, chatID) {
		return nil
	}

	r, ok := core.Rooms.Get(roomID)
	if !ok || !r.Active() {
		u.Answer(c, F(chatID, "room_not_active_cb"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	if index <= 0 || index > r.QueueLen() {
		u.Answer(c, F(chatID, "invalid_request"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	r.RemoveTrack(index - 1)

	if _, err := c.EditContent(u.ChatId, u.MessageId, &td.InputMessageRichMessage{
		Message: queueRich(r, shown),
	}, nil); err != nil {
		utils.Log.Errorf("Failed to edit queue message: %v", err)
	}

	u.Answer(c, F(chatID, "queue_delete_success"), &td.CallbackAnswerOpts{ShowAlert: true})
	return nil
}

func queueMoreHandler(c *td.Client, u *td.UpdateNewCallbackQuery) error {
	chatID := u.ChatId

	parts := strings.Split(u.DataString(), ":")
	if len(parts) != 2 || parts[0] != "qmore" {
		utils.Log.Warnf("Invalid queue more payload: %s", u.DataString())
		u.Answer(c, F(chatID, "invalid_request"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	roomID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		u.Answer(c, F(chatID, "invalid_request"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	r, ok := core.Rooms.Get(roomID)
	if !ok || !r.Active() {
		u.Answer(c, F(chatID, "room_not_active_cb"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	if _, err := c.EditContent(u.ChatId, u.MessageId, &td.InputMessageRichMessage{
		Message: queueRich(r, r.QueueLen()),
	}, nil); err != nil {
		utils.Log.Errorf("Failed to edit queue message: %v", err)
	}

	u.Answer(c, "", nil)
	return nil
}

func checkAdminOrAuth(c *td.Client, u *td.UpdateNewCallbackQuery, chatID int64) bool {
	if isAdminOrAuth(c, chatID, u.SenderUserId) {
		return true
	}

	mode, err := database.GetAdminMode(chatID)
	if err == nil && mode == database.AdminModeAdminsOnly {
		u.Answer(c, F(chatID, "only_admin_cb"), &td.CallbackAnswerOpts{ShowAlert: true})
	} else {
		u.Answer(c, F(chatID, "only_admin_or_auth_cb"), &td.CallbackAnswerOpts{ShowAlert: true})
	}
	return false
}

func cbDelete(c *td.Client, u *td.UpdateNewCallbackQuery) {
	msg, err := u.GetMessage(c)
	if err != nil || msg == nil {
		return
	}
	_ = msg.Delete(c, &td.DeleteMessagesOpts{Revoke: true})
}

func cbRespond(c *td.Client, u *td.UpdateNewCallbackQuery, text string, opts *td.SendMessageOpts) *td.Message {
	msg, err := u.GetMessage(c)
	if err != nil || msg == nil {
		utils.Log.Errorf("cbRespond: failed to get callback message: %v", err)
		return nil
	}
	m, err := msg.ReplyText(c, text, opts)
	if err != nil {
		utils.Log.Errorf("cbRespond: %v", err)
		return nil
	}
	return m
}

func handlePauseAction(c *td.Client, u *td.UpdateNewCallbackQuery, r *core.Room) error {
	chatID := u.ChatId
	utils.Log.Infof("Callback → pause, chatID=%d", chatID)

	if r.Paused() {
		u.Answer(c, F(chatID, "room_already_paused"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	if err := r.Pause(); err != nil {
		utils.Log.Errorf("Pause failed: %v", err)
		u.Answer(c, F(chatID, "cb_pause_failed"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	if r.Muted() {
		r.Unmute()
	}

	u.Answer(c, F(chatID, "cb_pause_success", locales.Arg{
		"position": utils.FormatDuration(r.Position()),
	}), &td.CallbackAnswerOpts{ShowAlert: true})
	applyActionEdit(c, u, r, "paused", "pause_success", core.GetPauseToggle(chatID, r))
	return nil
}

func handleResumeAction(c *td.Client, u *td.UpdateNewCallbackQuery, r *core.Room) error {
	chatID := u.ChatId
	utils.Log.Infof("Callback → resume, chatID=%d", chatID)

	if !r.Paused() {
		u.Answer(c, F(chatID, "cb_already_playing"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	if err := r.Resume(); err != nil {
		utils.Log.Errorf("Resume failed: %v", err)
		u.Answer(c, F(chatID, "cb_resume_failed"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	u.Answer(c, F(chatID, "cb_resume_success", locales.Arg{
		"position": utils.FormatDuration(r.Position()),
	}), &td.CallbackAnswerOpts{ShowAlert: true})
	applyActionEdit(c, u, r, "playing", "resume_success", core.GetPauseToggle(chatID, r))
	return nil
}

func handleReplayAction(c *td.Client, u *td.UpdateNewCallbackQuery, r *core.Room) error {
	chatID := u.ChatId
	sender, _ := c.GetUser(u.SenderUserId)
	utils.Log.Infof("Callback → replay, chatID=%d", chatID)

	statusMsg := cbRespond(c, u, F(chatID, "cb_replaying"), nil)
	if statusMsg == nil {
		utils.Log.Errorf("Failed to send replay status")
		return nil
	}

	if err := r.Replay(); err != nil {
		utils.Log.Errorf("Replay failed: %v", err)
		utils.EOR(c, statusMsg, F(chatID, "replay_failed", locales.Arg{
			"error": err.Error(),
		}), nil)
		u.Answer(c, F(chatID, "cb_replay_failed"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	statusMsg = sendNowPlaying(c, statusMsg, chatID, r, r.Track())
	r.SetStatusMsg(statusMsg)

	u.Answer(c, F(chatID, "cb_replay_success"), &td.CallbackAnswerOpts{ShowAlert: true})
	if _, err := u.EditMessageText(c, F(chatID, "cb_replay_edited", locales.Arg{
		"user": mention(sender, u.SenderUserId),
	}), &td.EditMessageTextOpts{ParseMode: "HTML"}); err != nil {
		utils.Log.Errorf("Edit error: %v", err)
	}
	return nil
}

func handleSkipAction(c *td.Client, u *td.UpdateNewCallbackQuery, r *core.Room) error {
	chatID := u.ChatId
	sender, _ := c.GetUser(u.SenderUserId)
	utils.Log.Infof("Callback → skip, chatID=%d", chatID)

	if r.QueueLen() == 0 {
		core.Rooms.Delete(r.ChatID)
		if _, err := u.EditMessageText(c, F(chatID, "skip_stopped", locales.Arg{
			"user": mention(sender, u.SenderUserId),
		}), &td.EditMessageTextOpts{ParseMode: "HTML"}); err != nil {
			utils.Log.Errorf("Edit error: %v", err)
		}
		u.Answer(c, F(chatID, "cb_skip_queue_empty"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	r.SetLoop(0)
	t := r.NextTrack()

	statusMsg := cbRespond(c, u, F(chatID, "stream_downloading_next"), nil)
	if statusMsg == nil {
		utils.Log.Errorf("Failed to send status message")
	}

	path, err := platforms.Download(context.Background(), t, statusMsg)
	if err != nil {
		utils.Log.Errorf("Download failed for %s: %v", t.URL, err)
		utils.EOR(c, statusMsg, F(chatID, "stream_download_fail", locales.Arg{
			"error": err.Error(),
		}), nil)
		u.Answer(c, F(chatID, "cb_skip_download_failed"), &td.CallbackAnswerOpts{ShowAlert: true})
		core.Rooms.Delete(r.ChatID)
		return nil
	}

	if err := r.Play(t, path); err != nil {
		utils.Log.Errorf("Play error: %v", err)
		utils.EOR(c, statusMsg, F(chatID, "stream_play_fail"), nil)
		u.Answer(c, F(chatID, "cb_skip_play_failed"), &td.CallbackAnswerOpts{ShowAlert: true})
		core.Rooms.Delete(r.ChatID)
		return nil
	}

	u.Answer(c, F(chatID, "cb_skip_success"), &td.CallbackAnswerOpts{ShowAlert: true})
	cbDelete(c, u)

	statusMsg = sendNowPlaying(c, statusMsg, chatID, r, t)
	r.SetStatusMsg(statusMsg)

	_, err = statusMsg.ReplyText(c, F(chatID, "cb_skip_edited", locales.Arg{
		"user": mention(sender, u.SenderUserId),
	}), &td.SendMessageOpts{ParseMode: "HTML"})

	return err
}

func handleStopAction(c *td.Client, u *td.UpdateNewCallbackQuery, r *core.Room) error {
	chatID := u.ChatId
	sender, _ := c.GetUser(u.SenderUserId)
	utils.Log.Infof("Callback → stop, chatID=%d", chatID)

	core.Rooms.Delete(r.ChatID)

	u.Answer(c, F(chatID, "cb_stop_success"), &td.CallbackAnswerOpts{ShowAlert: true})
	if _, err := u.EditMessageText(c, F(chatID, "stopped", locales.Arg{
		"user": mention(sender, u.SenderUserId),
	}), &td.EditMessageTextOpts{ParseMode: "HTML"}); err != nil {
		utils.Log.Errorf("Edit error: %v", err)
	}
	return nil
}

func handleMuteAction(c *td.Client, u *td.UpdateNewCallbackQuery, r *core.Room) error {
	chatID := u.ChatId

	if r.Muted() {
		u.Answer(c, F(chatID, "cb_already_muted"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	if err := r.Mute(); err != nil {
		utils.Log.Errorf("Mute failed: %v", err)
		u.Answer(c, F(chatID, "cb_mute_failed"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	u.Answer(c, F(chatID, "cb_mute_success"), &td.CallbackAnswerOpts{ShowAlert: true})
	applyActionEdit(c, u, r, "muted", "mute_success", core.GetMuteToggle(chatID, r))
	return nil
}

func handleUnmuteAction(c *td.Client, u *td.UpdateNewCallbackQuery, r *core.Room) error {
	chatID := u.ChatId

	if !r.Muted() {
		u.Answer(c, F(chatID, "cb_already_unmuted"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	if err := r.Unmute(); err != nil {
		utils.Log.Errorf("Unmute failed: %v", err)
		u.Answer(c, F(chatID, "cb_unmute_failed"), &td.CallbackAnswerOpts{ShowAlert: true})
		return nil
	}

	u.Answer(c, F(chatID, "cb_unmute_success"), &td.CallbackAnswerOpts{ShowAlert: true})
	applyActionEdit(c, u, r, "unmuted", "unmute_success", core.GetMuteToggle(chatID, r))
	return nil
}

func applyActionEdit(c *td.Client, u *td.UpdateNewCallbackQuery, r *core.Room, state, successKey string, toggle td.ReplyMarkup) {
	msg, err := u.GetMessage(c)
	if err != nil || msg == nil {
		return
	}

	status := r.StatusMsg()
	if status != nil && status.Id == msg.Id && status.ChatId == msg.ChatId {
		updatePlaybackMessageFor(c, msg, u.SenderUserId, r, state)
		return
	}

	track := r.Track()
	if track == nil {
		updatePlaybackMessageFor(c, status, u.SenderUserId, r, state)
		return
	}

	chatID := u.ChatId
	sender, _ := c.GetUser(u.SenderUserId)

	speedLine := ""
	if sp := r.Speed(); sp != 1.0 {
		speedLine = `<tr><td colspan="2">` + F(chatID, "speed_line", locales.Arg{
			"speed": fmt.Sprintf("%.2f", sp),
		}) + `</td></tr>`
	}

	text := F(chatID, successKey, locales.Arg{
		"title":      utils.EscapeHTML(utils.ShortTitle(track.Title, 25)),
		"url":        utils.EscapeHTML(track.URL),
		"position":   utils.FormatDuration(r.Position()),
		"duration":   utils.FormatDuration(track.Duration),
		"user":       mention(sender, u.SenderUserId),
		"speed_line": speedLine,
	})

	if _, err := msg.EditRichMessageHTML(c, text, &td.EditRichTextOpts{
		ReplyMarkup: toggle,
	}); err != nil {
		utils.Log.Errorf("Edit error: %v", err)
	}

	updatePlaybackMessageFor(c, status, u.SenderUserId, r, state)
}

func updatePlaybackMessageFor(c *td.Client, msg *td.Message, userID int64, r *core.Room, state string) {
	track := r.Track()
	if track == nil || msg == nil {
		return
	}

	chatID := msg.ChatID()
	safeTitle := utils.EscapeHTML(utils.ShortTitle(track.Title, 25))
	sender, _ := c.GetUser(userID)
	mention := mention(sender, userID)

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
	case "unmuted":
		msgText = F(chatID, "cb_unmute_message", locales.Arg{
			"url":   track.URL,
			"title": safeTitle,
			"user":  mention,
		})
	}

	markup := core.GetPlayMarkup(chatID, r, false)

	var err error
	switch {
	case isPhotoMessage(msg):
		_, err = msg.EditCaption(c, msgText, &td.EditMessageCaptionOpts{
			ParseMode:   "HTML",
			ReplyMarkup: markup,
		})
	default:
		_, err = msg.EditText(c, msgText, &td.EditMessageTextOpts{
			ParseMode:             "HTML",
			ReplyMarkup:           markup,
			DisableWebPagePreview: true,
		})
	}
	if err != nil {
		utils.Log.Errorf("Edit error: %v", err)
	}
}
