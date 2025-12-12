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
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/platforms"
	"main/internal/utils"
)

// TODO: Impment Cancel correctly
// flood_seconds -> fix float

func cancelHandler(cb *tg.CallbackQuery) error {
	chatID := cb.ChannelID()
	opt := &tg.CallbackOptions{Alert: true}

	if isAdmin, err := utils.IsChatAdmin(cb.Client, chatID, cb.SenderID); err != nil || !isAdmin {
		cb.Answer(
			F(chatID, "only_admin_or_auth_cb"),
			opt,
		)
		return tg.EndGroup
	}

	if cancel, ok := downloadCancels[chatID]; ok {
		cancel()
		delete(downloadCancels, chatID)
		cb.Answer("Download canceled.", opt)
	} else {
		cb.Answer("No download to cancel.", opt)
	}
	return tg.EndGroup
}

func closeHandler(cb *tg.CallbackQuery) error {
	cb.Answer("")
	cb.Delete()
	return tg.EndGroup
}

func emptyCBHandler(cb *tg.CallbackQuery) error {
	cb.Answer("")
	return tg.EndGroup
}

func roomHandle(cb *tg.CallbackQuery) error {
	opt := &tg.CallbackOptions{Alert: true}
	data := string(cb.Data)
	updateType := strings.TrimPrefix(data, "croom:")
	updateType = strings.TrimPrefix(data, "room:")

	if updateType == "" {
		gologging.WarnF("Missing action in data: %s", data)
		cb.Answer("⚠️ Invalid request.", opt)
		return tg.EndGroup
	}

	chatID := cb.ChannelID()

	var r *core.RoomState
	var ok bool

	if strings.HasPrefix(cb.DataString(), "croom:") {
		realChatID, err := database.GetCPlayID(chatID)
		if err != nil {
			gologging.ErrorF("Failed to get chat ID for cplay ID %d: %v", chatID, err)
			cb.Answer(F(chatID, "room_not_linked"), opt)
			return tg.EndGroup
		}
		chatID = realChatID
	}
	if ass, err := core.Assistants.ForChat(chatID); err != nil {
    gologging.ErrorF("Failed to get Assistant for %d: %v", chatID, err)
    cb.Answer(fmt.Sprintf("Failed to get Assistant for %d: %v", chatID, err), opt)
			return err
  } else {
	r, ok = core.GetRoom(chatID, ass)
    
  }
	
	if !ok || !r.IsActiveChat() {
		cb.Answer(F(chatID, "room_not_active_cb"), opt)
		if _, err := cb.Edit(F(chatID, "room_not_active")); err != nil {
			gologging.ErrorF("Edit error: %v", err)
		}
		return tg.EndGroup
	}
	if isAdmin, err := utils.IsChatAdmin(cb.Client, chatID, cb.SenderID); err != nil || !isAdmin {
		cb.Answer(
			F(chatID, "only_admin_or_auth_cb"),
			opt,
		)
		return tg.EndGroup
	}

	key := fmt.Sprintf("room:%d:%d", cb.Sender.ID, chatID)
	if remaining := utils.GetFlood(key); remaining > 0 {
		cb.Answer(F(chatID, "flood_seconds", locales.Arg{"duration": remaining.Seconds()}), opt)
		return tg.EndGroup
	}
	utils.SetFlood(key, 5*time.Second)

	switch {
	case updateType == "pause":
		gologging.InfoF("Callback → pause, chatID=%d", chatID)

		/*if r.IsMuted() {
			cb.Answer("🔇 The chat is muted. Please unmute first.", opt)
			return tg.EndGroup
		}*/

		if r.IsPaused() {
			remaining := r.RemainingResumeDuration()
			msg := utils.IfElse(
				remaining > 0,
				F(chatID, "room_already_paused_auto", locales.Arg{
					"duration": formatDuration(int(remaining.Seconds())),
				}),
				F(chatID, "room_already_paused"),
			)
			cb.Answer(msg, opt)
			return tg.EndGroup
		}

		if _, pauseErr := r.Pause(); pauseErr != nil {
			gologging.ErrorF("Pause failed: %v", pauseErr)
			cb.Answer(F(chatID, "room_pause_failed", locales.Arg{"error": pauseErr.Error()}), opt)
			return tg.EndGroup
		}
		if r.IsMuted() {
			r.Unmute() // unmute playback
		}

		cb.Answer("⏸️ Track paused at "+formatDuration(r.Position()), opt)

		mention := utils.MentionHTML(cb.Sender)
		track := r.Track()
		safeTitle := html.EscapeString(track.Title)

		msgText := fmt.Sprintf(
			"<b>⏸️ Track Paused</b>\n\n"+
				"<b>▫ Track:</b> <a href=\"%s\">%s</a>\n"+
				"<b>▫ Position:</b> %s / %s\n"+
				"<b>▫ Paused by:</b> %s",
			track.URL,
			utils.ShortTitle(safeTitle, 25),
			formatDuration(r.Position()),
			formatDuration(track.Duration),
			mention,
		)

		if _, err := cb.Edit(msgText, &tg.SendOptions{
			ParseMode:   "HTML",
			ReplyMarkup: core.GetPlayMarkup(r, false),
		}); err != nil {
			gologging.ErrorF("Edit error: %v", err)
		}

	case updateType == "resume":
		gologging.InfoF("Callback → resume, chatID=%d", chatID)

		if !r.IsPaused() {
			cb.Answer("ℹ️ Track is already playing.", opt)
			return tg.EndGroup
		}

		if _, err := r.Resume(); err != nil {
			gologging.ErrorF("Resume failed: %v", err)
			if _, e := cb.Answer("❌ Failed to resume playback.", opt); e != nil {
				gologging.ErrorF("Answer error: %v", e)
			}
			return tg.EndGroup
		}

		if _, err := cb.Answer(fmt.Sprintf("▶️ Resumed at %s.", formatDuration(r.Position())), opt); err != nil {
			gologging.ErrorF("Answer error: %v", err)
		}

		mention := utils.MentionHTML(cb.Sender)
		track := r.Track()

		msgText := fmt.Sprintf(
			"<b>🎵 Now Playing:</b>\n\n"+
				"<b>▫ Track:</b> <a href=\"%s\">%s</a>\n"+
				"<b>▫ Duration:</b> %s\n"+
				"<b>▫ Resumed by:</b> %s",
			track.URL,
			html.EscapeString(utils.ShortTitle(track.Title, 25)),
			formatDuration(track.Duration),
			mention,
		)

		if _, err := cb.Edit(msgText, &tg.SendOptions{
			ParseMode:   "HTML",
			ReplyMarkup: core.GetPlayMarkup(r, false),
		}); err != nil {
			gologging.ErrorF("Edit error: %v", err)
		}

	case updateType == "replay":
		gologging.InfoF("Callback → replay, chatID=%d", chatID)

		mystic, err := cb.Respond("🔁 <b>Replaying current track...</b>")
		if err != nil {
			gologging.ErrorF("Failed to send replay message: %v", err)
			return tg.EndGroup
		}

		if err := r.Replay(); err != nil {
			gologging.ErrorF("Replay failed: %v", err)
			utils.EOR(mystic, fmt.Sprintf("❌ <b>Replay Failed</b>\nError: <code>%v</code>", err))
			cb.Answer("❌ Failed to replay track.", opt)
			return tg.EndGroup
		}
		track := r.Track()

		trackTitle := html.EscapeString(utils.ShortTitle(track.Title, 25))
		totalDuration := formatDuration(track.Duration)

		msgText := fmt.Sprintf(
			"<b>🎵 Now Playing:</b>\n\n"+
				"<b>▫ Track:</b> <a href=\"%s\">%s</a>\n"+
				"<b>▫ Duration:</b> %s\n"+
				"<b>▫ Requested by:</b> %s\n"+
				"<b>▫ Replayed by:</b> %s",
			track.URL,
			trackTitle,
			totalDuration,
			track.Requester,
			utils.MentionHTML(cb.Sender),
		)

		cb.Answer("🔁 Track replayed.", opt)

		optSend := &tg.SendOptions{
			ParseMode:   "HTML",
			ReplyMarkup: core.GetPlayMarkup(r, false),
		}

		if track.Artwork != "" {
			optSend.Media = utils.CleanURL(track.Artwork)
		}

		mystic, _ = utils.EOR(mystic, msgText, optSend)
		r.SetMystic(mystic)

		if _, err := cb.Edit(fmt.Sprintf("🔁 Track replayed by %s.", utils.MentionHTML(cb.Sender))); err != nil {
			gologging.ErrorF("Edit error: %v", err)
		}

	case strings.HasPrefix(updateType, "seekback_"):
		parts := strings.Split(updateType, "_")
		seconds, err := strconv.Atoi(parts[1])
		if err != nil {
			cb.Answer("⚠️ Invalid seek value.", opt)
			return tg.EndGroup
		}

		// Clamp to start
		if r.Position() <= seconds {
			r.Seek(-int(r.Position()))
		} else {
			r.Seek(-seconds)
		}

		cb.Answer(fmt.Sprintf("⏪ Sought back %d seconds", seconds), opt)
		rp(cb, fmt.Sprintf("⏪ Sought back %d seconds — by %s", seconds, utils.MentionHTML(cb.Sender)))

	case strings.HasPrefix(updateType, "seek_"):
		parts := strings.Split(updateType, "_")
		seconds, err := strconv.Atoi(parts[1])
		if err != nil {
			cb.Answer("⚠️ Invalid seek value.", opt)
			return tg.EndGroup
		}

		// Warn if near end
		if (r.Track().Duration - r.Position()) <= seconds {
			cb.Answer(fmt.Sprintf("⚠️ Cannot seek forward %d seconds — about to reach end.", seconds), opt)

			return tg.EndGroup
		}

		r.Seek(seconds)
		cb.Answer(fmt.Sprintf("⏩ Sought %d seconds", seconds), opt)
		rp(cb, fmt.Sprintf("⏩ Sought %d seconds — by %s", seconds, utils.MentionHTML(cb.Sender)))

	case updateType == "skip":
		gologging.InfoF("Callback → skip, chatID=%d", chatID)

		mention := utils.MentionHTML(cb.Sender)

		if len(r.Queue()) == 0 && r.Loop() == 0 {
			r.Destroy()
			msgText := fmt.Sprintf(
				"⏹️ <b>Playback stopped.</b>\nQueue is empty.\n\n▫ Skipped by: %s",
				mention,
			)
			if _, err := cb.Edit(msgText, &tg.SendOptions{ParseMode: "HTML"}); err != nil {
				gologging.ErrorF("Edit error: %v", err)
			}
			if _, err := cb.Answer("⏹️ Playback stopped — queue empty.", opt); err != nil {
				gologging.ErrorF("Answer error: %v", err)
			}
			return tg.EndGroup
		}

		t := r.NextTrack()

		mystic, err := cb.Respond("📥 Downloading your next track...")
		if err != nil {
			gologging.ErrorF("Failed to send msg: %v", err)
		}

		path, err := platforms.Download(context.Background(), t, mystic)
		if err != nil {
			gologging.ErrorF("Download failed for %s: %v", t.URL, err)
			utils.EOR(mystic, "❌ Failed to download next track.")
			if _, err := cb.Answer("❌ Failed to download next track.", opt); err != nil {
				gologging.ErrorF("Answer error: %v", err)
			}
			return tg.EndGroup
		}

		if err := r.Play(t, path); err != nil {
			gologging.ErrorF("Play error: %v", err)
			utils.EOR(mystic, "❌ Failed to play next track.")
			if _, err := cb.Answer("❌ Failed to play next track.", opt); err != nil {
				gologging.ErrorF("Answer error: %v", err)
			}
			return tg.EndGroup
		}

		if _, err := cb.Answer("⏭️ Track skipped.", opt); err != nil {
			gologging.ErrorF("Answer error: %v", err)
		}

		_, err = cb.Delete()
		if err != nil {
			gologging.ErrorF("Delete error: %v", err)
		}

		title := utils.ShortTitle(t.Title, 25)
		safeTitle := html.EscapeString(title)
		msgText := fmt.Sprintf(
			"<b>🎵 Now Playing:</b>\n\n"+
				"<b>▫ Track:</b> <a href=\"%s\">%s</a>\n"+
				"<b>▫ Duration:</b> %s\n"+
				"<b>▫ Requested by:</b> %s",
			t.URL,
			safeTitle,
			formatDuration(t.Duration),
			t.Requester,
		)

		opt := &tg.SendOptions{
			ParseMode:   "HTML",
			ReplyMarkup: core.GetPlayMarkup(r, false),
		}

		if t.Artwork != "" {
			opt.Media = utils.CleanURL(t.Artwork)
		}

		mystic, _ = utils.EOR(mystic, msgText, opt)

		if _, err := mystic.Reply(fmt.Sprintf("⏭️ Skipped by %s", mention)); err != nil {
			gologging.ErrorF("Reply error: %v", err)
		}
		r.SetMystic(mystic)
		return tg.EndGroup

	case updateType == "stop":

		gologging.InfoF("Callback → stop, chatID=%d", chatID)

		r.Destroy()

		if _, err := cb.Answer("⏹️ Playback stopped.", opt); err != nil {
			gologging.ErrorF("Answer error: %v", err)
		}

		if _, err := cb.Edit(fmt.Sprintf("⏹️ Playback stopped and cleared by %s.", utils.MentionHTML(cb.Sender))); err != nil {
			gologging.ErrorF("Edit error: %v", err)
		}

	default:
		gologging.WarnF("Unknown callback type: %s", updateType)
		if _, err := cb.Answer("⚠️ Unknown action.", opt); err != nil {
			gologging.ErrorF("Answer error: %v", err)
		}
	}

	return tg.EndGroup
}

func rp(c *tg.CallbackQuery, t string) {
	msg, err := c.GetMessage()
	if err != nil {
		return
	}
	msg.Reply(t)
}
