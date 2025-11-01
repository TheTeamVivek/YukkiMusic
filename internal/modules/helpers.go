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
	"github.com/amarnathcjd/gogram/telegram"

	"github.com/TheTeamVivek/YukkiMusic/config"
	"github.com/TheTeamVivek/YukkiMusic/internal/core"
	"github.com/TheTeamVivek/YukkiMusic/internal/database"
	"github.com/TheTeamVivek/YukkiMusic/internal/state"
	"github.com/TheTeamVivek/YukkiMusic/internal/utils"
)

var (
	superGroupFilter    = telegram.FilterFunc(FilterSuperGroup)
	adminFilter         = telegram.FilterFunc(FilterChatAdmins)
	authFilter          = telegram.FilterFunc(FilterAuthUsers)
	ignoreChannelFilter = telegram.FilterFunc(FilterChannel)
	sudoOnlyFilter      = telegram.FilterFunc(FilterSudo)
	ownerFilter         = telegram.FilterFunc(FilterOwner)
)

func bool_(b bool) *bool {
	return &b
}

func getEffectiveRoom(m *telegram.NewMessage, cplay bool) (*core.RoomState, error) {
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

func sendPlayLogs(m *telegram.NewMessage, track *state.Track, queued bool) {
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
		_, err = core.Bot.SendMedia(config.LoggerID, utils.CleanURL(track.Artwork), &telegram.MediaOptions{Caption: sb.String()})
	} else {
		_, err = core.Bot.SendMessage(config.LoggerID, sb.String())
	}
	if err != nil {
		gologging.Error("Failed to send logger msg: " + err.Error())
	}
}

func FilterOwner(m *telegram.NewMessage) bool {
	if config.OwnerID == 0 || m.SenderID() != config.OwnerID {
		if m.IsPrivate()|| strings.HasSuffix(m.GetCommand(), core.BUser.Username){
			m.Reply("⚠️ Only the bot owner can use this command.")
		}
		return false
	}
	return true
}

func FilterSudo(m *telegram.NewMessage) bool {
	is, _ := database.IsSudo(m.SenderID())

	if config.OwnerID == 0 || (m.SenderID() != config.OwnerID && !is) {
		if m.IsPrivate() || strings.HasSuffix(m.GetCommand(), core.BUser.Username){
			m.Reply("⚠️ Only sudo users or the bot owner can use this command.")
		}
		return false
	}

	return true
}

func FilterChannel(m *telegram.NewMessage) bool {
	if _, ok := m.Message.FromID.(*telegram.PeerChannel); ok {
		return false
	}
	return true
}

func FilterAuthUsers(m *telegram.NewMessage) bool {
	isAdmin, err := utils.IsChatAdmin(m.Client, m.ChannelID(), m.SenderID())
	if err == nil && isAdmin {
		return true
	}

	isAuth, err := database.IsAuthUser(m.ChannelID(), m.SenderID())
	if err == nil && isAuth {
		return true
	}

	m.Reply(
		"⚠️ <b>Access Denied</b>\n" +
			"Only <b>admins</b> or <b>authorized users</b> can control this actions.\n\n" +
			"If you recently became an admin, use /reload to refresh your permissions.",
	)
	return false
}

func FilterSuperGroup(m *telegram.NewMessage) bool {
	/*if m.Message.FromID == nil || (m.SenderChat != nil && m.SenderChat.ID != 0) {
		m.Reply("⚠️ You are using Anonymous Admin Mode.\n\n👉 Switch back to your user account to use commands.")
		return false
	}*/

	if !FilterChannel(m) {
		return false
	}
	// Validate chat type
	switch m.ChatType() {
	case telegram.EntityChat:
		// EntityChat can be basic group or supergroup — allow only supergroup
		if m.Channel != nil && !m.Channel.Broadcast {
			database.AddServed(m.ChannelID())
			return true // Supergroup
		}
		warnAndLeave(m.Client, m.ChatID()) // Basic group → leave
		database.DeleteServed(m.ChannelID())
		return false

	case telegram.EntityChannel:
		return false // Pure channel chat → ignore

	case telegram.EntityUser:
		m.Reply("⚠️ This command can only be used in groups.")
		database.AddServed(m.ChannelID(), true)
		return false // Private chat → warn
	}

	return false
}

func FilterChatAdmins(m *telegram.NewMessage) bool {
	isAdmin, err := utils.IsChatAdmin(m.Client, m.ChannelID(), m.SenderID())
	if err != nil || !isAdmin {
		m.Reply(
			"⚠️ <b>Permission Denied!</b>\n" +
				"Only <b>admins</b> can control this actions. You can still play songs 🎵.\n\n" +
				"If you believe you are an admin, please use /reload to refresh your admin status.",
		)
		return false
	}
	return true
}

func SafeCallbackHandler(handler func(*telegram.CallbackQuery) error) func(*telegram.CallbackQuery) error {
	return func(cb *telegram.CallbackQuery) (err error) {
		if is, _ := database.IsMaintenance(); is {
			if cb.Sender.ID != config.OwnerID {
				if ok, _ := database.IsSudo(cb.Sender.ID); !ok {
					cb.Answer("⚠️ I'm under maintenance at the moment and temporarily unavailable. Please check back later.", &telegram.CallbackOptions{Alert: true})
					return telegram.EndGroup
				}
			}
		}
		defer func() {
			if r := recover(); r != nil {
				handlePanic(r, cb, true)
				err = fmt.Errorf("internal error occurred")
			}
		}()
		err = handler(cb)
		if err != nil {
			if errors.Is(err, telegram.EndGroup) {
				return err
			}
			handlePanic(err, cb, false)
			err = fmt.Errorf("internal error occurred")
		}
		return err
	}
}

func SafeMessageHandler(handler func(*telegram.NewMessage) error) func(*telegram.NewMessage) error {
	return func(m *telegram.NewMessage) (err error) {
		if is, _ := database.IsMaintenance(); is {
			if m.SenderID() != config.OwnerID {
				if ok, _ := database.IsSudo(m.SenderID()); !ok {
					if m.ChatType() == telegram.EntityUser || strings.HasSuffix(m.GetCommand(), core.BUser.Username) {
						msg := "⚠️ I'm under maintenance at the moment and temporarily unavailable. Please check back later."

						if reason, err := database.GetMaintReason(); err == nil && reason != "" {
							msg += "\n\n<i>📝 Reason: " + reason + "</i>"
						}
						m.Reply(msg)

					}
					return telegram.EndGroup
				}
			}
		}

		defer func() {
			if r := recover(); r != nil {
				handlePanic(r, m, true)
				err = fmt.Errorf("internal error occurred")
			}
		}()
		err = handler(m)
		if err != nil {
			if errors.Is(err, telegram.EndGroup) {
				return err
			}
			handlePanic(err, m, false)
			err = fmt.Errorf("internal error occurred")
		}
		return err
	}
}

func handlePanic(r, ctx interface{}, isPanic bool) {
	logger := gologging.GetLogger("Handlers")
	stack := html.EscapeString(string(debug.Stack()))

	var userMention, handlerType, chatInfo, messageInfo, errorMessage string
	var client *telegram.Client

	switch c := ctx.(type) {
	case *telegram.NewMessage:
		userMention = utils.MentionHTML(c.Sender)
		handlerType = "message"
		chatInfo = fmt.Sprintf("ChatID: %d", c.ChatID())
		messageInfo = fmt.Sprintf("Message: %s\nLink: %s", html.EscapeString(c.Text()), c.Link())
		errorMessage = html.EscapeString(fmt.Sprintf("%v", r))
		client = c.Client

	case *telegram.CallbackQuery:
		userMention = utils.MentionHTML(c.Sender)
		handlerType = "callback"
		chatInfo = fmt.Sprintf("ChatID: %d", c.ChatID)
		messageInfo = fmt.Sprintf("Data: %s", html.EscapeString(c.DataString()))
		errorMessage = html.EscapeString(fmt.Sprintf("%v", r))
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

		if _, sendErr := client.SendMessage(config.LoggerID, short, &telegram.SendOptions{ParseMode: "HTML"}); sendErr != nil {
			logger.ErrorF("Failed to send panic message to log chat: %v", sendErr)
		}
	}
}

func warnAndLeave(client *telegram.Client, chatID int64) {
	text := fmt.Sprintf(
		"This chat (ID: <code>%d</code>) is not a supergroup yet.\n"+
			"<b>⚠️ Please convert this chat to a supergroup then add me as admin.</b>\n\n"+
			"If you don't know how to convert, use this guide:\n"+
			"🔗 <a href=\"https://te.legra.ph/How-to-Convert-a-Group-to-a-Supergroup-01-02\">How to convert to a SuperGroup</a> \n\n"+
			"If you have any questions, join our support group:",
		chatID,
	)

	_, err := client.SendMessage(
		chatID,
		text,
		&telegram.SendOptions{
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

func getCommand(m *telegram.NewMessage) string {
	cmd := strings.SplitN(m.GetCommand(), "@", 2)[0]
	return cmd
}
