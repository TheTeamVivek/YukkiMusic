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

	"main/internal/utils"
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

var (
	chMutex    = &sync.Mutex{}
	ChatStates = make(map[int64]*ChatState)
)

func boolToPtr(b bool) *bool {
	return &b
}

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

// GetVoiceChatStatus returns whether a voice chat is active in the given chat.
func GetVoiceChatStatus(chatID int64, force ...bool) (bool, error) {
	s, _ := GetChatState(chatID, true)
	needFetch := len(force) > 0 && force[0] || s.GetVoiceChatStatus() == nil

	if needFetch {
		if err := updateVoiceChatStatus(s, chatID); err != nil {
			return false, err
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
	needFetch := len(force) > 0 && force[0] ||
		s.GetAssistantPresence() == nil ||
		s.GetAssistantBanned() == nil

	if needFetch {
		if err := updateAssistantStatus(s, chatID); err != nil {
			return false, err
		}
	}

	if *s.GetAssistantBanned() {
		return false, ErrAssistantBanned
	}

	if *s.GetAssistantPresence() {
		return true, nil
	}

	if err := ensureAssistantJoined(s, chatID); err != nil {
		return false, err
	}

	return true, nil
}

func updateVoiceChatStatus(s *ChatState, chatID int64) error {
	fullChat, err := fetchFullChat(chatID)
	if err != nil {
		return err
	}

	s.SetVoiceChatStatus(boolToPtr(fullChat.Call != nil))
	setInviteLinkIfNeeded(s, fullChat)

	return nil
}

// --- helpers ---

func fetchFullChat(chatID int64) (*telegram.ChannelFull, error) {
	fullChat, err := utils.GetFullChannel(Bot, chatID)
	if err != nil {
		switch {
		case telegram.MatchError(err, "CHANNEL_INVALID"),
			telegram.MatchError(err, "CHANNEL_PRIVATE"):
			return nil, ErrAdminPermissionRequired
		default:
			return nil, fmt.Errorf("%w: %v", ErrFetchFailed, err)
		}
	}
	return fullChat, nil
}

func setInviteLinkIfNeeded(s *ChatState, chat *telegram.ChannelFull) {
	if chat.ExportedInvite == nil || s.GetInviteLink() != "" {
		return
	}
	if l, ok := chat.ExportedInvite.(*telegram.ChatInviteExported); ok && l.Link != "" {
		s.SetInviteLink(l.Link)
	}
}

func updateAssistantStatus(s *ChatState, chatID int64) error {
	member, err := Bot.GetChatMember(chatID, UbUser.ID)
	if err != nil {
		return handleMemberFetchError(s, err)
	}

	if member == nil {
		s.SetAssistantPresence(boolToPtr(false))
		s.SetAssistantBanned(boolToPtr(false))
		return nil
	}

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
	return nil
}

func handleMemberFetchError(s *ChatState, err error) error {
	switch {
	case telegram.MatchError(err, "USER_NOT_PARTICIPANT"),
		telegram.MatchError(err, "PARTICIPANT_ID_INVALID"):
		s.SetAssistantPresence(boolToPtr(false))
		s.SetAssistantBanned(boolToPtr(false))
		return nil

	case telegram.MatchError(err, "CHAT_ADMIN_REQUIRED"),
		telegram.MatchError(err, "CHANNEL_PRIVATE"):
		return ErrAdminPermissionRequired

	default:
		return fmt.Errorf("%w: %v", ErrFetchFailed, err)
	}
}

func ensureAssistantJoined(s *ChatState, chatID int64) error {
	tryJoin := func() error {
		if s.GetInviteLink() == "" {
			if err := fetchAndSetInviteLink(s, chatID); err != nil {
				return err
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

	err := tryJoin()
	if telegram.MatchError(err, "INVITE_HASH_EXPIRED") {
		s.SetInviteLink("")
		if retryErr := tryJoin(); retryErr != nil {
			return fmt.Errorf("assistant join failed after refreshing invite: %v", retryErr)
		}
		return nil
	}

	if err != nil {
		return handleJoinError(err, chatID, s)
	}
	return nil
}

func fetchAndSetInviteLink(s *ChatState, chatID int64) error {
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
		return nil
	}
	return fmt.Errorf("%w: no valid invite link retrieved", ErrAssistantInviteLinkFetch)
}

func handleJoinError(err error, chatID int64, s *ChatState) error {
	switch {
	case telegram.MatchError(err, "INVITE_REQUEST_SENT"):
		return handleJoinRequestPending(chatID, s)

	case telegram.MatchError(err, "CHANNEL_PRIVATE"),
		telegram.MatchError(err, "CHANNEL_INVALID"):
		return ErrAssistantJoinRejected

	case telegram.MatchError(err, "CHANNELS_TOO_MUCH"),
		telegram.MatchError(err, "USER_CHANNELS_TOO_MUCH"):
		return ErrAssistantJoinRateLimited

	default:
		return fmt.Errorf("%w: %v", ErrAssistantInviteFailed, err)
	}
}

func handleJoinRequestPending(chatID int64, s *ChatState) error {
	iChat, errChat := Bot.ResolvePeer(chatID)
	if errChat != nil {
		return fmt.Errorf("%w: %v", ErrPeerResolveFailed, errChat)
	}

	iUser, errUser := Bot.ResolvePeer(UbUser.ID)
	if errUser != nil {
		return fmt.Errorf("%w: %v", ErrPeerResolveFailed, errUser)
	}

	iu, ok := iUser.(*telegram.InputPeerUser)
	if !ok {
		return fmt.Errorf("%w: failed to cast user to InputPeerUser", ErrPeerResolveFailed)
	}

	pUser := &telegram.InputUserObj{UserID: iu.UserID, AccessHash: iu.AccessHash}
	_, acceptErr := Bot.MessagesHideChatJoinRequest(true, iChat, pUser)
	if acceptErr == nil || telegram.MatchError(acceptErr, "USER_ALREADY_PARTICIPANT") {
		s.SetAssistantPresence(boolToPtr(true))
		s.SetAssistantBanned(boolToPtr(false))
		return nil
	}

	if telegram.MatchError(acceptErr, "CHAT_ADMIN_REQUIRED") || telegram.MatchError(acceptErr, "CHANNEL_PRIVATE") {
		return ErrAdminPermissionRequired
	}
	return ErrAssistantJoinRequestSent
}
