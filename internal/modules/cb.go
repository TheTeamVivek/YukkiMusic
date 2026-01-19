/*
  - This file is part of YukkiMusic.
    *

  - YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
  - Copyright (C) 2025 TheTeamVivek
    *
  - This program is free software: you can redistribute it and/or modify
  - it under the terms of the GNU General Public License as published by
  - the Free Software Foundation, either version 3 of the License, or
  - (at your option) any later version.
    *
  - This program is distributed in the hope that it will be useful,
  - but WITHOUT ANY WARRANTY; without even the implied warranty of
  - MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
  - GNU General Public License for more details.
    *
  - You should have received a copy of the GNU General Public License
  - along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package modules

import (
	"context"
	"fmt"
	"html"
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

// Action handlers map for cleaner dispatch
type actionHandler func(*tg.CallbackQuery, *core.RoomState, int64) error

var actionHandlers = map[string]actionHandler{
	"pause":  handlePauseAction,
	"resume": handleResumeAction,
	"replay": handleReplayAction,
	"skip":   handleSkipAction,
	"stop":   handleStopAction,
	"mute":   handleMuteAction,
	"unmute": handleUnmuteAction,
}

func cancelHandler(cb *tg.CallbackQuery) error {
	chatID := cb.ChannelID()
	opt := &tg.CallbackOptions{Alert: true}

	if !checkAdminOrAuth(cb, chatID, opt) {
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
	data := cb.DataString()

	// Parse action type
	action := strings.TrimPrefix(data, "croom:")
	action = strings.TrimPrefix(action, "room:")

	if action == "" {
		gologging.WarnF("Missing action in data: %s", data)
		cb.Answer(F(cb.ChannelID(), "invalid_request"), opt)
		return tg.ErrEndGroup
	}

	chatID := cb.ChannelID()

	// Handle cplay mode
	if strings.HasPrefix(cb.DataString(), "croom:") {
		realChatID, err := database.GetCPlayID(chatID)
		if err != nil {
			gologging.ErrorF(
				"Failed to get chat ID for cplay ID %d: %v",
				chatID,
				err,
			)
			cb.Answer(F(chatID, "room_not_linked"), opt)
			return tg.ErrEndGroup
		}
		chatID = realChatID
	}

	// Get room
	r, err := getRoomForCallback(chatID)
	if err != nil {
		if strings.Contains(err.Error(), "no active room") {
			cb.Answer(F(cb.ChannelID(), "room_not_active_cb"), opt)
			editMessage(cb, F(cb.ChannelID(), "room_no_active"))
		} else {
			cb.Answer(err.Error(), opt)
		}
		return tg.ErrEndGroup
	}

	// Check permissions
	if !checkAdminOrAuth(cb, chatID, opt) {
		return tg.ErrEndGroup
	}

	// Flood control
	if !checkFloodControl(cb, chatID, opt) {
		return tg.ErrEndGroup
	}

	// Handle seek actions
	if strings.HasPrefix(action, "seek") {
		return handleSeekAction(cb, r, action, opt)
	}

	// Dispatch to handler
	if handler, ok := actionHandlers[action]; ok {
		return handler(cb, r, chatID)
	}

	gologging.WarnF("Unknown callback type: %s", action)
	cb.Answer(F(cb.ChannelID(), "unknown_action"), opt)
	return tg.ErrEndGroup
}

// Helper functions

func getRoomForCallback(chatID int64) (*core.RoomState, error) {
	ass, err := core.Assistants.ForChat(chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assistant: %w", err)
	}

	r, ok := core.GetRoom(chatID, ass)
	if !ok || !r.IsActiveChat() {
		return nil, fmt.Errorf("no active room")
	}

	return r, nil
}

func checkAdminOrAuth(
	cb *tg.CallbackQuery,
	chatID int64,
	opt *tg.CallbackOptions,
) bool {
	isAdmin, err := utils.IsChatAdmin(cb.Client, chatID, cb.SenderID)
	if err != nil || !isAdmin {
		cb.Answer(F(cb.ChannelID(), "only_admin_or_auth_cb"), opt)
		return false
	}
	return true
}

func checkFloodControl(
	cb *tg.CallbackQuery,
	chatID int64,
	opt *tg.CallbackOptions,
) bool {
	key := fmt.Sprintf("room:%d:%d", cb.Sender.ID, chatID)
	if remaining := utils.GetFlood(key); remaining > 0 {
		cb.Answer(F(cb.ChannelID(), "flood_seconds", locales.Arg{
			"duration": int(remaining.Seconds()),
		}), opt)
		return false
	}
	utils.SetFlood(key, 5*time.Second)
	return true
}

func editMessage(cb *tg.CallbackQuery, text string) {
	if _, err := cb.Edit(text); err != nil {
		gologging.ErrorF("Edit error: %v", err)
	}
}

func replyToCallback(cb *tg.CallbackQuery, text string) {
	msg, err := cb.GetMessage()
	if err != nil {
		return
	}
	msg.Reply(text)
}

// Action handlers

func handlePauseAction(
	cb *tg.CallbackQuery,
	r *core.RoomState,
	chatID int64,
) error {
	opt := &tg.CallbackOptions{Alert: true}

	gologging.InfoF("Callback → pause, chatID=%d", chatID)

	if r.IsPaused() {
		remaining := r.RemainingResumeDuration()
		msg := utils.IfElse(
			remaining > 0,
			F(cb.ChannelID(), "room_already_paused_auto", locales.Arg{
				"duration": formatDuration(int(remaining.Seconds())),
			}),
			F(cb.ChannelID(), "room_already_paused"),
		)
		cb.Answer(msg, opt)
		return tg.ErrEndGroup
	}

	if _, err := r.Pause(); err != nil {
		gologging.ErrorF("Pause failed: %v", err)
		cb.Answer(F(cb.ChannelID(), "room_pause_failed", locales.Arg{
			"error": err.Error(),
		}), opt)
		return tg.ErrEndGroup
	}

	if r.IsMuted() {
		r.Unmute()
	}

	cb.Answer(F(cb.ChannelID(), "cb_pause_success", locales.Arg{
		"position": formatDuration(r.Position()),
	}), opt)

	updatePlaybackMessage(cb, r, "paused")
	return tg.ErrEndGroup
}

func handleResumeAction(
	cb *tg.CallbackQuery,
	r *core.RoomState,
	chatID int64,
) error {
	opt := &tg.CallbackOptions{Alert: true}

	gologging.InfoF("Callback → resume, chatID=%d", chatID)

	if !r.IsPaused() {
		cb.Answer(F(cb.ChannelID(), "cb_already_playing"), opt)
		return tg.ErrEndGroup
	}

	if _, err := r.Resume(); err != nil {
		gologging.ErrorF("Resume failed: %v", err)
		cb.Answer(F(cb.ChannelID(), "cb_resume_failed"), opt)
		return tg.ErrEndGroup
	}

	cb.Answer(F(cb.ChannelID(), "cb_resume_success", locales.Arg{
		"position": formatDuration(r.Position()),
	}), opt)

	updatePlaybackMessage(cb, r, "playing")
	return tg.ErrEndGroup
}

func handleReplayAction(
	cb *tg.CallbackQuery,
	r *core.RoomState,
	chatID int64,
) error {
	opt := &tg.CallbackOptions{Alert: true}

	gologging.InfoF("Callback → replay, chatID=%d", chatID)

	mystic, err := cb.Respond(F(cb.ChannelID(), "cb_replaying"))
	if err != nil {
		gologging.ErrorF("Failed to send replay message: %v", err)
		return tg.ErrEndGroup
	}

	if err := r.Replay(); err != nil {
		gologging.ErrorF("Replay failed: %v", err)
		utils.EOR(mystic, F(cb.ChannelID(), "replay_failed", locales.Arg{
			"error": err.Error(),
		}))
		cb.Answer(F(cb.ChannelID(), "cb_replay_failed"), opt)
		return tg.ErrEndGroup
	}

	track := r.Track()
	trackTitle := html.EscapeString(utils.ShortTitle(track.Title, 25))

	msgText := F(cb.ChannelID(), "stream_now_playing", locales.Arg{
		"url":      track.URL,
		"title":    trackTitle,
		"duration": formatDuration(track.Duration),
		"by":       track.Requester,
	})

	cb.Answer(F(cb.ChannelID(), "cb_replay_success"), opt)

	optSend := &tg.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: core.GetPlayMarkup(cb.ChannelID(), r, false),
	}
	if track.Artwork != "" {
		optSend.Media = utils.CleanURL(track.Artwork)
	}

	mystic, _ = utils.EOR(mystic, msgText, optSend)
	r.SetMystic(mystic)

	editMessage(cb, F(cb.ChannelID(), "cb_replay_edited", locales.Arg{
		"user": utils.MentionHTML(cb.Sender),
	}))
	return tg.ErrEndGroup
}

func handleSkipAction(
	cb *tg.CallbackQuery,
	r *core.RoomState,
	chatID int64,
) error {
	opt := &tg.CallbackOptions{Alert: true}

	gologging.InfoF("Callback → skip, chatID=%d", chatID)

	if len(r.Queue()) == 0 && r.Loop() == 0 {
		r.Destroy()
		editMessage(cb, F(cb.ChannelID(), "skip_stopped", locales.Arg{
			"user": utils.MentionHTML(cb.Sender),
		}))
		cb.Answer(F(cb.ChannelID(), "cb_skip_queue_empty"), opt)
		return tg.ErrEndGroup
	}

	t := r.NextTrack()

	mystic, err := cb.Respond(F(cb.ChannelID(), "stream_downloading_next"))
	if err != nil {
		gologging.ErrorF("Failed to send message: %v", err)
	}

	path, err := platforms.Download(context.Background(), t, mystic)
	if err != nil {
		gologging.ErrorF("Download failed for %s: %v", t.URL, err)
		utils.EOR(mystic, F(cb.ChannelID(), "stream_download_fail", locales.Arg{
			"error": err.Error(),
		}))
		cb.Answer(F(cb.ChannelID(), "cb_skip_download_failed"), opt)
		return tg.ErrEndGroup
	}

	if err := r.Play(t, path); err != nil {
		gologging.ErrorF("Play error: %v", err)
		utils.EOR(mystic, F(cb.ChannelID(), "stream_play_fail"))
		cb.Answer(F(cb.ChannelID(), "cb_skip_play_failed"), opt)
		return tg.ErrEndGroup
	}

	cb.Answer(F(cb.ChannelID(), "cb_skip_success"), opt)
	cb.Delete()

	title := utils.ShortTitle(t.Title, 25)
	safeTitle := html.EscapeString(title)

	msgText := F(cb.ChannelID(), "stream_now_playing", locales.Arg{
		"url":      t.URL,
		"title":    safeTitle,
		"duration": formatDuration(t.Duration),
		"by":       t.Requester,
	})

	sendOpt := &tg.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: core.GetPlayMarkup(cb.ChannelID(), r, false),
	}
	if t.Artwork != "" {
		sendOpt.Media = utils.CleanURL(t.Artwork)
	}

	mystic, _ = utils.EOR(mystic, msgText, sendOpt)
	replyToCallback(cb, F(cb.ChannelID(), "cb_skip_edited", locales.Arg{
		"user": utils.MentionHTML(cb.Sender),
	}))

	r.SetMystic(mystic)
	return tg.ErrEndGroup
}

func handleStopAction(
	cb *tg.CallbackQuery,
	r *core.RoomState,
	chatID int64,
) error {
	opt := &tg.CallbackOptions{Alert: true}

	gologging.InfoF("Callback → stop, chatID=%d", chatID)

	r.Destroy()

	cb.Answer(F(cb.ChannelID(), "cb_stop_success"), opt)
	editMessage(cb, F(cb.ChannelID(), "stopped", locales.Arg{
		"user": utils.MentionHTML(cb.Sender),
	}))

	return tg.ErrEndGroup
}

func handleMuteAction(
	cb *tg.CallbackQuery,
	r *core.RoomState,
	chatID int64,
) error {
	opt := &tg.CallbackOptions{Alert: true}

	if r.IsMuted() {
		remaining := r.RemainingUnmuteDuration()
		msg := utils.IfElse(
			remaining > 0,
			F(cb.ChannelID(), "mute_already_muted_with_time", locales.Arg{
				"duration": formatDuration(int(remaining.Seconds())),
			}),
			F(cb.ChannelID(), "mute_already_muted"),
		)
		cb.Answer(msg, opt)
		return tg.ErrEndGroup
	}

	if _, err := r.Mute(); err != nil {
		cb.Answer(F(cb.ChannelID(), "mute_failed", locales.Arg{
			"error": err.Error(),
		}), opt)
		return tg.ErrEndGroup
	}

	cb.Answer(F(cb.ChannelID(), "cb_mute_success"), opt)
	updatePlaybackMessage(cb, r, "muted")
	return tg.ErrEndGroup
}

func handleUnmuteAction(
	cb *tg.CallbackQuery,
	r *core.RoomState,
	chatID int64,
) error {
	opt := &tg.CallbackOptions{Alert: true}

	if !r.IsMuted() {
		cb.Answer(F(cb.ChannelID(), "unmute_already"), opt)
		return tg.ErrEndGroup
	}

	if _, err := r.Unmute(); err != nil {
		cb.Answer(F(cb.ChannelID(), "unmute_failed", locales.Arg{
			"error": err.Error(),
		}), opt)
		return tg.ErrEndGroup
	}

	cb.Answer(F(cb.ChannelID(), "cb_unmute_success"), opt)
	updatePlaybackMessage(cb, r, "playing")
	return tg.ErrEndGroup
}

func handleSeekAction(
	cb *tg.CallbackQuery,
	r *core.RoomState,
	action string,
	opt *tg.CallbackOptions,
) error {
	parts := strings.Split(action, "_")
	if len(parts) != 2 {
		cb.Answer(F(cb.ChannelID(), "invalid_request"), opt)
		return tg.ErrEndGroup
	}

	seconds, err := strconv.Atoi(parts[1])
	if err != nil {
		cb.Answer(F(cb.ChannelID(), "invalid_request"), opt)
		return tg.ErrEndGroup
	}

	isBackward := strings.HasPrefix(action, "seekback_")

	if isBackward {
		if r.Position() <= seconds {
			r.Seek(-int(r.Position()))
		} else {
			r.Seek(-seconds)
		}
		cb.Answer(F(cb.ChannelID(), "cb_seekback_success", locales.Arg{
			"seconds": seconds,
		}), opt)
		replyToCallback(cb, F(cb.ChannelID(), "cb_seekback_edited", locales.Arg{
			"seconds": seconds,
			"user":    utils.MentionHTML(cb.Sender),
		}))
	} else {
		if (r.Track().Duration - r.Position()) <= seconds {
			cb.Answer(F(cb.ChannelID(), "cb_seek_near_end", locales.Arg{
				"seconds": seconds,
			}), opt)
			return tg.ErrEndGroup
		}
		r.Seek(seconds)
		cb.Answer(F(cb.ChannelID(), "cb_seek_success", locales.Arg{
			"seconds": seconds,
		}), opt)
		replyToCallback(cb, F(cb.ChannelID(), "cb_seek_edited", locales.Arg{
			"seconds": seconds,
			"user":    utils.MentionHTML(cb.Sender),
		}))
	}

	return tg.ErrEndGroup
}

func updatePlaybackMessage(
	cb *tg.CallbackQuery,
	r *core.RoomState,
	state string,
) {
	track := r.Track()

	if track == nil {
		return
	}
	safeTitle := html.EscapeString(utils.ShortTitle(track.Title, 25))
	mention := utils.MentionHTML(cb.Sender)

	var msgText string
	switch state {
	case "paused":
		msgText = F(cb.ChannelID(), "cb_pause_message", locales.Arg{
			"url":      track.URL,
			"title":    safeTitle,
			"position": formatDuration(r.Position()),
			"duration": formatDuration(track.Duration),
			"user":     mention,
		})
	case "playing":
		msgText = F(cb.ChannelID(), "cb_resume_message", locales.Arg{
			"url":      track.URL,
			"title":    safeTitle,
			"duration": formatDuration(track.Duration),
			"user":     mention,
		})
	case "muted":
		msgText = F(cb.ChannelID(), "cb_mute_message", locales.Arg{
			"url":   track.URL,
			"title": safeTitle,
			"user":  mention,
		})
	}

	editMessage(cb, msgText)

	if _, err := cb.Edit(msgText, &tg.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: core.GetPlayMarkup(cb.ChannelID(), r, false),
	}); err != nil {
		gologging.ErrorF("Edit error: %v", err)
	}
}
