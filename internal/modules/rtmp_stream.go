package modules

import (
	"context"
	"errors"
	"fmt"
	"html"
	"strconv"
	"strings"
	"sync"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

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
	helpTexts["stream"] = `<i>Start RTMP live streaming to configured server.</i>

<u>Usage:</u>
<b>/stream &lt;query/URL&gt;</b> — Start streaming a track
<b>/stream [reply to audio/video]</b> — Stream replied media

<b>🎥 Features:</b>
• Live streaming to your RTMP server
• Supports audio and video
• Queue support (like /play)
• Real-time status monitoring

<b>⚙️ Setup Required:</b>
Before using this command, an admin must configure RTMP:
1. Open bot's private chat (DM)
2. Send: <code>/setrtmp &lt;chat_id&gt; &lt;rtmp_url&gt;</code>

<b>📝 Example Setup:</b>
In bot DM:
<code>/setrtmp -1001234567890 rtmps://dc5-1.rtmp.t.me/s/123:key</code>

Then in your chat:
<code>/stream never gonna give you up</code>

<b>⚠️ Important Notes:</b>
• RTMP streams have ~15-30s buffering delay
• Setup ONLY works in bot DM (for security)
• Use <code>/streamstop</code> to end stream
• Only admin/auth users can control streams
• We do NOT use Telegram's RTMP API - you provide your own server`

	helpTexts["streamstop"] = `<i>Stop current RTMP stream.</i>

<u>Usage:</u>
<b>/streamstop</b> — Stop the active stream

<b>⚠️ Note:</b>
Only admin/auth users can stop streams.`

	helpTexts["streamstatus"] = `<i>Check current RTMP stream status.</i>

<u>Usage:</u>
<b>/streamstatus</b> — Show stream information

<b>📊 Shows:</b>
• Stream state (playing/stopped)
• Current position
• RTMP server (masked for security)
• Configuration status`

	helpTexts["setrtmp"] = `<i>Configure RTMP streaming server (DM only).</i>

<u>Usage:</u>
<b>/setrtmp &lt;chat_id&gt; &lt;rtmp_url&gt;</b> — Set RTMP for a chat

<b>🔒 Security:</b>
• <b>This command ONLY works in DM</b> (private chat with bot)
• NEVER share RTMP credentials in groups
• Credentials are stored securely in database
• We do NOT use Telegram's RTMP API

<b>📋 URL Format:</b>
<code>rtmp://server/app/streamkey</code>
or
<code>rtmps://server/s/streamkey</code>

<b>📝 Examples:</b>

<b>Telegram Voice Chat:</b>
1. Start voice chat in your channel
2. Telegram gives you: <code>rtmps://dc5-1.rtmp.t.me/s/123:key</code>
3. In bot DM send: <code>/setrtmp -1001234567890 rtmps://dc5-1.rtmp.t.me/s/123:key</code>

<b>Custom RTMP Server:</b>
<code>/setrtmp -1001234567890 rtmp://live.example.com/stream/mykey</code>

<b>🔍 Getting Chat ID:</b>
• Forward message from chat to @userinfobot
• Or use <code>/id</code> command in the chat

<b>⚠️ Requirements:</b>
• You must be admin in target chat
• Bot must be member of target chat
• Command only works in bot's private chat

<b>💡 Why DM only?</b>
RTMP stream keys are like passwords. Configuring in DM prevents accidental exposure in group chats.`
}

// Get or create RTMP stream for chat
func getOrCreateRTMPStream(chatID int64) (*tg.RTMPStream, error) {
	rtmpStreamsMu.Lock()
	defer rtmpStreamsMu.Unlock()

	if stream, exists := rtmpStreams[chatID]; exists {
		return stream, nil
	}

	url, key, err := database.GetRTMP(chatID)
	if err != nil || url == "" || key == "" {
		return nil, fmt.Errorf("RTMP not configured. Admin must use /setrtmp in bot DM first")
	}

	stream, err := core.Bot.NewRTMPStream(chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to create RTMP stream: %w", err)
	}

	stream.SetLoopCount(0)
	stream.SetURL(url)
	stream.SetKey(key)

	stream.OnError(func(err error) {
		gologging.ErrorF("RTMP error in chat %d: %v", chatID, err)
		core.Bot.SendMessage(chatID, "⚠️ RTMP stream encountered an error. Check logs for details.")
	})

	rtmpStreams[chatID] = stream
	return stream, nil
}

func streamHandler(m *tg.NewMessage) error {
	return handleStream(m, false)
}

func handleStream(m *tg.NewMessage, force bool) error {
	chatID := m.ChannelID()

	url, key, err := database.GetRTMP(chatID)
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

	stream, err := getOrCreateRTMPStream(chatID)
	if err != nil {
		m.Reply(F(chatID, "rtmp_init_failed", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	if stream.State() == tg.StreamStatePlaying && !force {
		m.Reply(F(chatID, "rtmp_already_streaming"))
		return tg.ErrEndGroup
	}

	searchStr := ""
	if query != "" {
		searchStr = F(chatID, "searching_query", locales.Arg{
			"query": html.EscapeString(query),
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
		"title": html.EscapeString(utils.ShortTitle(track.Title, 25)),
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
				"title": html.EscapeString(utils.ShortTitle(track.Title, 25)),
				"error": html.EscapeString(err.Error()),
			}))
		}
		return tg.ErrEndGroup
	}

	// Start streaming
	utils.EOR(replyMsg, F(chatID, "rtmp_starting_stream"))

	if err := stream.Play(filePath); err != nil {
		utils.EOR(replyMsg, F(chatID, "rtmp_play_failed", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	// Success message
	title := html.EscapeString(utils.ShortTitle(track.Title, 25))
	msgText := F(chatID, "rtmp_now_streaming", locales.Arg{
		"url":      track.URL,
		"title":    title,
		"duration": formatDuration(track.Duration),
		"by":       mention,
	})

	opt := &tg.SendOptions{
		ParseMode: "HTML",
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
		m.Reply(F(chatID, "rtmp_not_streaming"))
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

// /streamstatus - Check RTMP status
func streamStatusHandler(m *tg.NewMessage) error {
	chatID := m.ChannelID()

	// Check if RTMP is configured (without exposing credentials)
	url, _, err := database.GetRTMP(chatID)
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
		m.Reply(F(chatID, "rtmp_configured_not_started", locales.Arg{
			"server": maskRTMPURL(url),
		}))
		return tg.ErrEndGroup
	}

	state := stream.State()
	pos := stream.CurrentPosition()

	var statusText string
	switch state {
	case tg.StreamStatePlaying:
		statusText = F(chatID, "rtmp_status_playing", locales.Arg{
			"position": formatDuration(int(pos.Seconds())),
			"server":   maskRTMPURL(url),
		})
	case tg.StreamStatePaused:
		statusText = F(chatID, "rtmp_status_paused", locales.Arg{
			"server": maskRTMPURL(url),
		})
	default:
		statusText = F(chatID, "rtmp_status_idle", locales.Arg{
			"server": maskRTMPURL(url),
		})
	}

	m.Reply(statusText)
	return tg.ErrEndGroup
}

// /setrtmp - Configure RTMP (DM only for security)
func setRTMPHandler(m *tg.NewMessage) error {
	if !filterChannel(m) {
		return tg.ErrEndGroup
	}

	m.Delete()

	switch m.ChatType() {
	case tg.EntityChat:
		m.Reply(F(m.ChannelID(), "rtmp_dm_only", locales.Arg{
			"cmd": "/setrtmp",
		}))
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

	isAdmin, err := utils.IsChatAdmin(m.Client, targetChatID, m.SenderID())
	if err != nil {
		m.Reply(F(m.ChannelID(), "rtmp_check_admin_failed", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}
	if !isAdmin {
		m.Reply(F(m.ChannelID(), "rtmp_not_admin"))
		return tg.ErrEndGroup
	}

	if err := database.SetRTMP(targetChatID, url, key); err != nil {
		m.Reply(F(m.ChannelID(), "rtmp_init_failed", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	rtmpStreamsMu.Lock()
	if stream, exists := rtmpStreams[targetChatID]; exists {
		stream.SetURL(url)
		stream.SetKey(key)
	}
	rtmpStreamsMu.Unlock()

	m.Reply(F(m.ChannelID(), "rtmp_configured_success", locales.Arg{
		"chat_id": targetChatID,
		"url":     url,
		"key":     maskKey(key),
	}))

	return tg.ErrEndGroup
}

func maskRTMPURL(url string) string {
	if idx := strings.Index(url, "://"); idx != -1 {
		proto := url[:idx+3]
		rest := url[idx+3:]
		if len(rest) > 10 {
			return proto + rest[:10] + "***"
		}
	}
	return url
}

func maskKey(k string) string {
	l := len(k)
	if l <= 4 {
		return "****"
	}
	if l <= 8 {
		return k[:2] + "****" + k[l-2:]
	}
	return k[:4] + "****" + k[l-4:]
}
