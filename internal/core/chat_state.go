/*
 * ● YukkiMusic
 * ○ A high-performance engine for streaming music in Telegram voicechats.
 *
 * Copyright (C) 2026 TheTeamVivek
 *
 * This program is free software: you can redistribute it and/or modify it under the
 * terms of the GNU General Public License as published by the Free Software Foundation,
 * either version 3 of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
 * PARTICULAR PURPOSE. See the GNU General Public License for more details.
 *
 * Repository: https://github.com/TheTeamVivek/YukkiMusic
 */

package core

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/utils"
)

var (
	ErrAdminPermissionRequired  = errors.New("admin permission required")
	ErrStateFetchFailed         = errors.New("state fetch failed")
	ErrAssistantNotAvailable    = errors.New("assistant unavailable")
	ErrAssistantInviteLinkFetch = errors.New("failed to fetch invite link")
	ErrInviteRequestSent        = errors.New("join request sent")
	ErrJoinFailed               = errors.New("assistant join failed")
)

type StateSnapshot struct {
	AssistantPresent bool
	AssistantBanned  bool
	VoiceChatActive  bool
}

type ChatState struct {
	mu sync.RWMutex

	ChatID int64

	Assistant *Assistant

	inviteLink string
	snapshot   StateSnapshot
	fetched    bool
}

var (
	chatStatesMu sync.Mutex
	chatStates   = map[int64]*ChatState{}
)

func GetChatState(chatID int64) (*ChatState, error) {
	chatStatesMu.Lock()
	state := chatStates[chatID]
	if state == nil {
		state = &ChatState{ChatID: chatID}
		chatStates[chatID] = state
	}
	chatStatesMu.Unlock()
	if err := state.ensureAssistant(); err != nil {
		gologging.ErrorF("chat_state: ensureAssistant failed for %d: %v", chatID, err)
		return nil, err
	}
	return state, nil
}

func DeleteChatState(chatID int64) {
	chatStatesMu.Lock()
	delete(chatStates, chatID)
	chatStatesMu.Unlock()
}

func (s *ChatState) Snapshot(force bool) (StateSnapshot, error) {
	s.mu.RLock()
	cached := s.snapshot
	fetched := s.fetched
	s.mu.RUnlock()
	if fetched && !force {
		return cached, nil
	}
	if err := s.refresh(); err != nil {
		gologging.ErrorF("chat_state: refresh failed for %d: %v", s.ChatID, err)
		return StateSnapshot{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshot, nil
}

func (s *ChatState) EnsureAssistantJoined(username string) error {
	if username != "" {
		if err := s.joinBy(username); err == nil {
			return nil
		}
	}

	err := s.joinByInviteLink()
	if err == nil {
		return nil
	}
	if telegram.MatchError(err, "INVITE_HASH_EXPIRED") {
		gologging.ErrorF("chat_state: invite expired for %d, retrying with refreshed link", s.ChatID)
		s.setInviteLink("")
		return s.joinByInviteLink()
	}
	return err
}

func (s *ChatState) SetAssistantPresent(v bool) {
	s.mu.Lock()
	s.snapshot.AssistantPresent = v
	s.fetched = true
	s.mu.Unlock()
}

func (s *ChatState) SetAssistantBanned(v bool) {
	s.mu.Lock()
	s.snapshot.AssistantBanned = v
	s.fetched = true
	s.mu.Unlock()
}

func (s *ChatState) SetVoiceChatActive(v bool) {
	s.mu.Lock()
	s.snapshot.VoiceChatActive = v
	s.fetched = true
	s.mu.Unlock()
}

func (s *ChatState) AssistantFetched() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.fetched
}

func (s *ChatState) refresh() error {
	gologging.DebugF("chat_state: refresh(chat=%d)", s.ChatID)
	if s.Assistant == nil {
		if err := s.ensureAssistant(); err != nil {
			return err
		}
	}
	full, err := utils.GetFullChannel(Bot, s.ChatID)
	if err != nil {
		gologging.ErrorF("chat_state: GetFullChannel failed for %d: %v", s.ChatID, err)
		if isAdminError(err) {
			return ErrAdminPermissionRequired
		}
		return fmt.Errorf("%w: %v", ErrStateFetchFailed, err)
	}

	member, err := Bot.GetChatMember(s.ChatID, s.Assistant.Self.ID)
	if err != nil {
		if telegram.MatchError(err, "USER_NOT_PARTICIPANT") || telegram.MatchError(err, "PARTICIPANT_ID_INVALID") {
			s.applySnapshot(false, false, full.Call != nil)
			gologging.DebugF("chat_state: assistant not participant in %d", s.ChatID)
			return nil
		}
		if isAdminError(err) {
			gologging.ErrorF("chat_state: admin permission required for GetChatMember in %d", s.ChatID)
			return ErrAdminPermissionRequired
		}
		return fmt.Errorf("%w: %v", ErrStateFetchFailed, err)
	}

	present, banned := membership(member)
	s.applySnapshot(present, banned, full.Call != nil)
	if full.ExportedInvite != nil {
		if inv, ok := full.ExportedInvite.(*telegram.ChatInviteExported); ok && inv.Link != "" {
			s.setInviteLink(inv.Link)
		}
	}
	return nil
}

func membership(m *telegram.Participant) (bool, bool) {
	if m == nil {
		return false, false
	}
	if m.Status == telegram.Restricted {
		if b, ok := m.Participant.(*telegram.ChannelParticipantBanned); ok && b.BannedRights.ViewMessages {
			return false, true
		}
		return true, false
	}
	switch m.Status {
	case telegram.Member, telegram.Admin, telegram.Creator:
		return true, false
	}
	return false, false
}

func (s *ChatState) joinBy(username string) error {
	gologging.DebugF("chat_state: joinBy username=%s chat=%d", username, s.ChatID)
	_, err := s.Assistant.Client.JoinChannel(username)
	if err == nil || telegram.MatchError(err, "USER_ALREADY_PARTICIPANT") {
		s.applySnapshot(true, false, s.snapshot.VoiceChatActive)
		return nil
	}
	return err
}

func (s *ChatState) joinByInviteLink() error {
	gologging.DebugF("chat_state: joinByInviteLink(chat=%d)", s.ChatID)
	link, err := s.resolveInviteLink()
	if err != nil {
		return err
	}
	_, err = s.Assistant.Client.JoinChannel(link)
	if err == nil || telegram.MatchError(err, "USER_ALREADY_PARTICIPANT") {
		s.applySnapshot(true, false, s.snapshot.VoiceChatActive)
		return nil
	}
	if telegram.MatchError(err, "INVITE_REQUEST_SENT") {
		gologging.InfoF("chat_state: invite request sent for %d, attempting approval", s.ChatID)
		if err := s.approveJoinRequest(); err != nil {
			return ErrInviteRequestSent
		}
		return nil
	}
	if telegram.MatchError(err, "USER_CHANNELS_TOO_MUCH") || telegram.MatchError(err, "CHANNELS_TOO_MUCH") {
		gologging.InfoF("chat_state: join limit reached for %d, leaving inactive chats", s.ChatID)
		s.leaveInactiveAssistantChats(5)
		time.Sleep(1 * time.Second)
		_, retryErr := s.Assistant.Client.JoinChannel(link)
		if retryErr == nil || telegram.MatchError(retryErr, "USER_ALREADY_PARTICIPANT") {
			s.applySnapshot(true, false, s.snapshot.VoiceChatActive)
			return nil
		}
		return retryErr
	}
	if isAdminError(err) {
		return ErrAdminPermissionRequired
	}
	return fmt.Errorf("%w: %v", ErrJoinFailed, err)
}

func (s *ChatState) resolveInviteLink() (string, error) {
	gologging.DebugF("chat_state: resolveInviteLink(chat=%d)", s.ChatID)
	s.mu.RLock()
	cached := s.inviteLink
	s.mu.RUnlock()
	if cached != "" {
		return cached, nil
	}
	inv, err := Bot.GetChatInviteLink(s.ChatID, &telegram.InviteLinkOptions{RequestNeeded: false})
	if err != nil {
		if isAdminError(err) {
			return "", ErrAdminPermissionRequired
		}
		return "", fmt.Errorf("%w: %v", ErrAssistantInviteLinkFetch, err)
	}
	link, ok := inv.(*telegram.ChatInviteExported)
	if !ok || link.Link == "" {
		return "", ErrAssistantInviteLinkFetch
	}
	s.setInviteLink(link.Link)
	return link.Link, nil
}

func (s *ChatState) approveJoinRequest() error {
	gologging.DebugF("chat_state: approveJoinRequest(chat=%d)", s.ChatID)
	chatPeer, err := Bot.ResolvePeer(s.ChatID)
	if err != nil {
		return err
	}
	userPeer, err := Bot.ResolvePeer(s.Assistant.Self.ID)
	if err != nil {
		return err
	}
	u, ok := userPeer.(*telegram.InputPeerUser)
	if !ok {
		return ErrJoinFailed
	}
	_, err = Bot.MessagesHideChatJoinRequest(
		true,
		chatPeer,
		&telegram.InputUserObj{UserID: u.UserID, AccessHash: u.AccessHash},
	)
	if err == nil || telegram.MatchError(err, "USER_ALREADY_PARTICIPANT") {
		s.applySnapshot(true, false, s.snapshot.VoiceChatActive)
		return nil
	}
	if isAdminError(err) {
		return ErrAdminPermissionRequired
	}
	return err
}

func (s *ChatState) ensureAssistant() error {
	s.mu.RLock()
	has := s.Assistant != nil
	s.mu.RUnlock()
	if has {
		return nil
	}
	if Assistants == nil || Assistants.Count() == 0 {
		gologging.ErrorF("chat_state: no assistants available for %d", s.ChatID)
		return ErrAssistantNotAvailable
	}
	ass, err := Assistants.ForChat(s.ChatID)
	if err != nil {
		gologging.ErrorF("chat_state: Assistants.ForChat failed for %d: %v", s.ChatID, err)
		return fmt.Errorf("%w: %v", ErrAssistantNotAvailable, err)
	}
	gologging.InfoF("chat_state: assistant assigned for %d", s.ChatID)
	s.mu.Lock()
	s.Assistant = ass
	s.mu.Unlock()
	return nil
}

func (s *ChatState) setInviteLink(link string) { s.mu.Lock(); s.inviteLink = link; s.mu.Unlock() }
func (s *ChatState) applySnapshot(p, b, v bool) {
	s.mu.Lock()
	s.snapshot = StateSnapshot{AssistantPresent: p, AssistantBanned: b, VoiceChatActive: v}
	s.fetched = true
	s.mu.Unlock()
}

func (s *ChatState) leaveInactiveAssistantChats(limit int) {
	gologging.DebugF("chat_state: leaveInactiveAssistantChats(chat=%d, limit=%d)", s.ChatID, limit)
	if s.Assistant == nil || s.Assistant.Client == nil || limit <= 0 {
		return
	}

	activeRooms := GetAllRooms()
	leftCount := 0
	err := s.Assistant.Client.IterDialogs(func(d *telegram.TLDialog) error {
		if d == nil || d.IsUser() {
			return nil
		}

		chatID := d.GetChannelID()
		if chatID == 0 || chatID == s.ChatID {
			return nil
		}

		if _, active := activeRooms[chatID]; active {
			return nil
		}

		leaveErr := s.Assistant.Client.LeaveChannel(chatID)
		if leaveErr != nil {
			if wait := telegram.GetFloodWait(leaveErr); wait > 0 {
				time.Sleep(time.Duration(wait) * time.Second)
			}
			return nil
		}

		leftCount++
		if leftCount >= limit {
			return telegram.ErrStopIteration
		}
		return nil
	}, &telegram.DialogOptions{
		Limit: limit * 20,
	})

	if err != nil && err != telegram.ErrStopIteration {
		gologging.WarnF("chat_state: IterDialogs failed while auto-leaving chats: %v", err)
	}
}

func isAdminError(err error) bool {
	return telegram.MatchError(err, "CHAT_ID_INVALID") || telegram.MatchError(err, "CHAT_ADMIN_REQUIRED") ||
		telegram.MatchError(err, "CHANNEL_PRIVATE") ||
		telegram.MatchError(err, "CHANNEL_INVALID")
}
