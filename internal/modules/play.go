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
	"strings"
	"time"

	td "github.com/AshokShau/gotdbot"
	tg "github.com/amarnathcjd/gogram/telegram"

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
	helpTexts["/cfplay"] = helpTexts["/fplay"]
	helpTexts["/fcplay"] = helpTexts["/fplay"]
	helpTexts["/vcplay"] = helpTexts["/vplay"]
	helpTexts["/cvplay"] = helpTexts["/vplay"]
}

func playCmd(o playOpts) func(*td.Client, *td.Message) error {
	return func(c *td.Client, m *td.Message) error { return handlePlay(c, m, o) }
}

var (
	playHandler   = playCmd(playOpts{})
	fplayHandler  = playCmd(playOpts{Force: true})
	cplayHandler  = playCmd(playOpts{CPlay: true})
	cfplayHandler = playCmd(playOpts{Force: true, CPlay: true})
	vplayHandler  = playCmd(playOpts{Video: true})
	fvplayHandler = playCmd(playOpts{Force: true, Video: true})
	vcplayHandler = playCmd(playOpts{CPlay: true, Video: true})

	fvcplayHandler = playCmd(playOpts{Force: true, CPlay: true, Video: true})
)

// playReq carries one /play request through its steps. A step that cannot
// continue returns an error whose text is shown to the user as is.
type playReq struct {
	playOpts

	c      *td.Client
	m      *td.Message
	chatID int64

	room   *core.Room
	status *td.Message // the "searching" message, later the player message
	tracks []*state.Track
	by     string // mention of the requester
	active bool   // something was already playing
	full   bool   // the queue limit cut the request short
}

func handlePlay(c *td.Client, m *td.Message, o playOpts) error {
	if !isSuperGroup(c, m) {
		return nil
	}

	p := &playReq{playOpts: o, c: c, m: m, chatID: m.ChatID()}
	for _, step := range []func() error{p.allow, p.prepare, p.fetch, p.admit, p.start} {
		if err := step(); err != nil {
			p.fail(err)
			return nil
		}
	}
	return nil
}

func (p *playReq) errf(key string, args ...locales.Arg) error {
	return errors.New(F(p.chatID, key, args...))
}

// show edits the status message, or replies if there is none yet.
func (p *playReq) show(text string) {
	if p.status != nil {
		utils.EOR(p.c, p.status, text, nil)
		return
	}
	p.m.ReplyText(p.c, text, nil)
}

func (p *playReq) fail(err error) {
	if errors.Is(err, core.ErrRoomDestroyed) {
		if p.status != nil {
			p.status.Delete(p.c, &td.DeleteMessagesOpts{Revoke: true})
		}
		return
	}
	p.show(err.Error())
}

// allow checks that the sender may use /play in this chat.
func (p *playReq) allow() error {
	adminsOnly, _ := database.PlayModeAdminsOnly(p.chatID)
	if !adminsOnly {
		return nil
	}
	if ok, err := utils.IsChatAdmin(p.c, p.chatID, p.m.SenderID()); err == nil && ok {
		return nil
	}
	if ok, _ := database.IsAuthorized(p.chatID, p.m.SenderID()); ok {
		return nil
	}
	return p.errf("playmode_restricted")
}

// prepare resolves the room and sends the "searching" message.
func (p *playReq) prepare() (err error) {
	if p.room, err = getEffectiveRoom(p.chatID, p.CPlay); err != nil {
		return err
	}
	p.room.Refresh()

	if config.QueueLimit > 0 && p.room.QueueLen() >= config.QueueLimit {
		return p.errf("queue_limit_reached", locales.Arg{"limit": config.QueueLimit})
	}

	query := extractPlayQuery(p.m.Text())
	if query == "" && p.m.ReplyToMessageID() == 0 {
		return p.errf("no_song_query", locales.Arg{"cmd": getCommand(p.m)})
	}

	text := F(p.chatID, "searching")
	if query != "" {
		text = F(p.chatID, "searching_query", locales.Arg{"query": utils.EscapeHTML(query)})
	}
	p.status, err = p.m.ReplyText(p.c, text, nil)
	return err
}

// fetch looks the tracks up and makes sure the assistant can join the call.
func (p *playReq) fetch() (err error) {
	if p.tracks, err = platforms.GetTracks(p.c, p.m, p.Video); err != nil {
		return err
	}
	if len(p.tracks) == 0 {
		return p.errf("no_song_found")
	}

	cs, err := core.ChatStateFor(p.room.ChatID)
	if err != nil {
		utils.Log.Errorf("Error getting chat state: %v", err)
		return p.wrap(err)
	}
	if err := p.joinCall(cs); err != nil {
		return err
	}

	p.active = p.room.Active()
	return nil
}

func (p *playReq) joinCall(cs *core.ChatState) error {
	st, err := cs.Status()
	if err != nil {
		utils.Log.Errorf("Error checking voicechat state: %v", err)
		return p.wrap(err)
	}

	// A banned assistant is checked first: the voice chat state is unknown
	// while it is banned.
	switch {
	case st.Presence == core.PresenceBanned:
		self := cs.Assistant().Self
		return p.errf("err_assistant_banned", locales.Arg{
			"user": mentionHTML(self.FirstName+" "+self.LastName, self.ID),
			"id":   utils.IntToStr(self.ID),
		})
	case st.Voice == core.VoiceInactive:
		return p.errf("err_no_active_voicechat")
	case st.Presence == core.PresenceJoined:
		return nil
	}

	if err := cs.Join(); err != nil {
		utils.Log.Errorf("Error joining assistant: %v", err)
		return p.wrap(err)
	}
	time.Sleep(time.Second)
	return nil
}

// wrap turns an internal error into the message shown to the user.
func (p *playReq) wrap(err error) error {
	return errors.New(getErrorMessage(p.chatID, err))
}

// admit applies the duplicate, duration and queue limits to the tracks.
func (p *playReq) admit() error {
	if len(p.tracks) == 1 && !p.Force && isTrackInQueue(p.room, p.tracks[0]) {
		return p.errf("play_already_in_queue", locales.Arg{
			"title": utils.EscapeHTML(utils.ShortTitle(p.tracks[0].Title, 35)),
		})
	}
	if config.QueueLimit == 0 && p.active && !p.Force {
		return p.errf("play_already_playing")
	}

	var kept []*state.Track
	var tooLong []string
	for _, t := range p.tracks {
		if t.Duration > config.DurationLimit {
			tooLong = append(tooLong, utils.EscapeHTML(utils.ShortTitle(t.Title, 35)))
			continue
		}
		kept = append(kept, t)
	}

	switch {
	case len(tooLong) == 0:
	case len(kept) == 0 && len(p.tracks) == 1:
		return p.errf("play_single_track_too_long", locales.Arg{
			"limit_mins": utils.FormatDuration(config.DurationLimit),
			"title":      tooLong[0],
		})
	case len(kept) == 0:
		return p.errf("play_all_tracks_skipped")
	default:
		p.show(buildSkippedTracksText(p.chatID, tooLong))
		time.Sleep(time.Second)
	}
	p.tracks = kept

	if config.QueueLimit >= 0 {
		free := max(config.QueueLimit-p.room.QueueLen(), 0)
		if free == 0 && (!p.active || p.Force) {
			free = 1 // the track that starts playing needs no queue slot
		}
		p.full = config.QueueLimit > 0 && free <= len(p.tracks)
		p.tracks = p.tracks[:min(free, len(p.tracks))]
	}
	return nil
}

// start plays or queues the tracks and answers the user.
func (p *playReq) start() error {
	sender, _ := p.m.GetUser(p.c)
	p.by = mention(sender, p.m.SenderID())
	starts := !p.active || p.Force

	for i, t := range p.tracks {
		t.Requester = p.by

		path := ""
		if i == 0 && starts {
			var err error
			if path, err = p.download(t); err != nil {
				return err
			}
		}
		if err := p.playTrack(t, path, p.Force && i == 0); err != nil {
			return err
		}
		sendPlayLogs(p.c, p.m, t, i > 0 || !starts)
	}

	p.reply(starts)
	return nil
}

func (p *playReq) reply(started bool) {
	first, n := p.tracks[0], len(p.tracks)

	switch {
	case started:
		msg := sendNowPlaying(p.c, p.status, p.chatID, p.room, first)
		if msg == nil {
			return
		}
		p.room.SetStatusMsg(msg)
		if n > 1 {
			msg.ReplyText(p.c, p.addedText(n-1), nil)
		}
	case n == 1:
		showTrackMessage(p.c, p.status, p.chatID, p.room, first,
			buildSingleQueueReply(p.chatID, p.room, first, p.by), true)
	default:
		utils.EOR(p.c, p.status, p.addedText(n), &td.EditMessageTextOpts{ParseMode: "HTML"})
	}
}

func (p *playReq) addedText(count int) string {
	var b strings.Builder
	b.WriteString(F(p.chatID, "play_added_multiple_header", locales.Arg{"count": count, "user": p.by}))
	b.WriteString("\n\n")
	if p.full {
		b.WriteString(F(p.chatID, "play_queue_limit_hint"))
		b.WriteString("\n")
	}
	b.WriteString(F(p.chatID, "play_queue_view_hint"))
	return b.String()
}

// download fetches the track that is about to start, with a cancel button for long ones.
func (p *playReq) download(t *state.Track) (string, error) {
	var opts *td.EditMessageTextOpts
	if t.Duration > 600 {
		opts = &td.EditMessageTextOpts{ReplyMarkup: core.GetCancelKeyboard(p.chatID)}
	}
	title := utils.EscapeHTML(utils.ShortTitle(t.Title, 25))
	if msg, _ := utils.EOR(p.c, p.status, F(p.chatID, "play_downloading_song", locales.Arg{"title": title}), opts); msg != nil {
		p.status = msg
	}

	ctx, cancel := context.WithCancel(context.Background())
	beginDownload(p.chatID, cancel, p.status)
	defer finishDownload(p.chatID)

	path, err := platforms.Download(ctx, t, p.status)
	switch {
	case errors.Is(err, context.Canceled):
		return "", p.errf("play_download_canceled", locales.Arg{"user": p.by})
	case err != nil:
		return "", p.errf("play_download_failed")
	}

	utils.Log.Infof("Downloaded track to %s", path)
	return path, nil
}

// playTrack starts a track, retrying the errors that are worth a retry.
func (p *playReq) playTrack(t *state.Track, path string, force bool) error {
	var err error
	for attempt := 1; attempt <= playMaxRetries; attempt++ {
		if err = p.room.Play(t, path, force); err == nil {
			return nil
		}
		utils.Log.Errorf("play attempt %d/%d failed: %v", attempt, playMaxRetries, err)

		wait, stop := p.onPlayError(err)
		if stop != nil {
			return stop
		}
		time.Sleep(wait)
	}
	return p.errf("play_failed", locales.Arg{"error": err.Error()})
}

// onPlayError sorts a failed Room.Play: either wait and retry, or stop with
// the error to show.
func (p *playReq) onPlayError(err error) (wait time.Duration, stop error) {
	msg := err.Error()
	flood := getFloodWait(err)

	switch {
	case errors.Is(err, core.ErrRoomDestroyed):
		return 0, err
	case flood > 0:
		return time.Duration(flood) * time.Second, nil
	case strings.Contains(msg, "INTERDC_X_CALL_ERROR"):
		return 2 * time.Second, nil
	case strings.Contains(msg, "GROUPCALL_INVALID"):
		return 0, p.closeRoom("play_unable")
	case errors.Is(err, ubot.ErrConnectionTimeout):
		return 0, p.closeRoom("err_connection_timeout")
	case strings.Contains(msg, "Streaming is not supported when using RTMP"):
		return 0, p.closeRoom("rtmp_play_unsupported")
	case strings.Contains(msg, "group call") && strings.Contains(msg, "is closed"):
		markVoiceChatInactive(p.room.ChatID)
		return 0, p.errf("err_no_active_voicechat")
	}
	return 0, nil
}

func (p *playReq) closeRoom(key string) error {
	core.Rooms.Delete(p.room.ChatID)
	return p.errf(key)
}

func extractPlayQuery(text string) string {
	parts := strings.SplitN(text, " ", 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func isTrackInQueue(r *core.Room, t *state.Track) bool {
	for _, qt := range r.Queue() {
		if qt.URL == t.URL || qt.ID == t.ID {
			return true
		}
	}
	return false
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

func buildSingleQueueReply(
	chatID int64,
	r *core.Room,
	track *state.Track,
	mention string,
) string {
	title := utils.EscapeHTML(utils.ShortTitle(track.Title, 25))

	return F(chatID, "play_added_to_queue_single", locales.Arg{
		"index":    r.QueueLen(),
		"url":      track.URL,
		"title":    title,
		"duration": utils.FormatDuration(track.Duration),
		"by":       mention,
	})
}

// markVoiceChatInactive records that no active voice chat is running in the
// room, as reported by u-bot.play.
func markVoiceChatInactive(roomID int64) {
	cs, err := core.ChatStateFor(roomID)
	if err != nil {
		utils.Log.Errorf("failed to get chat state to mark voice chat inactive: %v", err)
		return
	}
	cs.SetVoice(core.VoiceInactive)
}

// getFloodWait returns the retry-after seconds for a flood-wait error,
// whether surfaced as a TDLib error or an MTProto/ntgcalls error.
func getFloodWait(err error) int {
	if tde, ok := errors.AsType[*td.Error](err); ok {
		if wait := tde.GetRetryAfter(); wait > 0 {
			return wait
		}
	}

	return tg.GetFloodWait(err)
}

func getErrorMessage(chatID int64, err error) string {
	if err == nil {
		return ""
	}

	switch {
	case errors.Is(err, core.ErrAdminPermissionRequired):
		return F(chatID, "err_admin_permission_required")
	case errors.Is(err, core.ErrAssistantNotAvailable):
		return F(chatID, "err_assistant_get_failed", locales.Arg{"error": err.Error()})
	case errors.Is(err, core.ErrInviteRequestSent):
		return F(chatID, "err_assistant_join_request_sent")
	case errors.Is(err, core.ErrAssistantInviteLinkFetch):
		return F(chatID, "err_assistant_invite_link_fetch", locales.Arg{"error": err.Error()})
	case errors.Is(err, core.ErrJoinFailed):
		return F(chatID, "err_assistant_invite_failed", locales.Arg{"error": err.Error()})
	case errors.Is(err, core.ErrStateFetchFailed):
		return F(chatID, "err_fetch_failed", locales.Arg{"error": err.Error()})
	default:
		if fw, ok := errors.AsType[*core.FloodWaitError](err); ok {
			return floodWaitMessage(chatID, fw.Seconds)
		}
		return F(chatID, "err_unknown", locales.Arg{"error": err.Error()})
	}
}

// floodWaitMessage tells the user how long they must wait when the assistant
// is flood-waiting and no other assistant can join in its place.
func floodWaitMessage(chatID int64, wait int) string {
	if wait >= 60 {
		return F(chatID, "flood_minutes", locales.Arg{"duration": utils.FormatDuration(wait)})
	}
	return F(chatID, "flood_seconds", locales.Arg{"duration": wait})
}
