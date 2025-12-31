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
	"runtime/debug"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
	state "main/internal/core/models"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

var downloadCancels = make(map[int64]context.CancelFunc)

func getEffectiveRoom(m *tg.NewMessage, cplay bool) (*core.RoomState, error) {
	chatID := m.ChannelID()

	if cplay {
		cplayID, err := database.GetCPlayID(chatID)
		if err != nil || cplayID == 0 {
			return nil, errors.New(F(chatID, "cplay_id_not_set"))
		}
		chatID = cplayID
	}
	ass, err := core.Assistants.ForChat(chatID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get assistant for you chat: %w", err)
	}
	r, _ := core.GetRoom(chatID, ass, true)
	return r, nil
}

func sendPlayLogs(m *tg.NewMessage, track *state.Track, queued bool) {
	if config.LoggerID == 0 || config.LoggerID == m.ChatID() || config.LoggerID == m.ChannelID() {
		return
	}

	if is, err := database.IsLoggerEnabled(); err != nil {
		gologging.Error("Failed to get IsLoggerEnabled: " + err.Error())
		return
	} else if !is {
		return
	}

	var (
		sb  strings.Builder
		err error
	)

	chatID := m.ChannelID()

	header := F(chatID, "logger_playback_started")
	if queued {
		header = F(chatID, "logger_playback_queued")
	}

	// Header
	sb.WriteString("🎵 ")
	if m.Channel.Username != "" {
		fmt.Fprintf(&sb, "<b><a href=\"%s\">%s</a></b>\n\n", m.Link(), header)
	} else {
		fmt.Fprintf(&sb, "<b><u>%s</u></b>\n\n", header)
	}

	// artwork block
	if track.Artwork != "" {
		sb.WriteString("<blockquote>")
	}

	// Track
	fmt.Fprintf(&sb,
		"<b>%s</b> <a href=\"%s\">%s</a>\n",
		F(chatID, "logger_track"),
		track.URL,
		utils.ShortTitle(track.Title),
	)

	// Source
	fmt.Fprintf(&sb,
		"<b>%s</b> %s\n",
		F(chatID, "logger_source"),
		string(track.Source),
	)

	// Group
	fmt.Fprintf(&sb, "<b>%s</b> ", F(chatID, "logger_group"))
	if m.Channel.Username != "" {
		fmt.Fprintf(&sb, "@%s", m.Channel.Username)
	} else {
		sb.WriteString(m.Channel.Title)
	}
	fmt.Fprintf(&sb, " (%d)\n", m.ChannelID())

	// Requested by
	fmt.Fprintf(&sb, "<b>%s</b> ", F(chatID, "logger_requested_by"))
	if m.Sender.Username != "" {
		fmt.Fprintf(&sb, "@%s", m.Sender.Username)
	} else {
		sb.WriteString(utils.MentionHTML(m.Sender))
	}
	fmt.Fprintf(&sb, " (<code>%d</code>)\n", m.Sender.ID)

	// Timestamp
	fmt.Fprintf(&sb, "<b>%s</b> %s",
		F(chatID, "logger_timestamp"),
		time.Now().Format("2006-01-02 15:04:05"),
	)

	// Sending
	if track.Artwork != "" {
		sb.WriteString("\n</blockquote>")
		_, err = core.Bot.SendMedia(
			config.LoggerID,
			utils.CleanURL(track.Artwork),
			&tg.MediaOptions{Caption: sb.String()},
		)
	} else {
		_, err = core.Bot.SendMessage(config.LoggerID, sb.String())
	}

	if err != nil {
		gologging.Error("Failed to send logger msg: " + err.Error())
	}
}

func F(chatID int64, key string, values ...locales.Arg) string {
	lang, err := database.GetChatLanguage(chatID)
	if err != nil {
		gologging.Error("Failed to get language for " + utils.IntToStr(chatID) + " Got error " + err.Error())
		lang = config.DefaultLang
	}
	return FWithLang(lang, key, values...)
}

func FWithLang(lang, key string, values ...locales.Arg) string {
	var val locales.Arg
	if len(values) > 0 {
		val = values[0]
	}
	return locales.Get(lang, key, val)
}

func isLogger() (l bool) {
	var err error
	l, err = database.IsLoggerEnabled()
	if err != nil {
		gologging.Error("Failed to get IsLoggerEnabled, Err: " + err.Error())
	}
	return l
}

func SafeCallbackHandler(handler func(*tg.CallbackQuery) error) func(*tg.CallbackQuery) error {
	return func(cb *tg.CallbackQuery) (err error) {
		if is, _ := database.IsMaintenance(); is {
			if cb.Sender.ID != config.OwnerID {
				if ok, _ := database.IsSudo(cb.Sender.ID); !ok {
					cb.Answer(F(cb.ChannelID(), "maint", locales.Arg{"reason": ""}), &tg.CallbackOptions{Alert: true})
					return tg.ErrEndGroup
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
			if errors.Is(err, tg.ErrEndGroup) {
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
					return tg.ErrEndGroup
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
			if errors.Is(err, tg.ErrEndGroup) {
				gologging.Debug("Handler exited early (ErrEndGroup)")
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
	stack := html.EscapeString(string(debug.Stack()))

	var userMention, handlerType, chatInfo, messageInfo, errorMessage string
	var client *tg.Client

	switch c := ctx.(type) {
	case *tg.NewMessage:
		userMention = utils.MentionHTML(c.Sender)
		handlerType = "message"
		chatInfo = "ChatID: " + utils.IntToStr(c.ChannelID())
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
		gologging.ErrorF(logMsg, handlerType, userMention, chatInfo, messageInfo, r, stack)
	} else {
		gologging.ErrorF(logMsg, handlerType, userMention, chatInfo, messageInfo, r)
	}

	if config.LoggerID != 0 && client != nil {
		var short string
		if isPanic {
			short = fmt.Sprintf(shortMsg, handlerType, userMention, chatInfo, messageInfo, errorMessage, stack)
		} else {
			short = fmt.Sprintf(shortMsg, handlerType, userMention, chatInfo, messageInfo, errorMessage)
		}

		gologging.Error(short)
		if _, sendErr := client.SendMessage(config.LoggerID, short, &tg.SendOptions{ParseMode: "HTML"}); sendErr != nil {
			gologging.ErrorF("Failed to send panic message to log chat: %v", sendErr)
		}
	}
}

func warnAndLeave(client *tg.Client, chatID int64) {
	text := F(chatID, "supergroup_needed", locales.Arg{"chat_id": chatID})
	_, err := client.SendMessage(
		chatID,
		text,
		&tg.SendOptions{
			ReplyMarkup: core.AddMeMarkup(chatID),
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
		core.Assistants.WithAssistant(chatID, func(ass *core.Assistant) { ass.Client.LeaveChannel(chatID) })
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
