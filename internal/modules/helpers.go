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
	"errors"
	"fmt"
	"html"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/state"
	"main/internal/utils"
)

var (
	superGroupFilter    = tg.FilterFunc(FilterSuperGroup)
	adminFilter         = tg.FilterFunc(FilterChatAdmins)
	authFilter          = tg.FilterFunc(FilterAuthUsers)
	ignoreChannelFilter = tg.FilterFunc(FilterChannel)
	sudoOnlyFilter      = tg.FilterFunc(FilterSudo)
	ownerFilter         = tg.FilterFunc(FilterOwner)
)

func bool_(b bool) *bool {
	return &b
}

func eoe(e error) error {
	if e != nil {
		return e
	}
	return tg.EndGroup
}

func getEffectiveRoom(m *tg.NewMessage, cplay bool) (*core.RoomState, error) {
	chatID := m.ChannelID()
	if !cplay {
		r, _ := core.GetRoom(chatID, true)
		return r, nil
	}
	cplayID, err := database.GetCPlayID(chatID)
	if err != nil || cplayID == 0 {
		return nil, errors.New("cplay ID not set. Use /cplay --set <chat_id> to set it")
	}
	r, _ := core.GetRoom(cplayID, true)
	return r, nil
}

func getCbChatID(cb *tg.CallbackQuery) (int64, error) {
	// If private chat, just return sender ID directly
	if cb.IsPrivate() {
		return cb.SenderID, nil
	}

	// Otherwise, fetch the chat/channel info
	chat, err := cb.GetChannel()
	if err != nil {
		return 0, fmt.Errorf("get channel: %w", err)
	}

	chatID, err := utils.GetPeerID(cb.Client, chat.ID)
	if err != nil {
		return 0, fmt.Errorf("get peer ID: %w", err)
	}

	return chatID, nil
}

func sendPlayLogs(m *tg.NewMessage, track *state.Track, queued bool) {
	if config.LoggerID == 0 || config.LoggerID == m.ChatID() || config.LoggerID == m.ChannelID() {
		return
	}

	if is, err := database.IsLoggerEnabled(); err != nil {
		gologging.Error("Failed to get IsLoggerEnabled: " + err.Error())
	} else if !is {
		return
	}

	var (
		sb  strings.Builder
		err error
	)

	header := "Playback Started"
	if queued {
		header = "Playback Queued"
	}

	sb.WriteString("🎵 ")
	if m.Channel.Username != "" {
		sb.WriteString("<b><a href=\"")
		sb.WriteString(m.Link())
		sb.WriteString("\">")
		sb.WriteString(header)
		sb.WriteString("</a></b>\n\n")
	} else {
		sb.WriteString("<b><u>")
		sb.WriteString(header)
		sb.WriteString("</u></b>\n\n")
	}

	if track.Artwork != "" {
		sb.WriteString("<blockquote>")
	}

	sb.WriteString("<b>🎧 Track:</b> <a href=\"")
	sb.WriteString(track.URL)
	sb.WriteString("\">")
	sb.WriteString(utils.ShortTitle(track.Title))
	sb.WriteString("</a>\n")

	sb.WriteString("<b>🔗 Source:</b> ")
	sb.WriteString(string(track.Source))
	sb.WriteByte('\n')

	sb.WriteString("<b>📌 Group:</b> ")
	if m.Channel.Username != "" {
		sb.WriteByte('@')
		sb.WriteString(m.Channel.Username)
	} else {
		sb.WriteString(m.Channel.Title)
	}
	sb.WriteString(" (")
	sb.WriteString(strconv.FormatInt(m.ChannelID(), 10))
	sb.WriteString(")\n")

	sb.WriteString("<b>👤 Requested by:</b> ")
	if m.Sender.Username != "" {
		sb.WriteByte('@')
		sb.WriteString(m.Sender.Username)
	} else {
		sb.WriteString(utils.MentionHTML(m.Sender))
	}
	sb.WriteString(" (<code>")
	sb.WriteString(strconv.FormatInt(m.Sender.ID, 10))
	sb.WriteString("</code>)\n")

	sb.WriteString("<b>⏳ Timestamp:</b> ")
	sb.WriteString(time.Now().Format("2006-01-02 15:04:05"))

	if track.Artwork != "" {

		sb.WriteString("\n</blockquote>")
		_, err = core.Bot.SendMedia(config.LoggerID, utils.CleanURL(track.Artwork), &tg.MediaOptions{Caption: sb.String()})
	} else {
		_, err = core.Bot.SendMessage(config.LoggerID, sb.String())
	}
	if err != nil {
		gologging.Error("Failed to send logger msg: " + err.Error())
	}
}

func F(chatID int64, key string, values ...arg) string {
	lang, err := database.GetChatLanguage(chatID)
	if err != nil {
		gologging.Error("Failed to get language for " + utils.IntToStr(chatID) + " Got error " + err.Error())
		lang = config.DefaultLang
	}
	return FWithLang(lang, key, values...)
}

func FWithLang(lang, key string, values ...arg) string {
	var val arg
	if len(values) > 0 {
		val = values[0]
	}
	return locales.Get(lang, key, val)
}

func FilterOwner(m *tg.NewMessage) bool {
	if config.OwnerID == 0 || m.SenderID() != config.OwnerID {
		if m.IsPrivate() || strings.HasSuffix(m.GetCommand(), core.BUser.Username) {
			m.Reply(F(m.ChannelID(), "only_owner"))
		}
		return false
	}
	return true
}

func FilterSudo(m *tg.NewMessage) bool {
	is, _ := database.IsSudo(m.SenderID())

	if config.OwnerID == 0 || (m.SenderID() != config.OwnerID && !is) {
		if m.IsPrivate() || strings.HasSuffix(m.GetCommand(), core.BUser.Username) {
			m.Reply(F(m.ChannelID(), "only_sudo"))
		}
		return false
	}

	return true
}

func FilterChannel(m *tg.NewMessage) bool {
	if _, ok := m.Message.FromID.(*tg.PeerChannel); ok {
		return false
	}
	return true
}

func FilterAuthUsers(m *tg.NewMessage) bool {
	isAdmin, err := utils.IsChatAdmin(m.Client, m.ChannelID(), m.SenderID())
	if err == nil && isAdmin {
		return true
	}

	isAuth, err := database.IsAuthUser(m.ChannelID(), m.SenderID())
	if err == nil && isAuth {
		return true
	}

	m.Reply(F(m.ChannelID(), "only_admin_or_auth"))
	return false
}

func FilterSuperGroup(m *tg.NewMessage) bool {
	/*if m.Message.FromID == nil || (m.SenderChat != nil && m.SenderChat.ID != 0) {
		m.Reply("⚠️ You are using Anonymous Admin Mode.\n\n👉 Switch back to your user account to use commands.")
		return false
	}*/

	if !FilterChannel(m) {
		return false
	}
	// Validate chat type
	switch m.ChatType() {
	case tg.EntityChat:
		// EntityChat can be basic group or supergroup — allow only supergroup
		if m.Channel != nil && !m.Channel.Broadcast {
			database.AddServed(m.ChannelID())
			return true // Supergroup
		}
		warnAndLeave(m.Client, m.ChatID()) // Basic group → leave
		database.DeleteServed(m.ChannelID())
		return false

	case tg.EntityChannel:
		return false // Pure channel chat → ignore

	case tg.EntityUser:
		m.Reply(F(m.ChannelID(), "only_supergroup"))
		database.AddServed(m.ChannelID(), true)
		return false // Private chat → warn
	}

	return false
}

func FilterChatAdmins(m *tg.NewMessage) bool {
	isAdmin, err := utils.IsChatAdmin(m.Client, m.ChannelID(), m.SenderID())
	if err != nil || !isAdmin {
		m.Reply(F(m.ChannelID(), "only_admin"))
		return false
	}
	return true
}

func SafeCallbackHandler(handler func(*tg.CallbackQuery) error) func(*tg.CallbackQuery) error {
	return func(cb *tg.CallbackQuery) (err error) {
		if is, _ := database.IsMaintenance(); is {
			if cb.Sender.ID != config.OwnerID {
				if ok, _ := database.IsSudo(cb.Sender.ID); !ok {
					chatID, err := getCbChatID(cb)
					if err != nil {
						cb.Answer(FWithLang(config.DefaultLang, "chat_not_recognized"), &tg.CallbackOptions{Alert: true})
						return tg.EndGroup
					}
					cb.Answer(F(chatID, "maint", locales.Arg{"reason": ""}), &tg.CallbackOptions{Alert: true})
					return tg.EndGroup
				}
			}
		}
		defer func() {
			if r := recover(); r != nil {
				handlePanic(r, cb, true)
				err = fmt.Errorf("Some panics handled")
			}
		}()
		err = handler(cb)
		if err != nil {
			if errors.Is(err, tg.EndGroup) {
				return err
			}
			handlePanic(err, cb, false)
		}
		return err
	}
}

func SafeMessageHandler(handler func(*tg.NewMessage) error) func(*tg.NewMessage) error {
	return func(m *tg.NewMessage) (err error) {
		gologging.Info("Handling message from " + fmt.Sprint(m.SenderID()) + " in chat " + fmt.Sprint(m.ChannelID()))

		if is, _ := database.IsMaintenance(); is {
			gologging.Debug("Maintenance mode active")
			if m.SenderID() != config.OwnerID {
				if ok, _ := database.IsSudo(m.SenderID()); !ok {
					if m.ChatType() == tg.EntityUser || strings.HasSuffix(m.GetCommand(), core.BUser.Username) {
						reason, _ := database.GetMaintReason()
						reason = F(m.ChannelID(), "maint_reason", locales.Arg{"reason": reason})
						msg := F(m.ChannelID(), "maint", locales.Arg{"reason": reason})
						m.Reply(msg)
						gologging.Info("Sent maintenance notice to " + fmt.Sprint(m.SenderID()))
					}
					return tg.EndGroup
				}
			}
		}

		defer func() {
			if r := recover(); r != nil {
				gologging.Error("Recovered from panic: " + fmt.Sprint(r))
				handlePanic(r, m, true)
				err = fmt.Errorf("internal panic occurred")
			}
		}()

		if checkForHelpFlag(m) {
			cmd := getCommand(m)
			gologging.Debug("Help flag detected for command " + cmd)
			err = showHelpFor(m, cmd)
		} else {
			cmd := getCommand(m)
			gologging.Debug("Executing handler for command " + cmd)
			err = handler(m)
		}

		if err != nil {
			if errors.Is(err, tg.EndGroup) {
				gologging.Debug("Handler exited early (EndGroup)")
				return err
			}
			gologging.Error("Handler error: " + err.Error())
			handlePanic(err, m, false)
		} else {
			gologging.Info("Handler completed successfully for command " + getCommand(m))
		}

		return err
	}
}

func handlePanic(r, ctx interface{}, isPanic bool) {
	logger := gologging.GetLogger("Handlers")
	stack := html.EscapeString(string(debug.Stack()))

	var userMention, handlerType, chatInfo, messageInfo, errorMessage string
	var client *tg.Client

	switch c := ctx.(type) {
	case *tg.NewMessage:
		userMention = utils.MentionHTML(c.Sender)
		handlerType = "message"
		chatInfo = "ChatID: " + utils.IntToStr(c.ChatID())
		messageInfo = "Message: " + html.EscapeString(c.Text()) + "\nLink: " + c.Link()
		errorMessage = html.EscapeString(fmt.Sprint(r))
		client = c.Client

	case *tg.CallbackQuery:
		userMention = utils.MentionHTML(c.Sender)
		handlerType = "callback"
		chatInfo = "ChatID: " + utils.IntToStr(c.ChatID)
		messageInfo = "Data: " + html.EscapeString(c.DataString())
		errorMessage = html.EscapeString(fmt.Sprint(r))
		client = c.Client
	}

	logMsg := "🚨 Error in %s handler:\nFrom: %s\n%s\n%s\nError: `%v`"
	shortMsg := "<b>Error in %s handler</b>\n<b>From:</b> %s\n%s\n%s\n<b>Error:</b>\n<code>%s</code>"

	if isPanic {
		logMsg = "⚠️ Panic recovered in %s handler:\nFrom: %s\n%s\n%s\nError: `%v`\nStack:\n%s"
		shortMsg = "<b>⚠️ Panic in %s handler</b>\n<b>From:</b> %s\n%s\n%s\n<b>Error:</b>\n<code>%s</code>\n<pre>%s</pre>"
	}

	if isPanic {
		logger.ErrorF(logMsg, handlerType, userMention, chatInfo, messageInfo, r, stack)
	} else {
		logger.ErrorF(logMsg, handlerType, userMention, chatInfo, messageInfo, r)
	}

	if config.LoggerID != 0 && client != nil {
		var short string
		if isPanic {
			short = fmt.Sprintf(shortMsg, handlerType, userMention, chatInfo, messageInfo, errorMessage, stack)
		} else {
			short = fmt.Sprintf(shortMsg, handlerType, userMention, chatInfo, messageInfo, errorMessage)
		}

		if _, sendErr := client.SendMessage(config.LoggerID, short, &tg.SendOptions{ParseMode: "HTML"}); sendErr != nil {
			logger.ErrorF("Failed to send panic message to log chat: %v", sendErr)
		}
	}
}

func warnAndLeave(client *tg.Client, chatID int64) {
	text := F(chatID, "supergroup_needed", locales.Arg{"chat_id": chatID})
	_, err := client.SendMessage(
		chatID,
		text,
		&tg.SendOptions{
			ReplyMarkup: core.AddMeMarkup(core.BUser.Username),
			LinkPreview: false,
		},
	)
	if err != nil {
		gologging.ErrorF("Failed to send supergroup conversion message to chat %d: %v", chatID, err)
		return
	}

	go func() {
		time.Sleep(1 * time.Second)
		if err := client.LeaveChannel(chatID); err != nil {
			gologging.ErrorF("Failed to leave non-supergroup chatID=%d: %v", chatID, err)
		}
		core.UBot.LeaveChannel(chatID)
	}()
}

func formatDuration(sec int) string {
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60

	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s) // HH:MM:SS
	}
	return fmt.Sprintf("%02d:%02d", m, s) // MM:SS
}

func getCommand(m *tg.NewMessage) string {
	cmd := strings.SplitN(m.GetCommand(), "@", 2)[0]
	return cmd
}
