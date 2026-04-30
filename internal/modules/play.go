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

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
	state "main/internal/core/models"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/platforms"
	"main/internal/utils"
	"main/ubot"
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

func playHandler(m *tg.NewMessage) error   { return handlePlay(m, &playOpts{}) }
func fplayHandler(m *tg.NewMessage) error  { return handlePlay(m, &playOpts{Force: true}) }
func cfplayHandler(m *tg.NewMessage) error { return handlePlay(m, &playOpts{Force: true, CPlay: true}) }
func vplayHandler(m *tg.NewMessage) error  { return handlePlay(m, &playOpts{Video: true}) }
func fvplayHandler(m *tg.NewMessage) error { return handlePlay(m, &playOpts{Force: true, Video: true}) }
func vcplayHandler(m *tg.NewMessage) error { return handlePlay(m, &playOpts{CPlay: true, Video: true}) }
func fvcplayHandler(m *tg.NewMessage) error {
	return handlePlay(m, &playOpts{Force: true, CPlay: true, Video: true})
}
func cplayHandler(m *tg.NewMessage) error { return handlePlay(m, &playOpts{CPlay: true}) }

func handlePlay(m *tg.NewMessage, opts *playOpts) error {
	chatID := m.ChannelID()

	if !canUsePlayCommand(m, chatID) {
		m.Reply(F(chatID, "playmode_restricted"))
		return tg.ErrEndGroup
	}

	room, searchMsg, err := prepareRoomAndSearchMessage(m, opts.CPlay)
	if err != nil {
		return tg.ErrEndGroup
	}

	tracks, isActive, err := fetchTracksAndCheckStatus(m, searchMsg, room, opts.Video)
	if err != nil {
		return tg.ErrEndGroup
	}

	if len(tracks) == 1 && !opts.Force {
		if isTrackInQueue(room, tracks[0]) {
			utils.EOR(searchMsg, F(m.ChannelID(), "play_already_in_queue", locales.Arg{
				"title": utils.EscapeHTML(utils.ShortTitle(tracks[0].Title, 35)),
			}))
			return tg.ErrEndGroup
		}
	}

	tracks, availableSlots, err := filterAndTrimTracks(searchMsg, room, tracks)
	if err != nil {
		return tg.ErrEndGroup
	}

	mention := utils.MentionHTML(m.Sender)
	if err := playTracksAndRespond(m, searchMsg, room, tracks, mention, isActive, opts.Force, availableSlots); err != nil {
		return err
	}

	return tg.ErrEndGroup
}

func canUsePlayCommand(m *tg.NewMessage, chatID int64) bool {
	adminsOnly, _ := database.PlayModeAdminsOnly(chatID)
	if !adminsOnly {
		return true
	}

	isAdmin, err := utils.IsChatAdmin(m.Client, chatID, m.SenderID())
	if err == nil && isAdmin {
		return true
	}

	isAuth, _ := database.IsAuthorized(chatID, m.SenderID())
	return isAuth
}

func prepareRoomAndSearchMessage(
	m *tg.NewMessage,
	cplay bool,
) (*core.RoomState, *tg.NewMessage, error) {
	room, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return nil, nil, err
	}

	chatID := m.ChannelID()
	room.Parse()

	if len(room.Queue()) >= config.QueueLimit {
		m.Reply(F(chatID, "queue_limit_reached", locales.Arg{"limit": config.QueueLimit}))
		return nil, nil, fmt.Errorf("queue limit reached")
	}

	query := extractPlayQuery(m.Text())
	if query == "" && !m.IsReply() {
		m.Reply(F(chatID, "no_song_query", locales.Arg{"cmd": getCommand(m)}))
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

	replyMsg, err := m.Reply(statusText)
	if err != nil {
		gologging.ErrorF("Failed to send searching message: %v", err)
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
	m *tg.NewMessage,
	replyMsg *tg.NewMessage,
	r *core.RoomState,
	video bool,
) ([]*state.Track, bool, error) {
	tracks, err := safeGetTracks(m, replyMsg, m.ChannelID(), video)
	if err != nil {
		utils.EOR(replyMsg, err.Error())
		return nil, false, err
	}
	if len(tracks) == 0 {
		utils.EOR(replyMsg, F(m.ChannelID(), "no_song_found"))
		return nil, false, fmt.Errorf("no tracks found")
	}

	chatState, err := core.GetChatState(r.ID())
	if err != nil {
		gologging.ErrorF("Error getting chat state: %v", err)
		utils.EOR(replyMsg, getErrorMessage(m.ChannelID(), err))
		return nil, false, err
	}

	if err := ensureVoiceChatReady(m.ChannelID(), replyMsg, chatState); err != nil {
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
	chatID int64,
	replyMsg *tg.NewMessage,
	cs *core.ChatState,
) error {
	snap, err := cs.Snapshot(false)
	if err != nil {
		gologging.ErrorF("Error checking voicechat state: %v", err)
		utils.EOR(replyMsg, getErrorMessage(chatID, err))
		return err
	}
	if !snap.VoiceChatActive {
		err := fmt.Errorf("no active voice chat")
		utils.EOR(replyMsg, F(chatID, "err_no_active_voicechat"))
		return err
	}

	if snap.AssistantBanned {
		err := fmt.Errorf("assistant banned")
		utils.EOR(replyMsg, F(chatID, "err_assistant_banned", locales.Arg{
			"user": utils.MentionHTML(cs.Assistant.Self),
			"id":   utils.IntToStr(cs.Assistant.Self.ID),
		}))
		return err
	}

	if snap.AssistantPresent {
		return nil
	}

	username := ""
	if replyMsg.Channel != nil {
		username = replyMsg.Channel.Username
	}
	if err := cs.EnsureAssistantJoined(username); err != nil {
		gologging.ErrorF("Error joining assistant: %v", err)
		utils.EOR(replyMsg, getErrorMessage(chatID, err))
		return err
	}

	time.Sleep(1 * time.Second)
	return nil
}

func filterAndTrimTracks(
	replyMsg *tg.NewMessage,
	r *core.RoomState,
	tracks []*state.Track,
) ([]*state.Track, int, error) {
	chatID := replyMsg.ChannelID()
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
			utils.EOR(replyMsg, F(chatID, "play_single_track_too_long", locales.Arg{
				"limit_mins": utils.FormatDuration(config.DurationLimit),
				"title":      skippedTitles[0],
			}))
			return nil, 0, fmt.Errorf("single long track skipped")
		}

		utils.EOR(replyMsg, buildSkippedTracksText(chatID, skippedTitles))
		time.Sleep(1 * time.Second)
	}

	if len(accepted) == 0 {
		utils.EOR(replyMsg, F(chatID, "play_all_tracks_skipped"))
		return nil, 0, fmt.Errorf("all tracks skipped")
	}

	availableSlots := config.QueueLimit - len(r.Queue())
	if availableSlots < len(accepted) {
		accepted = accepted[:availableSlots]
		gologging.WarnF(
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
	m *tg.NewMessage,
	replyMsg *tg.NewMessage,
	r *core.RoomState,
	tracks []*state.Track,
	mention string,
	isActive, force bool,
	availableSlots int,
) error {
	chatID := m.ChannelID()

	for i, track := range tracks {
		track.Requester = mention

		filePath := ""
		if i == 0 && (!isActive || force) {
			path, err := downloadFirstTrack(m, replyMsg, chatID, mention, track)
			if err != nil {
				return tg.ErrEndGroup
			}
			filePath = path
		}

		if err := playTrackWithRetry(r, track, filePath, force && i == 0, replyMsg); err != nil {
			return err
		}

		sendPlayLogs(m, track, (isActive && !force) || i > 0)
	}

	return finalizePlayReply(
		m,
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
	m *tg.NewMessage,
	replyMsg *tg.NewMessage,
	chatID int64,
	mention string,
	track *state.Track,
) (string, error) {
	title := utils.EscapeHTML(utils.ShortTitle(track.Title, 25))
	var opt *tg.SendOptions
	if track.Duration > 600 {
		opt = &tg.SendOptions{ReplyMarkup: core.GetCancelKeyboard(chatID)}
	}

	replyMsg, _ = utils.EOR(
		replyMsg,
		F(chatID, "play_downloading_song", locales.Arg{"title": title}),
		opt,
	)

	ctx, cancel := context.WithCancel(context.Background())
	downloadCancels[chatID] = cancel
	defer func() {
		if _, ok := downloadCancels[chatID]; ok {
			delete(downloadCancels, chatID)
			cancel()
		}
	}()

	path, err := safeDownload(ctx, track, replyMsg, chatID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.EOR(
				replyMsg,
				F(chatID, "play_download_canceled", locales.Arg{"user": mention}),
			)
		} else {
			utils.EOR(replyMsg, F(chatID, "play_download_failed", locales.Arg{
				"title": title,
				"error": utils.EscapeHTML(err.Error()),
			}))
		}
		return "", err
	}

	gologging.InfoF("Downloaded track to %s", path)
	return path, nil
}

func finalizePlayReply(
	m *tg.NewMessage,
	replyMsg *tg.NewMessage,
	r *core.RoomState,
	tracks []*state.Track,
	mention string,
	isActive bool,
	force bool,
	availableSlots int,
) error {
	chatID := m.ChannelID()
	mainTrack := tracks[0]

	if !isActive || force {
		msg, opts := buildNowPlayingReply(chatID, r, mainTrack, mention)
		replyMsg, _ = utils.EOR(replyMsg, msg, opts)
		r.SetStatusMsg(replyMsg)

		if len(tracks) > 1 {
			replyMsg.Respond(
				buildMultiAddedText(
					chatID,
					len(tracks)-1,
					mention,
					availableSlots,
					len(tracks),
				),
			)
		}
		return nil
	}

	if len(tracks) == 1 {
		msg, opts := buildSingleQueueReply(chatID, r, mainTrack, mention)
		utils.EOR(replyMsg, msg, opts)
		return nil
	}

	utils.EOR(
		replyMsg,
		buildMultiAddedText(chatID, len(tracks), mention, availableSlots, len(tracks)),
	)
	return nil
}

func buildNowPlayingReply(
	chatID int64,
	r *core.RoomState,
	track *state.Track,
	mention string,
) (string, *tg.SendOptions) {
	title := utils.EscapeHTML(utils.ShortTitle(track.Title, 25))
	opt := &tg.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: core.GetPlayMarkup(chatID, r, false),
	}
	if track.Artwork != "" && shouldShowThumb(chatID) {
		opt.Media = utils.CleanURL(track.Artwork)
	}

	msg := F(chatID, "stream_now_playing", locales.Arg{
		"url":      track.URL,
		"title":    title,
		"duration": utils.FormatDuration(track.Duration),
		"by":       mention,
	})
	return msg, opt
}

func buildSingleQueueReply(
	chatID int64,
	r *core.RoomState,
	track *state.Track,
	mention string,
) (string, *tg.SendOptions) {
	title := utils.EscapeHTML(utils.ShortTitle(track.Title, 25))
	opt := &tg.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: core.GetPlayMarkup(chatID, r, true),
	}
	if track.Artwork != "" && shouldShowThumb(chatID) {
		opt.Media = utils.CleanURL(track.Artwork)
	}

	msg := F(chatID, "play_added_to_queue_single", locales.Arg{
		"index":    len(r.Queue()),
		"url":      track.URL,
		"title":    title,
		"duration": utils.FormatDuration(track.Duration),
		"by":       mention,
	})
	return msg, opt
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
	replyMsg *tg.NewMessage,
) error {
	for attempt := 1; attempt <= playMaxRetries; attempt++ {
		if r.IsDestroyed() {
			gologging.Info("Room destroyed during retry, aborting")
			replyMsg.Delete()
			return tg.ErrEndGroup
		}

		err := r.Play(track, filePath, force)
		if err == nil {
			if attempt > 1 {
				gologging.Info(
					"Successfully played after retry attempt " + utils.IntToStr(attempt),
				)
			}
			return nil
		}

		handled, stopErr := handlePlayAttemptError(err, attempt, replyMsg, r)
		if handled {
			if stopErr != nil {
				return stopErr
			}
			continue
		}

		if attempt == playMaxRetries {
			gologging.Error(
				"❌ Failed to play after " + utils.IntToStr(
					playMaxRetries,
				) + " attempts. Error: " + err.Error(),
			)
			utils.EOR(
				replyMsg,
				F(replyMsg.ChannelID(), "play_failed", locales.Arg{"error": err.Error()}),
			)
			return err
		}

		gologging.Error(
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
	replyMsg *tg.NewMessage,
	room *core.RoomState,
) (bool, error) {
	if wait := tg.GetFloodWait(err); wait > 0 {
		gologging.Error(
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
		gologging.Error("Voice connection timeout. Stopping call session...")
		utils.EOR(replyMsg, F(replyMsg.ChannelID(), "err_connection_timeout"))
		core.DeleteRoom(room.ID())
		return true, tg.ErrEndGroup
	}

	if strings.Contains(err.Error(), "Streaming is not supported when using RTMP") {
		utils.EOR(replyMsg, F(replyMsg.ChannelID(), "rtmp_streaming_not_supported"))
		core.DeleteRoom(room.ID())
		return true, tg.ErrEndGroup
	}

	if strings.Contains(err.Error(), "group call") &&
		strings.Contains(err.Error(), "is closed") {
		utils.EOR(replyMsg, F(replyMsg.ChannelID(), "err_no_active_voicechat"))
		return true, tg.ErrEndGroup
	}

	if tg.MatchError(err, "GROUPCALL_INVALID") {
		gologging.Error("GROUPCALL_INVALID err occurred. Returning...")
		core.DeleteRoom(room.ID())
		utils.EOR(replyMsg, F(replyMsg.ChannelID(), "play_unable"))
		return true, tg.ErrEndGroup
	}

	if tg.MatchError(err, "INTERDC_X_CALL_ERROR") {
		gologging.Error(
			"INTERDC_X_CALL_ERROR occurred. Retrying... (attempt " + utils.IntToStr(
				attempt,
			) + ")",
		)
		time.Sleep(2 * time.Second)
		return true, nil
	}

	return false, nil
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

// Both safeDownload and safeGetTracks re-raise panic because all command
// handlers are wrapped by SafeMessageHandler, which catches panics and sends
// the debug trace to the logger and the owner.
func safeGetTracks(
	m, replyMsg *tg.NewMessage,
	chatID int64,
	video bool,
) (tracks []*state.Track, err error) {
	defer func() {
		if r := recover(); r != nil {
			utils.EOR(replyMsg, F(chatID, "err_fetch_tracks"))
			panic(r)
		}
	}()

	return platforms.GetTracks(m, video)
}

func safeDownload(
	ctx context.Context,
	track *state.Track,
	replyMsg *tg.NewMessage,
	chatID int64,
) (path string, err error) {
	defer func() {
		if r := recover(); r != nil {
			utils.EOR(replyMsg, F(chatID, "err_download_internal"))
			panic(r)
		}
	}()

	return platforms.Download(ctx, track, replyMsg)
}
