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
	"strings"
	"sync"

	"github.com/amarnathcjd/gogram/telegram"
        "github.com/Laky-64/gologging"

	"main/config"
	"main/internal/utils"
)

type ChatState struct {
	mu               *sync.RWMutex
	ChatID           int64
	AssistantPresent *bool
	AssistantBanned  *bool
	VoiceChatActive  *bool
	InviteLink       string
}

var (
	ErrAdminPermissionRequired  = errors.New("admin permission required")
	ErrFetchFailed              = errors.New("failed to fetch chat info")
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

func GetChatState(chatID int64) *ChatState {
	chMutex.Lock()
	defer chMutex.Unlock()

	state, ok := ChatStates[chatID]

	if !ok {
		state = &ChatState{
			mu:     &sync.RWMutex{},
			ChatID: chatID,
		}
		ChatStates[chatID] = state
	}
	return state
}

func DeleteChatState(chatID int64) {
	chMutex.Lock()
	defer chMutex.Unlock()
	delete(ChatStates, chatID)
}

func (cs *ChatState) SetAssistantPresent(v bool) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.AssistantPresent = &v
}

func (cs *ChatState) SetAssistantBanned(v bool) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.AssistantBanned = &v
}

func (cs *ChatState) SetVoiceChatActive(v bool) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.VoiceChatActive = &v
}

func (cs *ChatState) SetInviteLink(link string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.InviteLink = link
}

func (cs *ChatState) IsActiveVC(force ...bool) (bool, error) {
	isForce := len(force) > 0 && force[0]

	if err := cs.ensureVoiceState(isForce); err != nil {
		return false, err
	}

	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.VoiceChatActive != nil && *cs.VoiceChatActive, nil
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

func (cs *ChatState) IsAssistantBanned(force ...bool) (bool, error) {
	isForce := len(force) > 0 && force[0]

	if err := cs.ensureAssistantState(isForce); err != nil {
		return false, err
	}

	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.AssistantBanned != nil && *cs.AssistantBanned, nil
}

func (cs *ChatState) IsAssistantPresent(force ...bool) (bool, error) {
	isForce := len(force) > 0 && force[0]

	if err := cs.ensureAssistantState(isForce); err != nil {
		return false, err
	}

	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.AssistantPresent != nil && *cs.AssistantPresent, nil
}

func (cs *ChatState) TryJoin() error {
	tryJoin := func() error {
		cs.mu.RLock()
		link := cs.InviteLink
		cs.mu.RUnlock()
		if link == "" {
			if err := fetchAndSetInviteLink(cs, cs.ChatID); err != nil {
				return err
			}
		}
		cs.mu.RLock()
		link = cs.InviteLink
		cs.mu.RUnlock()
		_, err := UBot.JoinChannel(link)
		if err == nil || telegram.MatchError(err, "USER_ALREADY_PARTICIPANT") {
			cs.SetAssistantPresent(true)
			cs.SetAssistantBanned(false)
			return nil
		}
		return err
	}

	err := tryJoin()
	if telegram.MatchError(err, "INVITE_HASH_EXPIRED") {
		cs.SetInviteLink("")
		if retryErr := tryJoin(); retryErr != nil {
			return fmt.Errorf("assistant join failed after refreshing invite: %v", retryErr)
		}
		return nil
	}

	if err != nil {
		return handleJoinError(err, cs.ChatID, cs)
	}
	return nil
}

// --- helpers ---

func (cs *ChatState) ensureVoiceState(force bool) error {
	cs.mu.RLock()
	need := cs.VoiceChatActive == nil || force
	cs.mu.RUnlock()

	if !need {
		return nil
	}

	fullChat, err := fetchFullChat(cs.ChatID)
	if err != nil {
		return err
	}

	isVCActive := fullChat.Call != nil
	cs.SetVoiceChatActive(isVCActive)

	if isVCActive && fullChat.ExportedInvite != nil {
		if l, ok := fullChat.ExportedInvite.(*telegram.ChatInviteExported); ok && l.Link != "" {
			cs.SetInviteLink(l.Link)
		}
	}

	return nil
}

func (cs *ChatState) ensureAssistantState(force bool) error {
	cs.mu.RLock()
	need := cs.AssistantPresent == nil || cs.AssistantBanned == nil || force
	cs.mu.RUnlock()

	if !need {
		return nil
	}

	member, err := Bot.GetChatMember(cs.ChatID, UbUser.ID)

	if err != nil {
gologging.Error("raw error of GetChatMember in core.ChatState"+err.Error())
		if strings.Contains(err.Error(), "there is no peer with id") {

			triggerAssistantStart()
				member, err = Bot.GetChatMember(cs.ChatID, UbUser.ID)
				if err != nil {
					return handleMemberFetchError(cs, err)
				}
				applyMemberStatus(cs, member)
				return nil
			
		}
		return handleMemberFetchError(cs, err)
	}

	applyMemberStatus(cs, member)
	return nil
}

func triggerAssistantStart(){

	_, sendErr := UBot.SendMessage(BUser.Username, "/start")
	if sendErr == nil {
		return
	}

	if config.LoggerID != 0 {
		UBot.SendMessage(config.LoggerID,
			"⚠️ Unable to get assistant state. Please start the bot with assistant id")
	}

	if config.OwnerID != 0 {
		UBot.SendMessage(config.OwnerID,
			"⚠️ Unable to get assistant state. Please start the bot with assistant id")
	}

	return
}

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

func applyMemberStatus(s *ChatState, member *telegram.Participant) {
	if member == nil {
		s.SetAssistantPresent(false)
		s.SetAssistantBanned(false)

		return
	}
	var p, b bool

	switch member.Status {
	case telegram.Kicked, telegram.Restricted:
		p = false
		b = true
	case telegram.Left:
		p = false
		b = false

	case telegram.Admin, telegram.Member:
		p = true
		b = false
	}

	s.SetAssistantPresent(p)
	s.SetAssistantBanned(b)
	return
}

func handleMemberFetchError(s *ChatState, err error) error {
	switch {
	case telegram.MatchError(err, "USER_NOT_PARTICIPANT"),
		telegram.MatchError(err, "PARTICIPANT_ID_INVALID"):
		s.SetAssistantPresent(false)
		s.SetAssistantBanned(false)
		return nil

	case telegram.MatchError(err, "CHAT_ADMIN_REQUIRED"),
		telegram.MatchError(err, "CHANNEL_PRIVATE"):
		return ErrAdminPermissionRequired

	default:
		return fmt.Errorf("%w: %v", ErrFetchFailed, err)
	}
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
		s.SetAssistantPresent(true)
		s.SetAssistantBanned(false)
		return nil
	}

	if telegram.MatchError(acceptErr, "CHAT_ADMIN_REQUIRED") || telegram.MatchError(acceptErr, "CHANNEL_PRIVATE") {
		return ErrAdminPermissionRequired
	}
	return ErrAssistantJoinRequestSent
}
