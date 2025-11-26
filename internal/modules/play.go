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
package modules

import (
	"context"
	"errors"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/platforms"
	"main/internal/state"
	"main/internal/utils"
)

const playMaxRetries = 3

func channelPlayHandler(m *telegram.NewMessage) error {
	m.Reply("⚠️ This handler is deprecated. Use <code><a>/cplay --set channel_id </code></a> to set your channel for playback.")
	return telegram.EndGroup
}

func playHandler(m *telegram.NewMessage) error {
	return handlePlay(m, false, false)
}

func fplayHandler(m *telegram.NewMessage) error {
	return handlePlay(m, true, false)
}

func cplayHandler(m *telegram.NewMessage) error {
	args := strings.Fields(m.Text())
	chatID := m.ChannelID()

	if len(args) > 1 && args[1] == "--set" {
		if len(args) < 3 {
			m.Reply(
				F(chatID, "cplay_usage"),
				&telegram.SendOptions{ParseMode: "HTML"},
			)
			return telegram.EndGroup
		}

		cplayIDStr := args[2]
		cplayID, err := strconv.ParseInt(cplayIDStr, 10, 64)
		if err != nil {
			m.Reply(
				F(chatID, "cplay_invalid_chat_id"),
				&telegram.SendOptions{ParseMode: "HTML"},
			)
			return telegram.EndGroup
		}

		peer, err := m.Client.ResolvePeer(cplayID)
		if err != nil {
			m.Reply(
				F(chatID, "cplay_resolve_peer_fail"),
				&telegram.SendOptions{ParseMode: "HTML"},
			)
			return telegram.EndGroup
		}

		chPeer, ok := peer.(*telegram.InputPeerChannel)
		if !ok {
			m.Reply(
				F(chatID, "cplay_invalid_target"),
				&telegram.SendOptions{ParseMode: "HTML"},
			)
			return telegram.EndGroup
		}

		fullChat, err := m.Client.ChannelsGetFullChannel(&telegram.InputChannelObj{
			ChannelID:  chPeer.ChannelID,
			AccessHash: chPeer.AccessHash,
		})
		if err != nil || fullChat == nil {
			gologging.ErrorF(
				"Failed to get full channel for cplay ID %d: %v",
				cplayID, err,
			)
			m.Reply(
				F(chatID, "cplay_channel_not_accessible"),
				&telegram.SendOptions{ParseMode: "HTML"},
			)
			return telegram.EndGroup
		}

		if err := database.SetCPlayID(m.ChannelID(), cplayID); err != nil {
			gologging.ErrorF(
				"Failed to set cplay ID for chat %d: %v",
				m.ChannelID(), err,
			)
			m.Reply(
				F(chatID, "cplay_save_error"),
				&telegram.SendOptions{ParseMode: "HTML"},
			)
			return err
		}

		m.Reply(
			F(chatID, "cplay_enabled", locales.Arg{
				"channel_id": cplayID,
			}),
			&telegram.SendOptions{ParseMode: "HTML"},
		)
		return telegram.EndGroup
	}

	return handlePlay(m, false, true)
}

func cfplayHandler(m *telegram.NewMessage) error {
	return handlePlay(m, true, true)
}

func handlePlay(m *telegram.NewMessage, force, cplay bool) error {
	mention := utils.MentionHTML(m.Sender)

	// 1️⃣ Prepare room + search message
	r, replyMsg, err := prepareRoomAndSearchMessage(m, cplay)
	if err != nil {
		return telegram.EndGroup
	}

	// 2️⃣ Tracks and assistant status
	tracks, isActive, err := fetchTracksAndCheckStatus(m, replyMsg, r)
	if err != nil {
		return telegram.EndGroup
	}

	// 3️⃣ Filter & trim
	tracks, availableSlots, err := filterAndTrimTracks(replyMsg, r, tracks)
	if err != nil {
		return telegram.EndGroup
	}

	// 4️⃣ Download, play, respond
	if err := playTracksAndRespond(m, replyMsg, r, tracks, mention, isActive, force, availableSlots); err != nil {
		return err
	}

	return telegram.EndGroup
}

func prepareRoomAndSearchMessage(m *telegram.NewMessage, cplay bool) (*core.RoomState, *telegram.NewMessage, error) {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return nil, nil, err
	}

	r.SetCPlay(cplay)
	r.Parse()

	if len(r.Queue) >= config.QueueLimit {
		m.Reply(fmt.Sprintf(
			"⚠️ Queue limit reached (%d tracks max). Use /clear to clear queue.",
			config.QueueLimit,
		))
		return nil, nil, fmt.Errorf("queue limit reached")
	}

	parts := strings.SplitN(m.Text(), " ", 2)
	query := ""
	if len(parts) > 1 {
		query = strings.TrimSpace(parts[1])
	}

	if query == "" && !m.IsReply() {
		m.Reply(fmt.Sprintf(
			"🎵 <b>Whoops! No song detected.</b> Type <b>%s</b> <i>song name</i> or reply to a <i>media</i> to get the music going!",
			getCommand(m),
		))
		return nil, nil, fmt.Errorf("no song query")
	}

	searchStr := "🔍🎶 Searching... ⚡✨"
	if query != "" {
		searchStr = "🔍🎶 Searching for: " + html.EscapeString(query) + "... ⚡✨"
	}

	replyMsg, err := m.Reply(searchStr)
	if err != nil {
		gologging.ErrorF("Failed to send searching message: %v", err)
		return nil, nil, err
	}

	return r, replyMsg, nil
}

func fetchTracksAndCheckStatus(
	m *telegram.NewMessage,
	replyMsg *telegram.NewMessage,
	r *core.RoomState,
) ([]*state.Track, bool, error) {
	tracks, err := safeGetTracks(m, replyMsg)
	if err != nil {
		utils.EOR(replyMsg, err.Error())
		return nil, false, err
	}

	if len(tracks) == 0 {
		utils.EOR(replyMsg, "❌ No tracks found.")
		return nil, false, fmt.Errorf("no tracks found")
	}

	isActive := r.IsActiveChat()

	if _, err := core.GetVoiceChatStatus(r.ChatID); err != nil {
		gologging.ErrorF("Error getting voice chat status: %v", err)
		utils.EOR(replyMsg, getAssistantErrorMessage(err))
		return nil, false, err
	}

	if _, err := core.GetAssistantStatus(r.ChatID); err != nil {
		gologging.ErrorF("Error getting assistant status: %v", err)
		utils.EOR(replyMsg, getAssistantErrorMessage(err))
		return nil, false, err
	}

	return tracks, isActive, nil
}

func filterAndTrimTracks(
	replyMsg *telegram.NewMessage,
	r *core.RoomState,
	tracks []*state.Track,
) ([]*state.Track, int, error) {
	var filteredTracks []*state.Track
	var skippedTracks []string

	for _, track := range tracks {
		if track.Duration > config.DurationLimit {
			skippedTracks = append(
				skippedTracks,
				html.EscapeString(utils.ShortTitle(track.Title, 35)),
			)
			continue
		}
		filteredTracks = append(filteredTracks, track)
	}

	if len(skippedTracks) > 0 {
		// CASE 1: Only one track and it was skipped
		if len(tracks) == 1 && len(filteredTracks) == 0 {
			msg := fmt.Sprintf(
				"⚠️ You can't play songs longer than %d minutes.\n<i>%s</i> was skipped.",
				config.DurationLimit/60,
				skippedTracks[0],
			)
			utils.EOR(replyMsg, msg)
			return nil, 0, fmt.Errorf("single long track skipped")
		}

		// CASE 2: multiple skipped tracks
		var b strings.Builder
		fmt.Fprintf(&b,
			"<b>⚠️ %d tracks were skipped (max duration %d mins):</b>\n",
			len(skippedTracks),
			config.DurationLimit/60,
		)

		for i, title := range skippedTracks {
			if i < 5 {
				fmt.Fprintf(&b, "— <i>%s</i>\n", title)
			} else {
				fmt.Fprintf(&b, "... and %d more.\n", len(skippedTracks)-i)
				break
			}
		}

		utils.EOR(replyMsg, b.String())
		time.Sleep(1 * time.Second)
	}

	tracks = filteredTracks

	if len(tracks) == 0 {
		utils.EOR(replyMsg, "❌ All found tracks were skipped due to duration limits.")
		return nil, 0, fmt.Errorf("all tracks skipped")
	}

	availableSlots := config.QueueLimit - len(r.Queue)
	if availableSlots < len(tracks) {
		tracks = tracks[:availableSlots]
		gologging.WarnF("Queue full — adding only %d tracks out of request.", availableSlots)
	}

	return tracks, availableSlots, nil
}

func playTracksAndRespond(
	m *telegram.NewMessage,
	replyMsg *telegram.NewMessage,
	r *core.RoomState,
	tracks []*state.Track,
	mention string,
	isActive, force bool,
	availableSlots int,
) error {
	for i, track := range tracks {
		track.BY = mention
		title := html.EscapeString(utils.ShortTitle(track.Title, 25))
		var filePath string

		// Download first track if needed
		if i == 0 && (!isActive || force) {
			var opt *telegram.SendOptions
			if track.Duration > 420 {
				opt = &telegram.SendOptions{ReplyMarkup: core.GetCancekKeyboard()}
			}

			replyMsg, _ = utils.EOR(replyMsg, fmt.Sprintf("📥 Downloading song \"%s\"", title), opt)

			ctx, cancel := context.WithCancel(context.Background())
			downloadCancels[r.ChatID] = cancel
			defer func() {
				if _, ok := downloadCancels[r.ChatID]; ok {
					delete(downloadCancels, r.ChatID)
					cancel()
				}
			}()

			path, err := safeDownload(ctx, track, replyMsg)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					utils.EOR(replyMsg, fmt.Sprintf("⚠️ Download canceled by %s.", mention))
				} else {
					utils.EOR(replyMsg, fmt.Sprintf("❌ Failed to download \"%s\". Error: %v", title, err))
				}
				return telegram.EndGroup
			}

			filePath = path
			gologging.InfoF("Downloaded track to %s", filePath)
		}

		// 🔁 play with retry handled by helper
		if err := playTrackWithRetry(r, track, filePath, force && i == 0, replyMsg); err != nil {
			return err
		}

		sendPlayLogs(m, track, (isActive && !force) || i > 0)
	}

	mainTrack := tracks[0]

	// ---------- Now Playing / Added to queue ----------
	if !isActive || (force && len(tracks) > 0) {
		title := html.EscapeString(utils.ShortTitle(mainTrack.Title, 25))
		btn := core.GetPlayMarkup(r, false)

		var opt telegram.SendOptions
		opt.ParseMode = "HTML"
		opt.ReplyMarkup = btn
		if mainTrack.Artwork != "" {
			opt.Media = utils.CleanURL(mainTrack.Artwork)
		}

		replyMsg, _ = utils.EOR(
			replyMsg,
			fmt.Sprintf(
				"<b>🎵 Now Playing:</b>\n\n<b>▫ Track:</b> <a href=\"%s\">%s</a>\n<b>▫ Duration:</b> %s\n<b>▫ Requested by:</b> %s",
				mainTrack.URL,
				title,
				formatDuration(mainTrack.Duration),
				mention,
			),
			&opt,
		)

		r.SetMystic(replyMsg)

		if len(tracks) > 1 {
			var b strings.Builder
			b.WriteString(fmt.Sprintf("➕ <b>Added %d tracks</b> by %s\n\n", len(tracks)-1, mention))
			if availableSlots <= len(tracks) {
				b.WriteString("⚠️ <i>Queue limit reached — some tracks were skipped.</i>\n")
			}
			b.WriteString("<i>Use /queue to view full list.</i>")
			replyMsg.Respond(b.String())
		}
	} else {
		if len(tracks) == 1 {
			title := html.EscapeString(utils.ShortTitle(mainTrack.Title, 25))
			btn := core.GetPlayMarkup(r, true)
			opt := &telegram.SendOptions{
				ParseMode:   "HTML",
				ReplyMarkup: btn,
			}
			if mainTrack.Artwork != "" {
				opt.Media = utils.CleanURL(mainTrack.Artwork)
			}
			replyMsg, _ = utils.EOR(
				replyMsg,
				fmt.Sprintf(
					"<b>🎵 Added to Queue:</b>\n\n<b>▫ Track:</b> <a href=\"%s\">%s</a>\n<b>▫ Duration:</b> %s\n<b>▫ Requested by:</b> %s",
					mainTrack.URL,
					title,
					formatDuration(mainTrack.Duration),
					mention,
				),
				opt,
			)
		} else {
			var b strings.Builder
			b.WriteString(fmt.Sprintf("➕ <b>Added %d tracks</b> by %s\n\n", len(tracks), mention))
			if availableSlots <= len(tracks) {
				b.WriteString("⚠️ <i>Queue limit reached — some tracks were skipped.</i>\n")
			}
			b.WriteString("<i>Use /queue to view full list.</i>")
			utils.EOR(replyMsg, b.String())
		}
	}

	return nil
}

func playTrackWithRetry(
	r *core.RoomState,
	track *state.Track,
	filePath string,
	force bool,
	replyMsg *telegram.NewMessage,
) error {
	for attempt := 1; attempt <= playMaxRetries; attempt++ {
		err := r.Play(track, filePath, force)
		if err == nil {
			if attempt > 1 {
				gologging.Info("Successfully played after retry attempt " + utils.IntToStr(attempt))
			}
			return nil
		}

		// FloodWait
		if wait := telegram.GetFloodWait(err); wait > 0 {
			gologging.Error("FloodWait detected (" + strconv.Itoa(wait) + "s). Retrying... (attempt " + utils.IntToStr(attempt) + ")")
			time.Sleep(time.Duration(wait) * time.Second)
			continue
		}

		// RTMP unsupported
		if strings.Contains(err.Error(), "Streaming is not supported when using RTMP") {
			utils.EOR(replyMsg, "⚠️ RTMP stream not supported right now.")
			r.Destroy()
			return telegram.EndGroup
		}

		// INTERDC_X_CALL_ERROR → retry
		if tg.MatchError(err, "INTERDC_X_CALL_ERROR") {
			gologging.Error("INTERDC_X_CALL_ERROR occurred. Retrying... (attempt " + utils.IntToStr(attempt) + ")")
			time.Sleep(2 * time.Second)
			continue
		}

		// Last attempt failed
		if attempt == playMaxRetries {
			gologging.Error("❌ Failed to play after " + utils.IntToStr(maxRetries) + " attempts. Error: " + err.Error())
			utils.EOR(replyMsg, "❌ Failed to play\nError: "+err.Error())
			return err
		}

		gologging.Error("Unexpected error occurred. Retrying... (attempt " + utils.IntToStr(attempt) + "): " + err.Error())
	}

	return nil
}

type msgFn func(error) string

var errMessageMap = map[error]msgFn{
	core.ErrAssistantBanned: func(_ error) string {
		return "<b>🚫 Assistant Restricted</b>\n\nI can't play music because " +
			utils.MentionHTML(core.UbUser) +
			"(UserID: <code>" + utils.IntToStr(core.UbUser.ID) +
			"</code>) is banned or removed from this chat.\n\n" +
			"<i><b>✅ Unbanned already?</b> Use /reload to refresh and sync.</i>"
	},

	core.ErrAdminPermissionRequired: func(_ error) string {
		return "⚠️ <b>Admin Permission Required</b>\n\n" +
			"I need <i>admin access</i> to manage and check members in this chat.\n\n" +
			"➤ <b>Promote me with</b> <code>Manage Chat / Invite Users</code> permission."
	},

	core.ErrAssistantJoinRateLimited: func(_ error) string {
		return "⚠️ Assistant cannot join because it has reached the maximum number of allowed groups."
	},

	core.ErrAssistantJoinRequestSent: func(_ error) string {
		return "⚠️ Assistant sent a join request, but I couldn't auto-approve it.\n\n" +
			"<i>✅ Please manually approve the request, then try again.</i>"
	},

	core.ErrAssistantInviteLinkFetch: func(e error) string {
		return "⚠️ Failed to fetch invite link:\n\n<i>" + e.Error() + "</i>"
	},

	core.ErrAssistantInviteFailed: func(e error) string {
		return "⚠️ Assistant failed to join this chat:\n\n<i>" + e.Error() + "</i>"
	},

	core.ErrAssistantJoinRejected: func(_ error) string {
		return "⚠️ Invite link is invalid or expired. Please regenerate a fresh invite link."
	},

	core.ErrNoActiveVoiceChat: func(_ error) string {
		return "<b>🎙️ No Active Voice Chat</b>\n\n" +
			"I can't join yet — please start a voice chat to begin playing music.\n\n" +
			"<i><b>✅ Already started one?</b> Use /reload to refresh and sync this chat.</i>"
	},

	core.ErrFetchFailed: func(e error) string {
		return "⚠️ Failed to fetch chat info:\n\n<i>" + e.Error() + "</i>"
	},

	core.ErrPeerResolveFailed: func(_ error) string {
		return "⚠️ Failed to resolve peer information. Try again later or re-add the assistant."
	},
}

func getAssistantErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	for key, fn := range errMessageMap {
		if errors.Is(err, key) {
			return fn(err)
		}
	}

	return "⚠️ Unknown Error Occurred:\n\n<i>" + err.Error() + "</i>"
}

// Both safeDownload and safeGetTracks re-raise panic because all command
// handlers are wrapped by SafeMessageHandler, which catches panics and sends
// the debug trace to the logger and the owner.

func safeGetTracks(m, replyMsg *telegram.NewMessage) (tracks []*state.Track, err error) {
	defer func() {
		if r := recover(); r != nil {
			utils.EOR(replyMsg, "⚠️ Failed to fetch track details.\nPlease try again later.")
			panic(r)
		}
	}()

	tracks, err = platforms.GetTracks(m)
	return tracks, err
}

func safeDownload(
	ctx context.Context,
	track *state.Track,
	replyMsg *telegram.NewMessage,
) (path string, err error) {
	defer func() {
		if r := recover(); r != nil {
			utils.EOR(replyMsg, "⚠️ Download failed due to an unexpected internal error.")
			panic(r)
		}
	}()

	path, err = platforms.Download(ctx, track, replyMsg)
	return path, err
}
