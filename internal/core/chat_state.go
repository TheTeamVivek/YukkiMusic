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
package core

import (
	"errors"
	"fmt"
	"sync"

	"github.com/amarnathcjd/gogram/telegram"
)

type ChatState struct {
	mu               *sync.RWMutex
	ChatID           int64
	AssistantPresent *bool
	AssistantBanned  *bool
	VoiceChatStatus  *bool
	InviteLink       string
}

var (
	chMutex    = &sync.Mutex{}
	ChatStates = make(map[int64]*ChatState)
)

func GetChatState(chatID int64, create ...bool) (*ChatState, bool) {
	chMutex.Lock()
	defer chMutex.Unlock()

	state, ok := ChatStates[chatID]

	if !ok && len(create) > 0 && create[0] {
		state = &ChatState{
			mu:     &sync.RWMutex{},
			ChatID: chatID,
		}
		ChatStates[chatID] = state
	}
	return state, ok
}

func DeleteChatState(chatID int64) {
	chMutex.Lock()
	defer chMutex.Unlock()

	_, ok := ChatStates[chatID]

	if ok {
		delete(ChatStates, chatID)
	}
}

func (cs *ChatState) SetAssistantPresence(present *bool) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.AssistantPresent = present
}

func (cs *ChatState) SetAssistantBanned(banned *bool) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.AssistantBanned = banned
}

func (cs *ChatState) SetVoiceChatStatus(active *bool) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.VoiceChatStatus = active
}

func (cs *ChatState) SetInviteLink(link string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.InviteLink = link
}

func (cs *ChatState) GetAssistantPresence() *bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.AssistantPresent
}

func (cs *ChatState) GetAssistantBanned() *bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.AssistantBanned
}

func (cs *ChatState) GetVoiceChatStatus() *bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.VoiceChatStatus
}

func (cs *ChatState) GetInviteLink() string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.InviteLink
}

var (
	ErrNoActiveVoiceChat        = errors.New("no active voice chat")
	ErrAdminPermissionRequired  = errors.New("admin permission required")
	ErrFetchFailed              = errors.New("failed to fetch chat info")
	ErrAssistantBanned          = errors.New("assistant banned in chat")
	ErrAssistantInviteLinkFetch = errors.New("failed to fetch invite link")
	ErrAssistantJoinRejected    = errors.New("invite link is invalid or expired")
	ErrAssistantJoinRateLimited = errors.New("assistant cannot join, rate limited")
	ErrAssistantJoinRequestSent = errors.New("assistant join request sent")
	ErrPeerResolveFailed        = errors.New("failed to resolve peer")
	ErrAssistantInviteFailed    = errors.New("assistant failed to join this chat")
)

func boolToPtr(b bool) *bool {
	return &b
}

// GetVoiceChatStatus returns whether a voice chat is active in the given chat.
func GetVoiceChatStatus(chatID int64, force ...bool) (bool, error) {
	s, _ := GetChatState(chatID, true)
	doFetch := len(force) > 0 && force[0]

	if s.GetVoiceChatStatus() == nil || doFetch {
		peer, err := Bot.ResolvePeer(chatID)
		if err != nil {
			return false, fmt.Errorf("%w: %v", ErrFetchFailed, err)
		}

		chPeer, ok := peer.(*telegram.InputPeerChannel)
		if !ok {
			return false, fmt.Errorf("%w: chatID %d is not an InputPeerChannel, got type %T", ErrFetchFailed, chatID, peer)
		}

		fullChat, err := Bot.ChannelsGetFullChannel(&telegram.InputChannelObj{
			ChannelID:  chPeer.ChannelID,
			AccessHash: chPeer.AccessHash,
		})
		if err != nil {
			if telegram.MatchError(err, "CHANNEL_INVALID") || telegram.MatchError(err, "CHANNEL_PRIVATE") {
				return false, ErrAdminPermissionRequired
			}
			return false, fmt.Errorf("%w: %v", ErrFetchFailed, err)
		}

		chat := fullChat.FullChat.(*telegram.ChannelFull)
		s.SetVoiceChatStatus(boolToPtr(chat.Call != nil))

		if chat.ExportedInvite != nil && s.GetInviteLink() == "" {
			if l, ok := chat.ExportedInvite.(*telegram.ChatInviteExported); ok && l.Link != "" {
				s.SetInviteLink(l.Link)
			}
		}
	}

	if !*s.GetVoiceChatStatus() {
		return false, ErrNoActiveVoiceChat
	}
	return true, nil
}

// GetAssistantStatus checks the presence and status of the assistant in a chat.
func GetAssistantStatus(chatID int64, force ...bool) (bool, error) {
	s, _ := GetChatState(chatID, true)
	doFetch := len(force) > 0 && force[0]

	if s.GetAssistantPresence() == nil || s.GetAssistantBanned() == nil || doFetch {
		member, err := Bot.GetChatMember(chatID, UbUser.ID)
		if err != nil {
			switch {
			case telegram.MatchError(err, "USER_NOT_PARTICIPANT"), telegram.MatchError(err, "PARTICIPANT_ID_INVALID"):
				s.SetAssistantPresence(boolToPtr(false))
				s.SetAssistantBanned(boolToPtr(false))
			case telegram.MatchError(err, "CHAT_ADMIN_REQUIRED"), telegram.MatchError(err, "CHANNEL_PRIVATE"):
				return false, ErrAdminPermissionRequired
			default:
				return false, fmt.Errorf("%w: %v", ErrFetchFailed, err)
			}
		} else if member == nil {
			s.SetAssistantPresence(boolToPtr(false))
			s.SetAssistantBanned(boolToPtr(false))
		} else {
			switch member.Status {
			case telegram.Kicked, telegram.Restricted:
				s.SetAssistantPresence(boolToPtr(false))
				s.SetAssistantBanned(boolToPtr(true))
			case telegram.Left:
				s.SetAssistantPresence(boolToPtr(false))
				s.SetAssistantBanned(boolToPtr(false))
			case telegram.Admin, telegram.Member:
				s.SetAssistantPresence(boolToPtr(true))
				s.SetAssistantBanned(boolToPtr(false))
			}
		}
	}

	if *s.GetAssistantBanned() {
		return false, ErrAssistantBanned
	}

	if !*s.GetAssistantPresence() {
		attemptJoin := func() error {
			if s.GetInviteLink() == "" {
				invLink, err := Bot.GetChatInviteLink(chatID, &telegram.InviteLinkOptions{RequestNeeded: false})
				if err != nil {
					switch {
					case telegram.MatchError(err, "CHAT_ID_INVALID"),
						telegram.MatchError(err, "CHAT_ADMIN_REQUIRED"),
						telegram.MatchError(err, "CHANNEL_PRIVATE"),
						telegram.MatchError(err, "CHANNEL_INVALID"):
						return ErrAdminPermissionRequired
					default:
						return fmt.Errorf("%w: %v", ErrAssistantInviteLinkFetch, err)
					}
				}
				if l, ok := invLink.(*telegram.ChatInviteExported); ok && l.Link != "" {
					s.SetInviteLink(l.Link)
				} else {
					return fmt.Errorf("%w: no valid invite link retrieved", ErrAssistantInviteLinkFetch)
				}
			}

			_, err := UBot.JoinChannel(s.GetInviteLink())
			if err == nil || telegram.MatchError(err, "USER_ALREADY_PARTICIPANT") {
				s.SetAssistantPresence(boolToPtr(true))
				s.SetAssistantBanned(boolToPtr(false))
				return nil
			}
			return err
		}

		err := attemptJoin()
		if telegram.MatchError(err, "INVITE_HASH_EXPIRED") {
			s.SetInviteLink("")
			err = attemptJoin()
			if err != nil {
				return false, fmt.Errorf("assistant join failed after refreshing invite: %v", err)
			}
		} else if err != nil {
			if telegram.MatchError(err, "INVITE_REQUEST_SENT") {
				iChat, errChat := Bot.ResolvePeer(chatID)
				if errChat != nil {
					return false, fmt.Errorf("%w: %v", ErrPeerResolveFailed, errChat)
				}
				iUser, errUser := Bot.ResolvePeer(UbUser.ID)
				if errUser != nil {
					return false, fmt.Errorf("%w: %v", ErrPeerResolveFailed, errUser)
				}
				var pUser *telegram.InputUserObj
				if iu, ok := iUser.(*telegram.InputPeerUser); ok {
					pUser = &telegram.InputUserObj{UserID: iu.UserID, AccessHash: iu.AccessHash}
				} else {
					return false, fmt.Errorf("%w: failed to cast user to InputPeerUser", ErrPeerResolveFailed)
				}
				_, acceptErr := Bot.MessagesHideChatJoinRequest(true, iChat, pUser)
				if acceptErr == nil || telegram.MatchError(acceptErr, "USER_ALREADY_PARTICIPANT") {
					s.SetAssistantPresence(boolToPtr(true))
					s.SetAssistantBanned(boolToPtr(false))
					return true, nil
				} else if telegram.MatchError(acceptErr, "CHAT_ADMIN_REQUIRED") || telegram.MatchError(acceptErr, "CHANNEL_PRIVATE") {
					return false, ErrAdminPermissionRequired
				}
				return false, ErrAssistantJoinRequestSent
			}

			switch {
			case telegram.MatchError(err, "CHANNEL_PRIVATE"), telegram.MatchError(err, "CHANNEL_INVALID"):
				return false, ErrAssistantJoinRejected
			case telegram.MatchError(err, "CHANNELS_TOO_MUCH"), telegram.MatchError(err, "USER_CHANNELS_TOO_MUCH"):
				return false, ErrAssistantJoinRateLimited
			default:
				return false, fmt.Errorf("%w: %v", ErrAssistantInviteFailed, err)
			}
		}
	}

	return true, nil
}
