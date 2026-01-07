/*
  - This file is part of YukkiMusic.
    *

  - YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
  - Copyright (C) 2025 TheTeamVivek
    *
  - This program is free software: you can redistribute it and/or modify
  - it under the terms of the GNU General Public License as published by
  - the Free Software Foundation, either version 3 of the License, or
  - (at your option) any later version.
    *
  - This program is distributed in the hope that it will be useful,
  - but WITHOUT ANY WARRANTY; without even the implied warranty of
  - MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
  - GNU General Public License for more details.
    *
  - You should have received a copy of the GNU General Public License
  - along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package core

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/utils"
)

var (
	ErrAdminPermissionRequired  = errors.New("admin permission required")
	ErrFetchFailed              = errors.New("failed to fetch chat info")
	ErrAssistantGetFailed       = errors.New("failed to get assistant")
	ErrAssistantInviteLinkFetch = errors.New("failed to fetch invite link")
	ErrAssistantJoinRejected    = errors.New("invite link invalid or expired")
	ErrAssistantJoinRateLimited = errors.New("rate limited")
	ErrAssistantJoinRequestSent = errors.New("join request sent")
	ErrAssistantInviteFailed    = errors.New("failed to join chat")
	ErrPeerResolveFailed        = errors.New("failed to resolve peer")
)

// =============================================================================
// CHAT STATE
// =============================================================================

type ChatState struct {
	mu         *sync.RWMutex
	ChatID     int64
	Assistant  *Assistant

	isPresent  *bool
	isBanned   *bool
	isVCActive *bool
	inviteLink string
}

var (
	stateMutex = &sync.Mutex{}
	states     = make(map[int64]*ChatState)
)

// =============================================================================
// STATE MANAGEMENT
// =============================================================================

func GetChatState(chatID int64) (*ChatState, error) {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	s, exists := states[chatID]
	if !exists {
		s = &ChatState{
			mu:     &sync.RWMutex{},
			ChatID: chatID,
		}
		states[chatID] = s
	}

	if _, err := s.ensureAssistant(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrAssistantGetFailed, err)
	}

	return s, nil
}

func DeleteChatState(chatID int64) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	delete(states, chatID)
}

// =============================================================================
// STATE SETTERS
// =============================================================================

func (s *ChatState) SetAssistantPresent(v bool) {
	s.mu.Lock()
	s.isPresent = &v
	s.mu.Unlock()
}

func (s *ChatState) SetAssistantBanned(v bool) {
	s.mu.Lock()
	s.isBanned = &v
	s.mu.Unlock()
}

func (s *ChatState) SetVoiceChatActive(v bool) {
	s.mu.Lock()
	s.isVCActive = &v
	s.mu.Unlock()
}

func (s *ChatState) SetInviteLink(link string) {
	s.mu.Lock()
	s.inviteLink = link
	s.mu.Unlock()
}

// =============================================================================
// STATE GETTERS
// =============================================================================

func (s *ChatState) GetAssistantPresence() *bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isPresent
}

func (s *ChatState) GetAssistantBanned() *bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isBanned
}

func (s *ChatState) IsStateUnknown() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isPresent == nil || s.isBanned == nil
}

// =============================================================================
// STATE QUERIES
// =============================================================================

func (s *ChatState) IsAssistantPresent(force ...bool) (bool, error) {
	if s.shouldRefresh(s.isPresent, force) {
		if err := s.RefreshAssistantState(); err != nil {
			return false, err
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isPresent != nil && *s.isPresent, nil
}

func (s *ChatState) IsAssistantBanned(force ...bool) (bool, error) {
	if s.shouldRefresh(s.isBanned, force) {
		if err := s.RefreshAssistantState(); err != nil {
			return false, err
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isBanned != nil && *s.isBanned, nil
}

func (s *ChatState) IsActiveVC(force ...bool) (bool, error) {
	if s.shouldRefresh(s.isVCActive, force) {
		if err := s.refreshVCState(); err != nil {
			return false, err
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isVCActive != nil && *s.isVCActive, nil
}

func (s *ChatState) shouldRefresh(flag *bool, force []bool) bool {
	if flag == nil {
		return true
	}
	return len(force) > 0 && force[0]
}

// =============================================================================
// STATE REFRESH
// =============================================================================

func (s *ChatState) RefreshAssistantState() error {
	member, err := Bot.GetChatMember(s.ChatID, s.Assistant.User.ID)
	if err != nil {
		return s.handleMemberFetchError(err)
	}

	s.applyMemberStatus(member)
	return nil
}

func (s *ChatState) refreshVCState() error {
	full, err := s.fetchFullChat()
	if err != nil {
		return err
	}

	s.SetVoiceChatActive(full.Call != nil)

	if full.Call != nil && full.ExportedInvite != nil {
		if inv, ok := full.ExportedInvite.(*telegram.ChatInviteExported); ok {
			s.SetInviteLink(inv.Link)
		}
	}

	return nil
}

func (s *ChatState) applyMemberStatus(m *telegram.Participant) {
	if m == nil {
		s.SetAssistantPresent(false)
		s.SetAssistantBanned(false)
		return
	}

	if m.Status == telegram.Restricted {
		// Status is "restricted" - could be muted OR banned
		// Check the actual participant type
		if banned, ok := m.Participant.(*telegram.ChannelParticipantBanned); ok {
			// Check ViewMessages to distinguish ban from mute
			if banned.BannedRights.ViewMessages {
				// ViewMessages=true means truly banned (can't access)
				s.SetAssistantPresent(false)
				s.SetAssistantBanned(true)
			} else {
				// ViewMessages=false means muted (can access but restricted)
				s.SetAssistantPresent(true)
				s.SetAssistantBanned(false)
			}
		} else {
			// Shouldn't happen, but default to accessible
			s.SetAssistantPresent(true)
			s.SetAssistantBanned(false)
		}
		return
	}

	// Handle other statuses
	statusMap := map[string]struct{ present, banned bool }{
		telegram.Left:    {false, false},
		telegram.Admin:   {true, false},
		telegram.Member:  {true, false},
		telegram.Creator: {true, false},
	}

	if state, ok := statusMap[m.Status]; ok {
		s.SetAssistantPresent(state.present)
		s.SetAssistantBanned(state.banned)
	}
}

func (s *ChatState) handleMemberFetchError(err error) error {
	if telegram.MatchError(err, "USER_NOT_PARTICIPANT") ||
		telegram.MatchError(err, "PARTICIPANT_ID_INVALID") {
		s.SetAssistantPresent(false)
		s.SetAssistantBanned(false)
		return nil
	}

	if telegram.MatchError(err, "CHAT_ADMIN_REQUIRED") ||
		telegram.MatchError(err, "CHANNEL_PRIVATE") {
		return ErrAdminPermissionRequired
	}

	gologging.Error("Member fetch error: " + err.Error())
	return fmt.Errorf("%w: %v", ErrFetchFailed, err)
}

// =============================================================================
// ASSISTANT JOIN
// =============================================================================

func (s *ChatState) TryJoin() error {
	err := s.attemptJoin()

	if telegram.MatchError(err, "INVITE_HASH_EXPIRED") {
		s.SetInviteLink("")
		return s.attemptJoin()
	}

	return err
}

func (s *ChatState) attemptJoin() error {
	link, err := s.getOrFetchInviteLink()
	if err != nil {
		return err
	}

	_, err = s.Assistant.Client.JoinChannel(link)
	if err == nil || telegram.MatchError(err, "USER_ALREADY_PARTICIPANT") {
		s.setJoinSuccess()
		return nil
	}

	return s.handleJoinError(err)
}

func (s *ChatState) getOrFetchInviteLink() (string, error) {
	s.mu.RLock()
	link := s.inviteLink
	s.mu.RUnlock()

	if link != "" {
		return link, nil
	}

	if err := s.fetchInviteLink(); err != nil {
		return "", err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.inviteLink, nil
}

func (s *ChatState) setJoinSuccess() {
	s.SetAssistantPresent(true)
	s.SetAssistantBanned(false)
}

func (s *ChatState) handleJoinError(err error) error {
	errorMap := map[string]error{
		"CHANNEL_PRIVATE":        ErrAssistantJoinRejected,
		"CHANNEL_INVALID":        ErrAssistantJoinRejected,
		"CHANNELS_TOO_MUCH":      ErrAssistantJoinRateLimited,
		"USER_CHANNELS_TOO_MUCH": ErrAssistantJoinRateLimited,
	}

	for pattern, mappedErr := range errorMap {
		if telegram.MatchError(err, pattern) {
			return mappedErr
		}
	}

	if telegram.MatchError(err, "INVITE_REQUEST_SENT") {
		return s.approveJoinRequest()
	}

	return fmt.Errorf("%w: %v", ErrAssistantInviteFailed, err)
}

func (s *ChatState) approveJoinRequest() error {
	chatPeer, err := Bot.ResolvePeer(s.ChatID)
	if err != nil {
		return err
	}

	userPeer, err := Bot.ResolvePeer(s.Assistant.User.ID)
	if err != nil {
		return err
	}

	inputUser, ok := userPeer.(*telegram.InputPeerUser)
	if !ok {
		return errors.New("invalid user peer")
	}

	user := &telegram.InputUserObj{
		UserID:     inputUser.UserID,
		AccessHash: inputUser.AccessHash,
	}

	_, err = Bot.MessagesHideChatJoinRequest(true, chatPeer, user)

	if err == nil || telegram.MatchError(err, "USER_ALREADY_PARTICIPANT") {
		s.setJoinSuccess()
		return nil
	}

	if telegram.MatchError(err, "CHAT_ADMIN_REQUIRED") ||
		telegram.MatchError(err, "CHANNEL_PRIVATE") {
		return ErrAdminPermissionRequired
	}

	return ErrAssistantJoinRequestSent
}

// =============================================================================
// HELPERS
// =============================================================================

func (s *ChatState) ensureAssistant() (*Assistant, error) {
	s.mu.RLock()
	ass := s.Assistant
	s.mu.RUnlock()

	if ass != nil {
		return ass, nil
	}

	if Assistants == nil || Assistants.Count() == 0 {
		return nil, errors.New("no assistants available")
	}

	ass, err := Assistants.ForChat(s.ChatID)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.Assistant = ass
	s.mu.Unlock()

	return ass, nil
}

func (s *ChatState) fetchInviteLink() error {
	inv, err := Bot.GetChatInviteLink(s.ChatID,
		&telegram.InviteLinkOptions{RequestNeeded: false})

	if err != nil {
		if s.isAdminError(err) {
			return ErrAdminPermissionRequired
		}
		return fmt.Errorf("%w: %v", ErrAssistantInviteLinkFetch, err)
	}

	if link, ok := inv.(*telegram.ChatInviteExported); ok && link.Link != "" {
		s.SetInviteLink(link.Link)
		return nil
	}

	return ErrAssistantInviteLinkFetch
}

func (s *ChatState) fetchFullChat() (*telegram.ChannelFull, error) {
	full, err := utils.GetFullChannel(Bot, s.ChatID)
	if err != nil {
		if s.isAdminError(err) {
			return nil, ErrAdminPermissionRequired
		}
		return nil, fmt.Errorf("%w: %v", ErrFetchFailed, err)
	}
	return full, nil
}

func (s *ChatState) isAdminError(err error) bool {
	return telegram.MatchError(err, "CHAT_ID_INVALID") ||
		telegram.MatchError(err, "CHAT_ADMIN_REQUIRED") ||
		telegram.MatchError(err, "CHANNEL_PRIVATE") ||
		telegram.MatchError(err, "CHANNEL_INVALID")
}
