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
	"fmt"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"github.com/TheTeamVivek/YukkiMusic/config"
	"github.com/TheTeamVivek/YukkiMusic/internal/core"
	"github.com/TheTeamVivek/YukkiMusic/internal/database"
	"github.com/TheTeamVivek/YukkiMusic/internal/utils"
)

func handleParticipantUpdate(p *telegram.ParticipantUpdate) error {
	shouldIgnore := false
	if is, _ := database.IsMaintenance(); is && p.ActorID() != config.OwnerID {
		if ok, _ := database.IsSudo(p.ActorID()); !ok {
			shouldIgnore = true
		}
	}

	if p.ChannelID() == 0 {
		return nil
	}

	logger := gologging.GetLogger("PU")

	chatID, err := utils.GetPeerID(p.Client, p.ChannelID())
	if err != nil {
		logger.ErrorF("Failed to resolve peer for chatID=%d, Error=%v", p.ChannelID(), err)
		return err
	}

	// Assistant State Logic
	if p.UserID() == core.UbUser.ID {
		s, _ := core.GetChatState(chatID, true)

		if p.IsLeft() {
			s.SetAssistantPresence(bool_(false))
			s.SetAssistantBanned(bool_(false))

		}
		if p.IsJoined() {
			s.SetAssistantPresence(bool_(true))
			s.SetAssistantBanned(bool_(false))

		}

		// Check if Assistant is restricted/banned/kicked
		if p.IsBanned() || p.IsKicked() || (p.New != nil && isAssistantRestricted(p.New)) {
			logger.DebugF("Assistant restricted in chatID %d", chatID)

			s.SetAssistantPresence(bool_(false))
			core.DeleteRoom(chatID)

			// Try Unban, if fails - mark banned + notify
			if ok, err := p.Unban(); err != nil || !ok {
				s.SetAssistantBanned(bool_(true))

				if !shouldIgnore {
					_, err := p.Client.SendMessage(chatID, fmt.Sprintf(
						"⚠️ <b>Assistant Restricted</b>\n\nThe assistant %s (ID: <code>%d</code>) has been <b>banned</b> or restricted in this chat.\nI cannot play music, manage the queue, or clean up tracks while this lasts.\n\n<i>Unban the assistant to restore all music features 🎵</i>",
						utils.MentionHTML(core.UbUser), core.UbUser.ID,
					))
					if err != nil {
						logger.ErrorF("Failed to send assistant restricted warning message ChatID: %d, Error %v", chatID, err)
					}
				}
			}

		}

		// Fallback State Detection (Initial Join or Missing Cache)
		if s.GetAssistantPresence() == nil || s.GetAssistantBanned() == nil {
			member, err := p.Client.GetChatMember(chatID, core.UbUser.ID)
			if err != nil {
				if telegram.MatchError(err, "USER_NOT_PARTICIPANT") {
					s.SetAssistantPresence(bool_(false))
					s.SetAssistantBanned(bool_(false))
				} else {
					logger.ErrorF("Error getting assistant membership; ChatId: %d, Error: %v", chatID, err)
				}
			} else {
				switch member.Status {
				case telegram.Kicked, telegram.Restricted:
					s.SetAssistantPresence(bool_(false))
					s.SetAssistantBanned(bool_(true))
				case telegram.Left:
					s.SetAssistantPresence(bool_(false))
					s.SetAssistantBanned(bool_(false))
				case telegram.Admin, telegram.Member:
					s.SetAssistantPresence(bool_(true))
					s.SetAssistantBanned(bool_(false))
				}
			}
		}
	}

	// If main bot is demoted → leave
	if p.IsDemoted() {
		if p.UserID() == core.BUser.ID {
			core.DeleteRoom(chatID)
			core.DeleteChatState(chatID)
			p.Client.SendMessage(chatID, "<b>⚠️ Permission Lost</b>\n\nI’ve been <b>demoted from admin</b> and can’t perform my tasks anymore.\nWithout proper rights, I’m unable to function here, so I’ll be <i>leaving the chat now</i>.\n\n<b>Goodbye 👋</b>")
			p.Client.LeaveChannel(chatID)
			core.UBot.LeaveChannel(chatID)
			return nil
		}
		utils.RemoveChatAdmin(p.Client, chatID, p.UserID())
	}

	// On Promotion → Update cache
	if p.IsPromoted() {
		utils.AddChatAdmin(p.Client, chatID, p.UserID())
	}

	return nil
}

func handleActions(m *telegram.NewMessage) error {
	logger := gologging.GetLogger("Actions")
	chatID := m.ChannelID()

	// Only allow in super groups
	if m.ChatType() == telegram.EntityChat && (m.Channel == nil || !m.Channel.Megagroup) {
		warnAndLeave(m.Client, chatID)
		return telegram.EndGroup
	}

	isMaintenance, _ := database.IsMaintenance()

	switch action := m.Action.(type) {

	case *telegram.MessageActionGroupCall:
		if isMaintenance {
			logger.DebugF("Maintenance mode active, skipping VC updates in %d", chatID)
			return telegram.EndGroup
		}

		core.DeleteRoom(chatID)
		s, _ := core.GetChatState(chatID, true)
		if action.Duration == 0 {
			s.SetVoiceChatStatus(bool_(true))
			m.Respond("📢 Voice chat started!\nUse /play <song> to play music")
			logger.DebugF("Voice chat started in %d", chatID)
		} else {
			s.SetVoiceChatStatus(bool_(false))
			m.Respond("📴 Voice chat ended!\nAll queues cleared")
			logger.DebugF("Voice chat ended in %d", chatID)
		}

	case *telegram.MessageActionChatAddUser:
		for _, uid := range action.Users {
			logger.DebugF("User added to chatID %d: %d", chatID, uid)
			if uid == core.BUser.ID {
				if is, _ := database.IsMaintenance(); is && m.SenderID() != config.OwnerID {
					if ok, _ := database.IsSudo(m.SenderID()); !ok {
						msg := "⚠️ I'm currently under maintenance and will be back soon. Thanks for your patience! ⏰"
						if reason, err := database.GetMaintReason(); err == nil && reason != "" {
							msg += "\n\n<i>📝 Reason: " + reason + "</i>"
						}
						m.Reply(msg)

						m.Client.LeaveChannel(chatID)
						logger.DebugF("Bot left chatID %d due to maintenance", chatID)
						return telegram.EndGroup
					}
				}

				// Normal join flow
				m.Respond("🎵 Thanks for adding me! Use /play <song> to start playing music.")
				logger.DebugF("Bot added to chatID %d", chatID)
				database.AddServed(chatID)
				return telegram.EndGroup
			}
		}

	case *telegram.MessageActionChatDeleteUser:
		logger.DebugF("User removed from chatID %d: %d", chatID, action.UserID)
		if action.UserID == core.BUser.ID {
			core.UBot.LeaveChannel(chatID)
			core.DeleteRoom(chatID)
			core.DeleteChatState(chatID)
			database.DeleteServed(chatID)
			logger.DebugF("Bot left chatID %d", chatID)
			return telegram.EndGroup
		}
	}

	return telegram.EndGroup
}

func isAssistantRestricted(newParticipant telegram.ChannelParticipant) bool {
	_, left := newParticipant.(*telegram.ChannelParticipantLeft)
	_, banned := newParticipant.(*telegram.ChannelParticipantBanned)
	return left || banned
}
