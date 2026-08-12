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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"yukkimusic/internal/logger"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	state "yukkimusic/internal/core/models"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/platforms"
	"yukkimusic/internal/utils"
	"yukkimusic/ubot"
)

type playOpts struct {
	Force bool
	CPlay bool
	Video bool
}

const playMaxRetries = 3

func init() {
	helpTexts["/play"] = `<i>Play a song in the voice chat from YouTube, Spotify, or other sources.</i>

<u>Usage:</u>
<b>/play [query/URL]</b> — Search and play a song
<b>/play [reply to audio/video]</b> — Play replied media

<b>🎵 Supported Sources:</b>
• YouTube (videos, playlists)
• Spotify (tracks, albums, playlists)
• SoundCloud
• Direct audio/video links

<b>⚙️ Features:</b>
• Queue support - adds to end if already playing
• Auto-join voice chat if not present
• Duration limit check
• Multiple track support (playlists)

<b>💡 Examples:</b>
<code>/play never gonna give you up</code>
<code>/play https://youtu.be/dQw4w9WgXcQ</code>

<b>⚠️ Notes:</b>
• Bot must have proper permissions in voice chat
• Tracks exceeding duration limit will be skipped
• Use <code>/queue</code> to view upcoming tracks
• Use <code>/fplay</code> to force play (skip queue)`

	helpTexts["/fplay"] = `<i>Force play a song, skipping the current queue.</i>

<u>Usage:</u>
<b>/fplay [query/URL]</b> — Force play immediately
<b>/fplay [reply to audio/video]</b> — Force play replied media

<b>🎵 Behavior:</b>
• Stops current playback
• Clears queue
• Starts playing immediately

<b>🔒 Restrictions:</b>
• Only <b>chat admins</b> or <b>authorized users</b> can use this

<b>💡 Example:</b>
<code>/fplay urgent announcement track</code>

<b>⚠️ Note:</b>
This command is useful for urgent playback needs but will disrupt the current queue.`

	helpTexts["/vplay"] = `<i>Play video content in voice chat (video mode).</i>

<u>Usage:</u>
<b>/vplay [query/URL]</b> — Play video
<b>/vplay [reply to video]</b> — Play replied video

<b>📹 Features:</b>
• Full video playback support
• Audio + Video streaming
• Same queue system as audio

<b>⚠️ Notes:</b>
• Requires video streaming permissions
• Use <code>/fvplay</code> for force video play`

	helpTexts["/fvplay"] = `<i>Force play video content, skipping queue.</i>

<u>Usage:</u>
<b>/fvplay [query/URL]</b> — Force play video immediately

<b>🔒 Restrictions:</b>
• Admin/auth only command

<b>💡 Use Case:</b>
Immediate video playback when something urgent needs to be shown.`

	helpTexts["/cplay"] = `<i>Play in linked channel's voice chat.</i>

<u>Usage:</u>
<b>/cplay [query]</b> — Play in linked channel

<b>⚙️ Setup Required:</b>
First use <code>/setcplay [channel_id]</code>

<b>⚠️ Note:</b>
All c* commands work the same as regular commands but affect the linked channel.`

	helpTexts["/channelplay"] = `<i>Configure linked channel for channel play mode.</i>

<u>Usage:</u>
<b>/channelplay [channel_id]</b> — Set linked channel

<b>⚙️ Behavior:</b>
• Links a channel to current group
• All <code>c*</code> commands affect linked channel
• Channel must be accessible by bot

<b>🔒 Restrictions:</b>
• Only <b>chat admins</b> can configure

<b>💡 Examples:</b>
<code>/setcplay -1001234567890</code>

<b>⚠️ Notes:</b>
• Get channel ID using forward + @userinfobot
• Bot must be admin in linked channel
• Use <code>/cplay</code> after setup`
	helpTexts["/setcplay"] = helpTexts["/channelplay"]

	helpTexts["/playforce"] = helpTexts["/fplay"]
	helpTexts["/fcplay"] = helpTexts["/cfplay"]
	helpTexts["/cvplay"] = helpTexts["/vcplay"]
}

func playHandler(c *td.Client, m *td.Message) error  { return handlePlay(c, m, &playOpts{}) }
func fplayHandler(c *td.Client, m *td.Message) error { return handlePlay(c, m, &playOpts{Force: true}) }
func cfplayHandler(c *td.Client, m *td.Message) error {
	return handlePlay(c, m, &playOpts{Force: true, CPlay: true})
}
func vplayHandler(c *td.Client, m *td.Message) error { return handlePlay(c, m, &playOpts{Video: true}) }
func fvplayHandler(c *td.Client, m *td.Message) error {
	return handlePlay(c, m, &playOpts{Force: true, Video: true})
}

func vcplayHandler(c *td.Client, m *td.Message) error {
	return handlePlay(c, m, &playOpts{CPlay: true, Video: true})
}

func fvcplayHandler(c *td.Client, m *td.Message) error {
	return handlePlay(c, m, &playOpts{Force: true, CPlay: true, Video: true})
}
func cplayHandler(c *td.Client, m *td.Message) error { return handlePlay(c, m, &playOpts{CPlay: true}) }

func handlePlay(c *td.Client, m *td.Message, opts *playOpts) error {
	chatID := m.ChatID()

	if !canUsePlayCommand(c, m, chatID) {
		_, err := m.ReplyText(c, F(chatID, "playmode_restricted"), nil)
		return err
	}

	room, searchMsg, err := prepareRoomAndSearchMessage(c, m, opts.CPlay)
	if err != nil {
		return nil
	}

	tracks, isActive, err := fetchTracksAndCheckStatus(c, m, searchMsg, room, opts.Video)
	if err != nil {
		return nil
	}

	if len(tracks) == 1 && !opts.Force {
		if isTrackInQueue(room, tracks[0]) {
			utils.EOR(c, searchMsg, F(m.ChatID(), "play_already_in_queue", locales.Arg{
				"title": utils.EscapeHTML(utils.ShortTitle(tracks[0].Title, 35)),
			}), nil)
			return nil
		}
	}

	tracks, availableSlots, err := filterAndTrimTracks(c, searchMsg, room, tracks)
	if err != nil {
		return nil
	}

	sender, _ := m.GetUser(c)
	mention := mentionOf(sender, m.SenderID())
	if err := playTracksAndRespond(c, m, searchMsg, room, tracks, mention, isActive, opts.Force, availableSlots); err != nil {
		return err
	}

	return nil
}

func canUsePlayCommand(c *td.Client, m *td.Message, chatID int64) bool {
	adminsOnly, _ := database.PlayModeAdminsOnly(chatID)
	if !adminsOnly {
		return true
	}

	isAdmin, err := utils.IsChatAdmin(c, chatID, m.SenderID())
	if err == nil && isAdmin {
		return true
	}

	isAuth, _ := database.IsAuthorized(chatID, m.SenderID())
	return isAuth
}

func prepareRoomAndSearchMessage(
	c *td.Client,
	m *td.Message,
	cplay bool,
) (*core.RoomState, *td.Message, error) {
	room, err := getEffectiveRoom(m.ChatID(), cplay)
	if err != nil {
		if _, rerr := m.ReplyText(c, err.Error(), nil); rerr != nil {
			return nil, nil, rerr
		}
		return nil, nil, err
	}

	chatID := m.ChatID()
	room.Parse()

	if len(room.Queue()) >= config.QueueLimit {
		if _, rerr := m.ReplyText(c, F(chatID, "queue_limit_reached", locales.Arg{"limit": config.QueueLimit}), nil); rerr != nil {
			return nil, nil, rerr
		}
		return nil, nil, fmt.Errorf("queue limit reached")
	}

	query := extractPlayQuery(m.Text())
	if query == "" && m.ReplyToMessageID() == 0 {
		if _, rerr := m.ReplyText(c, F(chatID, "no_song_query", locales.Arg{"cmd": getCommand(m)}), nil); rerr != nil {
			return nil, nil, rerr
		}
		return nil, nil, fmt.Errorf("no song query")
	}

	statusText := F(chatID, "searching")
	if query != "" {
		statusText = F(
			chatID,
			"searching_query",
			locales.Arg{"query": utils.EscapeHTML(query)},
		)
	}

	replyMsg, err := m.ReplyText(c, statusText, nil)
	if err != nil {
		logger.Errorf("Failed to send searching message: %v", err)
		return nil, nil, err
	}

	return room, replyMsg, nil
}

func extractPlayQuery(text string) string {
	parts := strings.SplitN(text, " ", 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func fetchTracksAndCheckStatus(
	c *td.Client,
	m *td.Message,
	replyMsg *td.Message,
	r *core.RoomState,
	video bool,
) ([]*state.Track, bool, error) {
	tracks, err := safeGetTracks(c, m, replyMsg, m.ChatID(), video)
	if err != nil {
		utils.EOR(c, replyMsg, err.Error(), nil)
		return nil, false, err
	}
	if len(tracks) == 0 {
		utils.EOR(c, replyMsg, F(m.ChatID(), "no_song_found"), nil)
		return nil, false, fmt.Errorf("no tracks found")
	}

	chatState, err := core.GetChatState(r.ID)
	if err != nil {
		logger.Errorf("Error getting chat state: %v", err)
		utils.EOR(c, replyMsg, getErrorMessage(m.ChatID(), err), nil)
		return nil, false, err
	}

	if err := ensureVoiceChatReady(c, m.ChatID(), replyMsg, chatState); err != nil {
		return nil, false, err
	}

	return tracks, r.IsActiveChat(), nil
}

func isTrackInQueue(r *core.RoomState, t *state.Track) bool {
	activeTrack := r.Track()
	if activeTrack != nil && (activeTrack.URL == t.URL || activeTrack.ID == t.ID) {
		return true
	}

	for _, qt := range r.Queue() {
		if qt.URL == t.URL || qt.ID == t.ID {
			return true
		}
	}
	return false
}

func ensureVoiceChatReady(
	c *td.Client,
	chatID int64,
	replyMsg *td.Message,
	cs *core.ChatState,
) error {
	snap, err := cs.Snapshot(false)
	if err != nil {
		logger.Errorf("Error checking voicechat state: %v", err)
		utils.EOR(c, replyMsg, getErrorMessage(chatID, err), nil)
		return err
	}

	if snap.VoiceChatActive != nil && !*snap.VoiceChatActive {
		err := fmt.Errorf("no active voice chat")
		utils.EOR(c, replyMsg, F(chatID, "err_no_active_voicechat"), nil)
		return err
	}

	if snap.AssistantBanned {
		err := fmt.Errorf("assistant banned")
		utils.EOR(c, replyMsg, F(chatID, "err_assistant_banned", locales.Arg{
			"user": mentionOfAssistant(cs.Assistant),
			"id":   utils.IntToStr(cs.Assistant.Self.ID),
		}), nil)
		return err
	}

	if snap.AssistantPresent {
		return nil
	}

	username := ""
	if chat, err := replyMsg.GetChat(c); err == nil {
		if ct, ok := chat.Type.(*td.ChatTypeSupergroup); ok {
			if sg, err := c.GetSupergroup(ct.SupergroupId); err == nil &&
				sg.Usernames != nil && len(sg.Usernames.ActiveUsernames) > 0 {
				username = sg.Usernames.ActiveUsernames[0]
			}
		}
	}
	if err := cs.EnsureAssistantJoined(username); err != nil {
		logger.Errorf("Error joining assistant: %v", err)
		utils.EOR(c, replyMsg, getErrorMessage(chatID, err), nil)
		return err
	}

	time.Sleep(1 * time.Second)
	return nil
}

func filterAndTrimTracks(
	c *td.Client,
	replyMsg *td.Message,
	r *core.RoomState,
	tracks []*state.Track,
) ([]*state.Track, int, error) {
	chatID := replyMsg.ChatID()
	accepted := make([]*state.Track, 0, len(tracks))
	skippedTitles := make([]string, 0)

	for _, track := range tracks {
		if track.Duration > config.DurationLimit {
			skippedTitles = append(
				skippedTitles,
				utils.EscapeHTML(utils.ShortTitle(track.Title, 35)),
			)
			continue
		}
		accepted = append(accepted, track)
	}

	if len(skippedTitles) > 0 {
		if len(tracks) == 1 && len(accepted) == 0 {
			utils.EOR(c, replyMsg, F(chatID, "play_single_track_too_long", locales.Arg{
				"limit_mins": utils.FormatDuration(config.DurationLimit),
				"title":      skippedTitles[0],
			}), nil)
			return nil, 0, fmt.Errorf("single long track skipped")
		}

		utils.EOR(c, replyMsg, buildSkippedTracksText(chatID, skippedTitles), nil)
		time.Sleep(1 * time.Second)
	}

	if len(accepted) == 0 {
		utils.EOR(c, replyMsg, F(chatID, "play_all_tracks_skipped"), nil)
		return nil, 0, fmt.Errorf("all tracks skipped")
	}

	availableSlots := config.QueueLimit - len(r.Queue())
	if availableSlots < len(accepted) {
		accepted = accepted[:availableSlots]
		logger.Warnf(
			"Queue full — adding only %d tracks out of requested.",
			availableSlots,
		)
	}

	return accepted, availableSlots, nil
}

func buildSkippedTracksText(chatID int64, skippedTitles []string) string {
	var b strings.Builder
	b.WriteString(F(chatID, "play_multiple_tracks_too_long_header", locales.Arg{
		"count":      len(skippedTitles),
		"limit_mins": config.DurationLimit / 60,
	}))
	b.WriteString("\n")

	for i, title := range skippedTitles {
		if i < 5 {
			b.WriteString(
				F(
					chatID,
					"play_multiple_tracks_too_long_item",
					locales.Arg{"title": title},
				) + "\n",
			)
			continue
		}

		b.WriteString(
			F(
				chatID,
				"play_multiple_tracks_too_long_more",
				locales.Arg{"remaining": len(skippedTitles) - i},
			) + "\n",
		)
		break
	}

	return b.String()
}

func playTracksAndRespond(
	c *td.Client,
	m *td.Message,
	replyMsg *td.Message,
	r *core.RoomState,
	tracks []*state.Track,
	mention string,
	isActive, force bool,
	availableSlots int,
) error {
	chatID := m.ChatID()

	for i, track := range tracks {
		track.Requester = mention

		filePath := ""
		if i == 0 && (!isActive || force) {
			path, err := downloadFirstTrack(c, replyMsg, chatID, mention, track)
			if err != nil {
				return nil
			}
			filePath = path
		}

		if err := playTrackWithRetry(r, track, filePath, force && i == 0, c, replyMsg); err != nil {
			return err
		}

		sendPlayLogs(c, m, track, (isActive && !force) || i > 0)
	}

	return finalizePlayReply(
		c,
		replyMsg,
		r,
		tracks,
		mention,
		isActive,
		force,
		availableSlots,
	)
}

func downloadFirstTrack(
	c *td.Client,
	replyMsg *td.Message,
	chatID int64,
	mention string,
	track *state.Track,
) (string, error) {
	title := utils.EscapeHTML(utils.ShortTitle(track.Title, 25))
	var editOpts *td.EditTextMessageOpts
	if track.Duration > 600 {
		editOpts = &td.EditTextMessageOpts{ReplyMarkup: core.GetCancelKeyboard(chatID)}
	}

	replyMsg, _ = utils.EOR(
		c,
		replyMsg,
		F(chatID, "play_downloading_song", locales.Arg{"title": title}),
		editOpts,
	)

	ctx, cancel := context.WithCancel(context.Background())
	downloads.begin(chatID, cancel, replyMsg)
	defer downloads.finish(chatID)

	path, err := downloadTrack(ctx, c, track, replyMsg, chatID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.EOR(
				c,
				replyMsg,
				F(chatID, "play_download_canceled", locales.Arg{"user": mention}),
				nil,
			)
		} else {
			utils.EOR(c, replyMsg, F(chatID, "play_download_failed"), nil)
		}
		return "", err
	}

	logger.Infof("Downloaded track to %s", path)
	return path, nil
}

func finalizePlayReply(
	c *td.Client,
	replyMsg *td.Message,
	r *core.RoomState,
	tracks []*state.Track,
	mention string,
	isActive bool,
	force bool,
	availableSlots int,
) error {
	chatID := replyMsg.ChatID()
	mainTrack := tracks[0]

	if !isActive || force {
		replyMsg = sendNowPlaying(c, replyMsg, chatID, r, mainTrack)
		if replyMsg != nil {
			r.SetStatusMsg(replyMsg)
		}

		if len(tracks) > 1 {
			replyMsg.ReplyText(
				c,
				buildMultiAddedText(
					chatID,
					len(tracks)-1,
					mention,
					availableSlots,
					len(tracks),
				),
				nil,
			)
		}
		return nil
	}

	if len(tracks) == 1 {
		showTrackMessage(
			c,
			replyMsg,
			chatID,
			r,
			mainTrack,
			buildSingleQueueReply(chatID, r, mainTrack, mention),
			true,
		)
		return nil
	}

	utils.EOR(
		c,
		replyMsg,
		buildMultiAddedText(chatID, len(tracks), mention, availableSlots, len(tracks)),
		&td.EditTextMessageOpts{ParseMode: "HTML"},
	)
	return nil
}

func buildSingleQueueReply(
	chatID int64,
	r *core.RoomState,
	track *state.Track,
	mention string,
) string {
	title := utils.EscapeHTML(utils.ShortTitle(track.Title, 25))

	return F(chatID, "play_added_to_queue_single", locales.Arg{
		"index":    len(r.Queue()),
		"url":      track.URL,
		"title":    title,
		"duration": utils.FormatDuration(track.Duration),
		"by":       mention,
	})
}

func buildMultiAddedText(
	chatID int64,
	count int,
	mention string,
	availableSlots, trackCount int,
) string {
	var b strings.Builder
	b.WriteString(
		F(
			chatID,
			"play_added_multiple_header",
			locales.Arg{"count": count, "user": mention},
		),
	)
	b.WriteString("\n\n")

	if availableSlots <= trackCount {
		b.WriteString(F(chatID, "play_queue_limit_hint"))
		b.WriteString("\n")
	}

	b.WriteString(F(chatID, "play_queue_view_hint"))
	return b.String()
}

func playTrackWithRetry(
	r *core.RoomState,
	track *state.Track,
	filePath string,
	force bool,
	c *td.Client,
	replyMsg *td.Message,
) error {
	for attempt := 1; attempt <= playMaxRetries; attempt++ {
		if r.IsDestroyed() {
			logger.Info("Room destroyed during retry, aborting")
			replyMsg.Delete(c, true)
			return nil
		}

		err := r.Play(track, filePath, force)
		if err == nil {
			if attempt > 1 {
				logger.Info(
					"Successfully played after retry attempt " + utils.IntToStr(attempt),
				)
			}
			return nil
		}

		handled, stopErr := handlePlayAttemptError(err, attempt, c, replyMsg, r)
		if handled {
			if stopErr != nil {
				return stopErr
			}
			continue
		}

		if attempt == playMaxRetries {
			logger.Error(
				"❌ Failed to play after " + utils.IntToStr(
					playMaxRetries,
				) + " attempts. Error: " + err.Error(),
			)
			utils.EOR(
				c,
				replyMsg,
				F(replyMsg.ChatID(), "play_failed", locales.Arg{"error": err.Error()}),
				nil,
			)
			return err
		}

		logger.Error(
			"Unexpected error occurred. Retrying... (attempt " + utils.IntToStr(
				attempt,
			) + "): " + err.Error(),
		)
	}

	return nil
}

func handlePlayAttemptError(
	err error,
	attempt int,
	c *td.Client,
	replyMsg *td.Message,
	room *core.RoomState,
) (bool, error) {
	if wait := getFloodWait(err); wait > 0 {
		logger.Error(
			"FloodWait detected (" + strconv.Itoa(
				wait,
			) + "s). Retrying... (attempt " + utils.IntToStr(
				attempt,
			) + ")",
		)
		time.Sleep(time.Duration(wait) * time.Second)
		return true, nil
	}

	if errors.Is(err, ubot.ErrConnectionTimeout) {
		logger.Error("Voice connection timeout. Stopping call session...")
		utils.EOR(c, replyMsg, F(replyMsg.ChatID(), "err_connection_timeout"), nil)
		core.DeleteRoom(room.ID)
		return true, nil
	}

	if strings.Contains(err.Error(), "Streaming is not supported when using RTMP") {
		logger.Error("RTMP/live-stream voice chat detected, cannot play. Cleaning up...")
		core.DeleteRoom(room.ID)
		utils.EOR(c, replyMsg, F(replyMsg.ChatID(), "rtmp_play_unsupported"), nil)
		return true, nil
	}

	if strings.Contains(err.Error(), "group call") &&
		strings.Contains(err.Error(), "is closed") {
		markVoiceChatInactive(room.ID)
		utils.EOR(c, replyMsg, F(replyMsg.ChatID(), "err_no_active_voicechat"), nil)
		return true, nil
	}

	if strings.Contains(err.Error(), "GROUPCALL_INVALID") {
		logger.Error("GROUPCALL_INVALID err occurred. Returning...")
		core.DeleteRoom(room.ID)
		utils.EOR(c, replyMsg, F(replyMsg.ChatID(), "play_unable"), nil)
		return true, nil
	}

	if strings.Contains(err.Error(), "INTERDC_X_CALL_ERROR") {
		logger.Error(
			"INTERDC_X_CALL_ERROR occurred. Retrying... (attempt " + utils.IntToStr(
				attempt,
			) + ")",
		)
		time.Sleep(2 * time.Second)
		return true, nil
	}

	return false, nil
}

// markVoiceChatInactive records that no active voice chat is running in the
// room, as reported by u-bot.play.
func markVoiceChatInactive(roomID int64) {
	cs, err := core.GetChatState(roomID)
	if err != nil {
		logger.Errorf("failed to get chat state to mark voice chat inactive: %v", err)
		return
	}
	cs.SetVoiceChatActive(false)
}

// getFloodWait returns the retry-after seconds for a flood-wait error,
// whether surfaced as a TDLib error or an MTProto/ntgcalls error.
func getFloodWait(err error) int {
	var tde *td.Error
	if errors.As(err, &tde) {
		if wait := tde.GetRetryAfter(); wait > 0 {
			return wait
		}
	}

	msg := err.Error()
	if _, after, ok := strings.Cut(msg, "FLOOD_WAIT_"); ok {
		if wait, cerr := strconv.Atoi(after); cerr == nil {
			return wait
		}
	}
	return 0
}

type msgFn func(chatID int64, err error) string

var errMessageMap = map[error]msgFn{
	core.ErrAdminPermissionRequired: func(chatID int64, _ error) string {
		return F(chatID, "err_admin_permission_required")
	},
	core.ErrAssistantNotAvailable: func(chatID int64, e error) string {
		return F(chatID, "err_assistant_get_failed", locales.Arg{"error": e.Error()})
	},
	core.ErrInviteRequestSent: func(chatID int64, _ error) string {
		return F(chatID, "err_assistant_join_request_sent")
	},
	core.ErrAssistantInviteLinkFetch: func(chatID int64, e error) string {
		return F(
			chatID,
			"err_assistant_invite_link_fetch",
			locales.Arg{"error": e.Error()},
		)
	},
	core.ErrJoinFailed: func(chatID int64, e error) string {
		return F(chatID, "err_assistant_invite_failed", locales.Arg{"error": e.Error()})
	},
	core.ErrStateFetchFailed: func(chatID int64, e error) string {
		return F(chatID, "err_fetch_failed", locales.Arg{"error": e.Error()})
	},
}

func getErrorMessage(chatID int64, err error) string {
	if err == nil {
		return ""
	}

	for key, fn := range errMessageMap {
		if errors.Is(err, key) {
			return fn(chatID, err)
		}
	}

	return F(chatID, "err_unknown", locales.Arg{"error": err.Error()})
}

// downloadTrack and safeGetTracks re-raise panics on failure.
func safeGetTracks(
	c *td.Client,
	m, replyMsg *td.Message,
	chatID int64,
	video bool,
) (tracks []*state.Track, err error) {
	defer func() {
		if r := recover(); r != nil {
			utils.EOR(c, replyMsg, F(chatID, "err_fetch_tracks"), nil)
			panic(r)
		}
	}()

	return platforms.GetTracks(c, m, video)
}

// downloadTrack runs the platform download, showing an internal-error message
// and re-raising any panic so the caller can catch it at the top level.
func downloadTrack(
	ctx context.Context,
	c *td.Client,
	track *state.Track,
	replyMsg *td.Message,
	chatID int64,
) (path string, err error) {
	defer func() {
		if r := recover(); r != nil {
			utils.EOR(c, replyMsg, F(chatID, "err_download_internal"), nil)
			panic(r)
		}
	}()

	return platforms.Download(ctx, track, replyMsg)
}
