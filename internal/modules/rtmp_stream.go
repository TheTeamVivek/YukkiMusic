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
	"strconv"
	"strings"
	"sync"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

var (
	rtmpStreams   = make(map[int64]*tg.RTMPStream)
	rtmpStreamsMu sync.RWMutex
)

func init() {
	helpTexts["stream"] = `<i>Start RTMP stream in this chat.</i>

<u>Usage:</u>
<b>/stream &lt;query/URL&gt;</b>
<b>/stream [reply to audio/video]</b>`

	helpTexts["streamstop"] = `<i>Stop the current RTMP stream.</i>

<u>Usage:</u>
<b>/streamstop</b>`

	helpTexts["streamstatus"] = `<i>Check RTMP stream status.</i>

<u>Usage:</u>
<b>/streamstatus</b>`

	helpTexts["setrtmp"] = `<i>Set RTMP URL in bot DM.</i>

<u>Usage:</u>
<b>/setrtmp &lt;chat_id&gt; &lt;rtmp_url&gt;</b>`
}

// Get or create RTMP stream for chat
func getOrCreateRTMPStream(chatID int64, url, key string) *tg.RTMPStream {
	rtmpStreamsMu.Lock()
	defer rtmpStreamsMu.Unlock()

	if stream, exists := rtmpStreams[chatID]; exists {
		return stream
	}

	stream, err := core.Bot.NewRTMPStream(chatID)
	if err != nil {
		return nil
	}

	stream.SetLoopCount(0)
	stream.SetURL(url)
	stream.SetKey(key)

	stream.OnError(func(chatID int64, err error) {
		gologging.ErrorF("RTMP error in chat %d: %v", chatID, err)
		core.Bot.SendMessage(
			chatID,
			"⚠️ RTMP stream encountered an error. Check logs for details.",
		)
	})

	rtmpStreams[chatID] = stream
	return stream
}

func streamHandler(m *tg.NewMessage) error {
	return handleStream(m, false)
}

func handleStream(m *tg.NewMessage, force bool) error {
	chatID := m.ChannelID()

	url, key, err := database.RTMP(chatID)
	if err != nil || url == "" || key == "" {
		m.Reply(F(chatID, "rtmp_not_configured", locales.Arg{
			"cmd": "/setrtmp",
		}))
		return tg.ErrEndGroup
	}

	parts := strings.SplitN(m.Text(), " ", 2)
	query := ""
	if len(parts) > 1 {
		query = strings.TrimSpace(parts[1])
	}

	if query == "" && !m.IsReply() {
		m.Reply(F(chatID, "no_song_query", locales.Arg{
			"cmd": getCommand(m),
		}))
		return tg.ErrEndGroup
	}

	stream := getOrCreateRTMPStream(chatID, url, key)
	if stream == nil {
		m.Reply("failed to create rtmp stream")
		return tg.ErrEndGroup
	}

	if stream.State() == tg.StreamStatePlaying && !force {
		m.Reply(F(chatID, "rtmp_already_streaming"))
		return tg.ErrEndGroup
	}

	searchStr := ""
	if query != "" {
		searchStr = F(chatID, "searching_query", locales.Arg{
			"query": utils.EscapeHTML(query),
		})
	} else {
		searchStr = F(chatID, "searching")
	}

	replyMsg, err := m.Reply(searchStr)
	if err != nil {
		gologging.ErrorF("Failed to send searching message: %v", err)
		return tg.ErrEndGroup
	}

	tracks, err := safeGetTracks(m, replyMsg, chatID, false)
	if err != nil {
		utils.EOR(replyMsg, err.Error())
		return tg.ErrEndGroup
	}

	if len(tracks) == 0 {
		utils.EOR(replyMsg, F(chatID, "no_song_found"))
		return tg.ErrEndGroup
	}

	track := tracks[0]
	mention := utils.MentionHTML(m.Sender)
	track.Requester = mention

	// Download track
	downloadingText := F(chatID, "play_downloading_song", locales.Arg{
		"title": utils.EscapeHTML(utils.ShortTitle(track.Title, 25)),
	})
	replyMsg, _ = utils.EOR(replyMsg, downloadingText)

	ctx, cancel := context.WithCancel(context.Background())
	downloadCancels[chatID] = cancel
	defer func() {
		if _, ok := downloadCancels[chatID]; ok {
			delete(downloadCancels, chatID)
			cancel()
		}
	}()

	filePath, err := safeDownload(ctx, track, replyMsg, chatID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.EOR(replyMsg, F(chatID, "play_download_canceled", locales.Arg{
				"user": mention,
			}))
		} else {
			utils.EOR(replyMsg, F(chatID, "play_download_failed", locales.Arg{
				"title": utils.EscapeHTML(utils.ShortTitle(track.Title, 25)),
				"error": utils.EscapeHTML(err.Error()),
			}))
		}
		return tg.ErrEndGroup
	}

	// Start streaming

	if err := stream.Play(filePath); err != nil {
		utils.EOR(replyMsg, F(chatID, "rtmp_play_failed", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	// Success message
	btn := tg.NewKeyboard()
	stopBtn := tg.Button.Data(F(chatID, "CONFIRM_STOP_BTN"), "rtmp_stop")
	if !config.DisableColour {
			stopBtn.Danger()
	}
	btn.AddRow(stopBtn)

	title := utils.EscapeHTML(utils.ShortTitle(track.Title, 25))
	msgText := F(chatID, "rtmp_now_streaming", locales.Arg{
		"url":      track.URL,
		"title":    title,
		"duration": utils.FormatDuration(track.Duration),
		"by":       mention,
	})

	opt := &tg.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: btn.Build(),
	}

	if track.Artwork != "" {
		opt.Media = utils.CleanURL(track.Artwork)
	}

	utils.EOR(replyMsg, msgText, opt)
	return tg.ErrEndGroup
}

// /streamstop - Stop RTMP stream
func streamStopHandler(m *tg.NewMessage) error {
	chatID := m.ChannelID()

	rtmpStreamsMu.RLock()
	stream, exists := rtmpStreams[chatID]
	rtmpStreamsMu.RUnlock()

	if !exists || stream.State() != tg.StreamStatePlaying {
		m.Reply(F(chatID, "room_no_active"))
		return tg.ErrEndGroup
	}

	if err := stream.Stop(); err != nil {
		m.Reply(F(chatID, "rtmp_stop_failed", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	m.Reply(F(chatID, "rtmp_stopped", locales.Arg{
		"user": utils.MentionHTML(m.Sender),
	}))

	return tg.ErrEndGroup
}

func rtmpStopCallbackHandler(cb *tg.CallbackQuery) error {
	chatID := cb.ChannelID()
	opt := &tg.CallbackOptions{Alert: true}

	if !checkAdminOrAuth(cb, chatID) {
		return tg.ErrEndGroup
	}

	rtmpStreamsMu.RLock()
	stream, exists := rtmpStreams[chatID]
	rtmpStreamsMu.RUnlock()

	if !exists || stream.State() != tg.StreamStatePlaying {
		cb.Answer(F(chatID, "room_no_active"), opt)
		return tg.ErrEndGroup
	}

	if err := stream.Stop(); err != nil {
		cb.Answer(F(chatID, "rtmp_stop_failed", locales.Arg{
			"error": err.Error(),
		}), opt)
		return tg.ErrEndGroup
	}

	_, _ = cb.Edit(F(chatID, "rtmp_stopped", locales.Arg{
		"user": utils.MentionHTML(cb.Sender),
	}))
	cb.Answer(F(chatID, "cb_stop_success"), &tg.CallbackOptions{})
	return tg.ErrEndGroup
}

// /streamstatus - Check RTMP status
func streamStatusHandler(m *tg.NewMessage) error {
	chatID := m.ChannelID()

	// Check if RTMP is configured (without exposing credentials)
	url, _, err := database.RTMP(chatID)
	if err != nil || url == "" {
		m.Reply(F(chatID, "rtmp_not_configured", locales.Arg{
			"cmd": "/setrtmp",
		}))
		return tg.ErrEndGroup
	}

	rtmpStreamsMu.RLock()
	stream, exists := rtmpStreams[chatID]
	rtmpStreamsMu.RUnlock()

	if !exists {
		// RTMP configured but not initialized yet
		m.Reply(F(chatID, "room_no_active"))
		return tg.ErrEndGroup
	}

	state := stream.State()
	pos := stream.CurrentPosition()

	var statusText string
	switch state {
	case tg.StreamStatePlaying:
		statusText = F(chatID, "rtmp_status_playing", locales.Arg{
			"position": utils.FormatDuration(int(pos.Seconds())),
		})
	default:
		statusText = F(chatID, "room_no_active")
	}

	m.Reply(statusText)
	return tg.ErrEndGroup
}

// /setrtmp - Configure RTMP (DM only for security)
func setRTMPHandler(m *tg.NewMessage) error {
	if !filterChannel(m) {
		return tg.ErrEndGroup
	}

	switch m.ChatType() {
	case tg.EntityChat:
		m.Reply(F(m.ChannelID(), "rtmp_dm_only"))
		return tg.ErrEndGroup
	case tg.EntityUser:
	default:
		return tg.ErrEndGroup
	}

	args := strings.Fields(m.Text())

	if len(args) < 3 {
		m.Reply(F(m.ChannelID(), "rtmp_setup_usage"))
		return tg.ErrEndGroup
	}

	cid := args[1]
	raw := args[2]

	idx := strings.LastIndex(raw, "/")
	if idx <= 0 || idx == len(raw)-1 {
		m.Reply(F(m.ChannelID(), "rtmp_parse_failed", locales.Arg{
			"error": "invalid RTMP format",
		}))
		return tg.ErrEndGroup
	}

	url := raw[:idx+1]
	key := raw[idx+1:]

	if url == "" || key == "" {
		m.Reply(F(m.ChannelID(), "rtmp_parse_failed", locales.Arg{
			"error": "empty url or key",
		}))
		return tg.ErrEndGroup
	}

	targetChatID, err := strconv.ParseInt(cid, 10, 64)
	if err != nil {
		m.Reply(F(m.ChannelID(), "rtmp_invalid_chat_id"))
		return tg.ErrEndGroup
	}

	if err := database.SetRTMP(targetChatID, url, key); err != nil {
		m.Reply(F(m.ChannelID(), "generic_error", locales.Arg{"error": err.Error()}))
		return tg.ErrEndGroup
	}

	rtmpStreamsMu.Lock()
	if stream, exists := rtmpStreams[targetChatID]; exists {
		stream.SetURL(url)
		stream.SetKey(key)
	}
	rtmpStreamsMu.Unlock()

	m.Reply(F(m.ChannelID(), "rtmp_configured_success", locales.Arg{"chat_id": targetChatID}))

	return tg.ErrEndGroup
}

func clearRTMPState(chatID int64) {
	rtmpStreamsMu.Lock()
	defer rtmpStreamsMu.Unlock()

	if stream, ok := rtmpStreams[chatID]; ok {
		_ = stream.Stop()
		delete(rtmpStreams, chatID)
	}
}
