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
	"strings"
	"sync"
	"time"

	td "github.com/AshokShau/gotdbot"
	"github.com/amarnathcjd/gogram/telegram"
	"yukkimusic/internal/logger"

	"yukkimusic/internal/utils"
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
		logger.Errorf("chat_state: ensureAssistant failed for %d: %v", chatID, err)
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
		logger.Errorf("chat_state: refresh failed for %d: %v", s.ChatID, err)
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
		logger.Errorf("chat_state: invite expired for %d, retrying with refreshed link", s.ChatID)
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
	logger.Debugf("chat_state: refresh(chat=%d)", s.ChatID)
	if s.Assistant == nil {
		if err := s.ensureAssistant(); err != nil {
			return err
		}
	}
	full, err := utils.GetFullChat(TDBot, s.ChatID)
	if err != nil {
		logger.Errorf("chat_state: GetFullChat failed for %d: %v", s.ChatID, err)
		if isAdminError(err) {
			return ErrAdminPermissionRequired
		}
		return fmt.Errorf("%w: %v", ErrStateFetchFailed, err)
	}

	voiceActive := false
	if full.Chat.VideoChat != nil && full.Chat.VideoChat.GroupCallId != 0 {
		voiceActive = true
	}

	member, err := TDBot.GetChatMember(s.ChatID, &td.MessageSenderUser{UserId: s.Assistant.Self.ID})
	if err != nil {
		if isAdminError(err) {
			logger.Errorf("chat_state: admin permission required for GetChatMember in %d", s.ChatID)
			return ErrAdminPermissionRequired
		}
		return fmt.Errorf("%w: %v", ErrStateFetchFailed, err)
	}

	present, banned := membership(member)
	s.applySnapshot(present, banned, voiceActive)
	if full.SupergroupFullInfo != nil &&
		full.SupergroupFullInfo.InviteLink != nil &&
		full.SupergroupFullInfo.InviteLink.InviteLink != "" {
		s.setInviteLink(full.SupergroupFullInfo.InviteLink.InviteLink)
	}
	return nil
}

func membership(m *td.ChatMember) (bool, bool) {
	if m == nil {
		return false, false
	}
	switch st := m.Status.(type) {
	case *td.ChatMemberStatusBanned:
		return false, true
	case *td.ChatMemberStatusRestricted:
		if !st.IsMember {
			return false, true
		}
		return true, false
	case *td.ChatMemberStatusMember, *td.ChatMemberStatusAdministrator, *td.ChatMemberStatusCreator:
		return true, false
	}
	return false, false
}

func (s *ChatState) joinBy(username string) error {
	logger.Debugf("chat_state: joinBy username=%s chat=%d", username, s.ChatID)
	_, err := s.Assistant.Client.JoinChannel(username)
	if err == nil || telegram.MatchError(err, "USER_ALREADY_PARTICIPANT") {
		s.applySnapshot(true, false, s.snapshot.VoiceChatActive)
		return nil
	}
	return err
}

func (s *ChatState) joinByInviteLink() error {
	logger.Debugf("chat_state: joinByInviteLink(chat=%d)", s.ChatID)
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
		logger.Infof("chat_state: invite request sent for %d, attempting approval", s.ChatID)
		if err := s.approveJoinRequest(); err != nil {
			return ErrInviteRequestSent
		}
		return nil
	}
	if telegram.MatchError(err, "USER_CHANNELS_TOO_MUCH") || telegram.MatchError(err, "CHANNELS_TOO_MUCH") {
		logger.Infof("chat_state: join limit reached for %d, leaving inactive chats", s.ChatID)
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
	logger.Debugf("chat_state: resolveInviteLink(chat=%d)", s.ChatID)
	s.mu.RLock()
	cached := s.inviteLink
	s.mu.RUnlock()
	if cached != "" {
		return cached, nil
	}
	full, err := utils.GetFullChat(TDBot, s.ChatID)
	if err != nil {
		if isAdminError(err) {
			return "", ErrAdminPermissionRequired
		}
		return "", fmt.Errorf("%w: %v", ErrAssistantInviteLinkFetch, err)
	}
	if full.SupergroupFullInfo != nil &&
		full.SupergroupFullInfo.InviteLink != nil &&
		full.SupergroupFullInfo.InviteLink.InviteLink != "" {
		s.setInviteLink(full.SupergroupFullInfo.InviteLink.InviteLink)
		return full.SupergroupFullInfo.InviteLink.InviteLink, nil
	}
	inv, err := TDBot.CreateChatInviteLink(s.ChatID, 0, 0, "", nil)
	if err != nil {
		if isAdminError(err) {
			return "", ErrAdminPermissionRequired
		}
		return "", fmt.Errorf("%w: %v", ErrAssistantInviteLinkFetch, err)
	}
	if inv == nil || inv.InviteLink == "" {
		return "", ErrAssistantInviteLinkFetch
	}
	s.setInviteLink(inv.InviteLink)
	return inv.InviteLink, nil
}

func (s *ChatState) approveJoinRequest() error {
	logger.Debugf("chat_state: approveJoinRequest(chat=%d)", s.ChatID)
	err := TDBot.ProcessChatJoinRequest(
		s.ChatID,
		s.Assistant.Self.ID,
		&td.ProcessChatJoinRequestOpts{Approve: true},
	)
	if err == nil {
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
		logger.Errorf("chat_state: no assistants available for %d", s.ChatID)
		return ErrAssistantNotAvailable
	}
	ass, err := Assistants.ForChat(s.ChatID)
	if err != nil {
		logger.Errorf("chat_state: Assistants.ForChat failed for %d: %v", s.ChatID, err)
		return fmt.Errorf("%w: %v", ErrAssistantNotAvailable, err)
	}
	logger.Infof("chat_state: assistant assigned for %d", s.ChatID)
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
	logger.Debugf("chat_state: leaveInactiveAssistantChats(chat=%d, limit=%d)", s.ChatID, limit)
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
		Limit: int32(limit * 20),
	})

	if err != nil && err != telegram.ErrStopIteration {
		logger.Warnf("chat_state: IterDialogs failed while auto-leaving chats: %v", err)
	}
}

func isAdminError(err error) bool {
	if telegram.MatchError(err, "CHAT_ID_INVALID") || telegram.MatchError(err, "CHAT_ADMIN_REQUIRED") ||
		telegram.MatchError(err, "CHANNEL_PRIVATE") || telegram.MatchError(err, "CHANNEL_INVALID") {
		return true
	}

	var tde *td.Error
	if errors.As(err, &tde) {
		msg := strings.ToLower(tde.Message)
		return tde.Code == 404 ||
			strings.Contains(msg, "chat not found") ||
			strings.Contains(msg, "chat admin privileges") ||
			strings.Contains(msg, "not enough rights") ||
			strings.Contains(msg, "channel private")
	}
	return false
}
