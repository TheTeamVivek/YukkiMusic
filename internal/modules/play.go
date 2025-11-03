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

	"main/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/platforms"
	"main/internal/state"
	"main/internal/utils"
)

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

	if len(args) > 1 && args[1] == "--set" {
		if len(args) < 3 {
			m.Reply("<b>Usage:</b> <code>/cplay --set &lt;channel_id&gt;</code>", telegram.SendOptions{ParseMode: "HTML"})
			return telegram.EndGroup
		}

		cplayIDStr := args[2]
		cplayID, err := strconv.ParseInt(cplayIDStr, 10, 64)
		if err != nil {
			m.Reply("<b>Invalid Chat ID:</b> Please provide a valid integer ID for the channel.", telegram.SendOptions{ParseMode: "HTML"})
			return telegram.EndGroup
		}

		peer, err := m.Client.ResolvePeer(cplayID)
		if err != nil {
			m.Reply("<b>Failed to resolve peer:</b> Could not fetch channel details. Ensure I can access this channel.", telegram.SendOptions{ParseMode: "HTML"})
			return telegram.EndGroup
		}

		chPeer, ok := peer.(*telegram.InputPeerChannel)
		if !ok {
			m.Reply("<b>Invalid Target:</b> The provided ID is not a valid channel.", telegram.SendOptions{ParseMode: "HTML"})
			return telegram.EndGroup
		}

		fullChat, err := m.Client.ChannelsGetFullChannel(&telegram.InputChannelObj{
			ChannelID:  chPeer.ChannelID,
			AccessHash: chPeer.AccessHash,
		})
		if err != nil || fullChat == nil {
			gologging.GetLogger("CPlay").ErrorF("Failed to get full channel for cplay ID %d: %v", cplayID, err)
			m.Reply("<b>Channel Not Accessible:</b> Could not retrieve channel information. Please ensure I am an administrator in the target channel.", telegram.SendOptions{ParseMode: "HTML"})
			return telegram.EndGroup
		}

		if err := database.SetCPlayID(m.ChannelID(), cplayID); err != nil {
			gologging.GetLogger("CPlay").ErrorF("Failed to set cplay ID for chat %d: %v", m.ChannelID(), err)
			m.Reply("<b>Error:</b> Failed to save CPlay settings.", telegram.SendOptions{ParseMode: "HTML"})
			return err
		}

		m.Reply(fmt.Sprintf("✅ <b>Channel Play enabled.</b> All <code>/c</code> commands will now work in channel <code>%d</code>.", cplayID), telegram.SendOptions{ParseMode: "HTML"})
		return telegram.EndGroup
	}

	return handlePlay(m, false, true)
}

func cfplayHandler(m *telegram.NewMessage) error {
	return handlePlay(m, true, true)
}

func handlePlay(m *telegram.NewMessage, force, cplay bool) error {
	logger := gologging.GetLogger("Play")
	mention := utils.MentionHTML(m.Sender)
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.EndGroup
	}
	r.SetCPlay(cplay)
	r.Parse()
	if len(r.Queue) >= config.QueueLimit {
		m.Reply(fmt.Sprintf("⚠️ Queue limit reached (%d tracks max). Use /clear to clear queue.", config.QueueLimit))
		return telegram.EndGroup
	}
	parts := strings.SplitN(m.Text(), " ", 2)
	query := ""
	if len(parts) > 1 {
		query = strings.TrimSpace(parts[1])
	}
	if query == "" && !m.IsReply() {
		m.Reply(fmt.Sprintf("🎵 <b>Whoops! No song detected.</b> Type <b>%s</b> <i>song name</i> or reply to a <i>media</i> to get the music going!", getCommand(m)))
		return telegram.EndGroup
	}
	searchStr := "🔍🎶 Searching... ⚡✨"
	if query != "" {
		searchStr = "🔍🎶 Searching for: " + html.EscapeString(query) + "... ⚡✨"
	}
	replyMsg, err := m.Reply(searchStr)
	if err != nil {
		logger.ErrorF("Failed to send searching message: %v", err)
		return telegram.EndGroup
	}
	tracks, err := safeGetTracks(m, replyMsg)
	if err != nil {
		utils.EOR(replyMsg, err.Error())
		return telegram.EndGroup
	}
	if len(tracks) == 0 {
		utils.EOR(replyMsg, "❌ No tracks found.")
		return telegram.EndGroup
	}
	isActive := r.IsActiveChat()
	if _, err := core.GetVoiceChatStatus(r.ChatID); err != nil {
		logger.ErrorF("Error getting voice chat status: %v", err)
		utils.EOR(replyMsg, getAssistantErrorMessage(err))
		return telegram.EndGroup
	}
	if _, err := core.GetAssistantStatus(r.ChatID); err != nil {
		logger.ErrorF("Error getting assistant status: %v", err)
		utils.EOR(replyMsg, getAssistantErrorMessage(err))
		return telegram.EndGroup
	}
	var filteredTracks []*state.Track
	var skippedTracks []string
	for _, track := range tracks {
		if track.Duration > config.DurationLimit {
			skippedTracks = append(skippedTracks, html.EscapeString(utils.ShortTitle(track.Title, 35)))
			continue
		}
		filteredTracks = append(filteredTracks, track)
	}
	if len(skippedTracks) > 0 {
		var b strings.Builder
		b.WriteString(fmt.Sprintf("<b>⚠️ %d tracks were skipped (max duration %d mins):</b>\n", len(skippedTracks), config.DurationLimit/60))
		for i, title := range skippedTracks {
			if i < 5 { // Limit to showing 5 tracks to avoid flood
				b.WriteString(fmt.Sprintf("— <i>%s</i>\n", title))
			} else {
				b.WriteString(fmt.Sprintf("... and %d more.\n", len(skippedTracks)-i))
				break
			}
		}
		utils.EOR(replyMsg, b.String())
		if len(tracks) == 1 && len(filteredTracks) == 0 {
			return telegram.EndGroup
		}
		<-time.After(1 * time.Second)
	}
	tracks = filteredTracks
	if len(tracks) == 0 {
		utils.EOR(replyMsg, "❌ All found tracks were skipped due to duration limits.")
		return telegram.EndGroup
	}
	availableSlots := config.QueueLimit - len(r.Queue)
	if availableSlots < len(tracks) {
		tracks = tracks[:availableSlots]
		logger.WarnF("Queue full — adding only %d tracks out of request.", availableSlots)
	}
	for i, track := range tracks {
		track.BY = mention
		title := html.EscapeString(utils.ShortTitle(track.Title, 25))
		var filePath string
		if i == 0 && (!isActive || force) {
			replyMsg, _ = utils.EOR(replyMsg, fmt.Sprintf("📥 Downloading song \"%s\"", title))
			path, err := platforms.Download(context.Background(), track, replyMsg)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					utils.EOR(replyMsg, fmt.Sprintf("⚠️ Download canceled by %s.", mention))
				} else {
					utils.EOR(replyMsg, fmt.Sprintf("❌ Failed to download \"%s\". Error: %v", title, err))
				}
				return telegram.EndGroup
			}
			filePath = path
			logger.InfoF("Downloaded track to %s", filePath)
		}
		if err := r.Play(track, filePath, force && i == 0); err != nil {
			utils.EOR(replyMsg, "❌ Failed to play\nError: "+err.Error())
			return err
		}
		sendPlayLogs(m, track, (isActive && !force) || i > 0)
	}
	mainTrack := tracks[0]
	if !isActive || (force && len(tracks) > 0) {
		title := html.EscapeString(utils.ShortTitle(mainTrack.Title, 25))
		btn := core.GetPlayMarkup(r, false)
		var opt telegram.SendOptions
		opt.ParseMode = "HTML"
		opt.ReplyMarkup = btn
		if mainTrack.Artwork != "" {
			opt.Media = utils.CleanURL(mainTrack.Artwork)
		}
		replyMsg, _ = utils.EOR(replyMsg, fmt.Sprintf("<b>🎵 Now Playing:</b>\n\n<b>▫ Track:</b> <a href=\"%s\">%s</a>\n<b>▫ Duration:</b> %s\n<b>▫ Requested by:</b> %s", mainTrack.URL, title, formatDuration(mainTrack.Duration), mention), opt)
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
			var opt telegram.SendOptions
			opt.ParseMode = "HTML"
			opt.ReplyMarkup = btn
			if mainTrack.Artwork != "" {
				opt.Media = utils.CleanURL(mainTrack.Artwork)
			}
			replyMsg, _ = utils.EOR(replyMsg, fmt.Sprintf("<b>🎵 Added to Queue:</b>\n\n<b>▫ Track:</b> <a href=\"%s\">%s</a>\n<b>▫ Duration:</b> %s\n<b>▫ Requested by:</b> %s", mainTrack.URL, title, formatDuration(mainTrack.Duration), mention), opt)
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
	return telegram.EndGroup
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
