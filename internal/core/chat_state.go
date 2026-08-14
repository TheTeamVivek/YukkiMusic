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

	"yukkimusic/internal/logger"

	td "github.com/AshokShau/gotdbot"
	"github.com/amarnathcjd/gogram/telegram"
)

// Flood-wait handling while the assistant joins a chat.
const (
	shortFloodWait = 10 // seconds: sleep and retry on the same assistant
	longFloodWait  = 30 // seconds: above this the user is told to wait
)

var (
	ErrAdminPermissionRequired  = errors.New("admin permission required")
	ErrStateFetchFailed         = errors.New("state fetch failed")
	ErrAssistantNotAvailable    = errors.New("assistant unavailable")
	ErrAssistantInviteLinkFetch = errors.New("failed to fetch invite link")
	ErrInviteRequestSent        = errors.New("join request sent")
	ErrJoinFailed               = errors.New("assistant join failed")
)

// FloodWaitError reports a flood wait that could not be worked around, so the
// user has to wait before retrying.
type FloodWaitError struct{ Seconds int }

func (e *FloodWaitError) Error() string {
	return fmt.Sprintf("flood wait: %d seconds", e.Seconds)
}

// StateSnapshot is a cached view of the assistant's membership and the
// voice-chat status in a chat.
type StateSnapshot struct {
	AssistantPresent bool
	AssistantBanned  bool
	VoiceChatActive  *bool
}

// ChatState tracks the per-chat state needed to run the assistant in a voice
// chat: which assistant is bound, whether it has joined, and whether a voice
// chat is active. StateSnapshot is embedded so the snapshot fields are read
// and written directly.
type ChatState struct {
	mu sync.RWMutex

	ChatID int64

	Assistant *Assistant

	StateSnapshot
	inviteLink string
	fetched    bool
}

var (
	chatStatesMu sync.Mutex
	chatStates   = map[int64]*ChatState{}
)

// ChatStateFor returns the cached ChatState for a chat, binding an assistant
// to it on first use.
func ChatStateFor(chatID int64) (*ChatState, error) {
	chatStatesMu.Lock()
	state := chatStates[chatID]
	if state == nil {
		state = &ChatState{ChatID: chatID}
		chatStates[chatID] = state
	}
	chatStatesMu.Unlock()
	if err := state.bindAssistant(); err != nil {
		logger.Errorf("chat_state: bindAssistant failed for %d: %v", chatID, err)
		return nil, err
	}
	return state, nil
}

func DropChatState(chatID int64) {
	chatStatesMu.Lock()
	delete(chatStates, chatID)
	chatStatesMu.Unlock()
}

// FullChat fetches the supergroup full info for a chat ID. TDLib supergroup
// chat IDs carry a -100 prefix, while getSupergroupFullInfo expects the raw
// supergroup id, so no separate getChat lookup is needed.
func FullChat(c *td.Client, chatID int64) (*td.SupergroupFullInfo, error) {
	return c.GetSupergroupFullInfo(-chatID - 1_000_000_000_000)
}

// Snapshot returns the cached state, refreshing it first when none is cached
// yet.
func (s *ChatState) Snapshot() (StateSnapshot, error) {
	if snap, ok := s.cached(); ok {
		return snap, nil
	}
	return s.refresh()
}

// Refresh forces a refresh of the cached state from Telegram.
func (s *ChatState) Refresh() (StateSnapshot, error) {
	return s.refresh()
}

// Join makes sure the assistant is a member of the chat. For public chats the
// resolved username is used as the join target (no admin rights needed); on
// failure it falls back to the real invite-link flow. Short flood waits are
// slept through, longer ones switch to another assistant, and if none can
// join a FloodWaitError is returned so the user can be told to wait.
func (s *ChatState) Join() error {
	if username := s.publicUsername(); username != "" {
		s.setLink("https://t.me/" + username)
	}

	err := s.joinWithRetry()
	if err == nil {
		return nil
	}

	wait := telegram.GetFloodWait(err)
	if wait == 0 {
		// Not a flood error: retry with the real invite link.
		s.setLink("")
		return s.joinWithRetry()
	}
	if wait <= shortFloodWait {
		logger.Errorf("chat_state: flood wait %ds while joining %d, retrying", wait, s.ChatID)
		time.Sleep(time.Duration(wait) * time.Second)
		return s.joinWithRetry()
	}

	// Longer flood waits: join with another assistant instead of blocking.
	if s.switchAssistant() {
		return nil
	}

	if wait < longFloodWait {
		logger.Errorf("chat_state: flood wait %ds while joining %d, retrying", wait, s.ChatID)
		time.Sleep(time.Duration(wait) * time.Second)
		return s.joinWithRetry()
	}

	return &FloodWaitError{Seconds: wait}
}

func (s *ChatState) SetPresent(v bool) {
	s.mu.Lock()
	s.AssistantPresent = v
	s.fetched = true
	s.mu.Unlock()
}

func (s *ChatState) SetBanned(v bool) {
	s.mu.Lock()
	s.AssistantBanned = v
	s.fetched = true
	s.mu.Unlock()
}

func (s *ChatState) SetVoiceChat(v bool) {
	s.mu.Lock()
	s.VoiceChatActive = &v
	s.fetched = true
	s.mu.Unlock()
}

func (s *ChatState) Fetched() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.fetched
}

// refresh fetches the assistant membership, voice-chat status and invite link
// from Telegram and stores them as the new snapshot.
func (s *ChatState) refresh() (StateSnapshot, error) {
	logger.Debugf("chat_state: refresh(chat=%d)", s.ChatID)
	if s.Assistant == nil {
		if err := s.bindAssistant(); err != nil {
			return StateSnapshot{}, err
		}
	}
	full, err := FullChat(Bot, s.ChatID)
	if err != nil {
		logger.Errorf("chat_state: FullChat failed for %d: %v", s.ChatID, err)
		if isAdminError(err) {
			return StateSnapshot{}, ErrAdminPermissionRequired
		}
		return StateSnapshot{}, fmt.Errorf("%w: %v", ErrStateFetchFailed, err)
	}

	member, err := Bot.GetChatMember(s.ChatID, &td.MessageSenderUser{UserId: s.Assistant.Self.ID})
	if err != nil {
		if isAdminError(err) {
			logger.Errorf("chat_state: admin permission required for GetChatMember in %d", s.ChatID)
			return StateSnapshot{}, ErrAdminPermissionRequired
		}
		return StateSnapshot{}, fmt.Errorf("%w: %v", ErrStateFetchFailed, err)
	}

	present, banned := membership(member)

	// A bot cannot observe the group call state (getGroupCall is user-only and
	// the group call id is always 0 for bots), so voice chat activity is
	// queried through the assistant like Grogram's fullChannel.Call != nil.
	vcActive := s.VoiceChatActive
	if active, err := s.voiceChatActive(); err != nil {
		logger.Warnf("chat_state: voiceChatActive failed for %d: %v", s.ChatID, err)
	} else {
		vcActive = &active
	}
	s.setState(present, banned, vcActive)
	if full != nil &&
		full.InviteLink != nil &&
		full.InviteLink.InviteLink != "" {
		s.setLink(full.InviteLink.InviteLink)
	}
	return s.current(), nil
}

// voiceChatActive reports whether an active voice/video chat is running in the
// chat. The bot's TDLib client cannot observe this, so it is queried through
// the assistant's MTProto client (getFullChannel -> Call != nil), mirroring
// Grogram's ChannelFull.Call check.
func (s *ChatState) voiceChatActive() (bool, error) {
	if s.Assistant == nil || s.Assistant.Client == nil {
		return false, ErrAssistantNotAvailable
	}

	peer, err := s.Assistant.Client.ResolvePeer(s.ChatID)
	if err != nil {
		return false, fmt.Errorf("resolve peer: %w", err)
	}

	switch p := peer.(type) {
	case *telegram.InputPeerChannel:
		full, err := s.Assistant.Client.ChannelsGetFullChannel(&telegram.InputChannelObj{
			ChannelID:  p.ChannelID,
			AccessHash: p.AccessHash,
		})
		if err != nil {
			return false, fmt.Errorf("get full channel: %w", err)
		}
		ch, ok := full.FullChat.(*telegram.ChannelFull)
		return ok && ch.Call != nil, nil

	case *telegram.InputPeerChat:
		full, err := s.Assistant.Client.MessagesGetFullChat(p.ChatID)
		if err != nil {
			return false, fmt.Errorf("get full chat: %w", err)
		}
		ch, ok := full.FullChat.(*telegram.ChatFullObj)
		return ok && ch.Call != nil, nil
	}

	return false, nil
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

// publicUsername returns the chat's public username, or "" when the chat has
// none (private group/channel). Public chats let the assistant join without
// needing an invite link or admin rights.
func (s *ChatState) publicUsername() string {
	chat, err := Bot.GetChat(s.ChatID)
	if err != nil {
		return ""
	}
	ct, ok := chat.Type.(*td.ChatTypeSupergroup)
	if !ok {
		return ""
	}
	sg, err := Bot.GetSupergroup(ct.SupergroupId)
	if err != nil || sg == nil || sg.Usernames == nil {
		return ""
	}
	if len(sg.Usernames.ActiveUsernames) == 0 {
		return ""
	}
	return sg.Usernames.ActiveUsernames[0]
}

// joinWithRetry joins using the cached invite link, retrying once after
// refreshing the link when it has expired.
func (s *ChatState) joinWithRetry() error {
	err := s.joinByLink()
	if err == nil {
		return nil
	}
	if !telegram.MatchError(err, "INVITE_HASH_EXPIRED") {
		return err
	}
	logger.Errorf("chat_state: invite expired for %d, retrying with refreshed link", s.ChatID)
	s.setLink("")
	return s.joinByLink()
}

func (s *ChatState) joinByLink() error {
	logger.Debugf("chat_state: joinByLink(chat=%d)", s.ChatID)
	link, err := s.resolveLink()
	if err != nil {
		return err
	}
	_, err = s.Assistant.Client.JoinChannel(link)
	if err == nil || telegram.MatchError(err, "USER_ALREADY_PARTICIPANT") {
		s.setState(true, false, s.VoiceChatActive)
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
		s.leaveInactiveChats(5)
		time.Sleep(1 * time.Second)
		_, retryErr := s.Assistant.Client.JoinChannel(link)
		if retryErr == nil || telegram.MatchError(retryErr, "USER_ALREADY_PARTICIPANT") {
			s.setState(true, false, s.VoiceChatActive)
			return nil
		}
		return retryErr
	}
	if isAdminError(err) {
		return ErrAdminPermissionRequired
	}
	return fmt.Errorf("%w: %v", ErrJoinFailed, err)
}

func (s *ChatState) resolveLink() (string, error) {
	logger.Debugf("chat_state: resolveLink(chat=%d)", s.ChatID)
	s.mu.RLock()
	cached := s.inviteLink
	s.mu.RUnlock()
	if cached != "" {
		return cached, nil
	}
	full, err := FullChat(Bot, s.ChatID)
	if err != nil {
		if isAdminError(err) {
			return "", ErrAdminPermissionRequired
		}
		return "", fmt.Errorf("%w: %v", ErrAssistantInviteLinkFetch, err)
	}
	if full != nil &&
		full.InviteLink != nil &&
		full.InviteLink.InviteLink != "" {
		s.setLink(full.InviteLink.InviteLink)
		return full.InviteLink.InviteLink, nil
	}
	inv, err := Bot.CreateChatInviteLink(s.ChatID, 0, 0, "", nil)
	if err != nil {
		if isAdminError(err) {
			return "", ErrAdminPermissionRequired
		}
		return "", fmt.Errorf("%w: %v", ErrAssistantInviteLinkFetch, err)
	}
	if inv == nil || inv.InviteLink == "" {
		return "", ErrAssistantInviteLinkFetch
	}
	s.setLink(inv.InviteLink)
	return inv.InviteLink, nil
}

func (s *ChatState) approveJoinRequest() error {
	logger.Debugf("chat_state: approveJoinRequest(chat=%d)", s.ChatID)
	err := Bot.ProcessChatJoinRequest(
		s.ChatID,
		s.Assistant.Self.ID,
		&td.ProcessChatJoinRequestOpts{Approve: true},
	)
	if err == nil {
		s.setState(true, false, s.VoiceChatActive)
		return nil
	}
	if isAdminError(err) {
		return ErrAdminPermissionRequired
	}
	return err
}

// switchAssistant tries to join the chat with each other assistant, keeping
// the first one that succeeds. It reports whether the chat was joined.
func (s *ChatState) switchAssistant() bool {
	original := s.Assistant
	if original == nil {
		return false
	}

	for idx := 1; idx <= Assistants.Count(); idx++ {
		if idx == original.Index+1 {
			continue
		}
		ass, err := Assistants.Get(idx)
		if err != nil {
			continue
		}
		s.bind(ass)
		if err := s.joinByLink(); err == nil {
			Assistants.Assign(s.ChatID, idx)
			s.rebindRoom(ass)
			logger.Infof("chat_state: switched assistant for %d to index %d", s.ChatID, idx)
			return true
		}
	}

	s.bind(original)
	return false
}

// bindAssistant binds the chat to its assistant, caching the choice.
func (s *ChatState) bindAssistant() error {
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

// bind switches the chat to a different assistant without touching the cached
// assignment.
func (s *ChatState) bind(ass *Assistant) {
	s.mu.Lock()
	s.Assistant = ass
	s.mu.Unlock()
}

// rebindRoom points the room (if any) at the assistant now serving the chat.
func (s *ChatState) rebindRoom(ass *Assistant) {
	if room, ok := GetRoom(s.ChatID, nil, false); ok {
		room.SetAssistant(ass)
	}
}

// leaveInactiveChats makes room for joining a new chat by leaving older chats
// that are not currently streaming.
func (s *ChatState) leaveInactiveChats(limit int) {
	logger.Debugf("chat_state: leaveInactiveChats(chat=%d, limit=%d)", s.ChatID, limit)
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

func (s *ChatState) cached() (StateSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.StateSnapshot, s.fetched
}

func (s *ChatState) current() StateSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.StateSnapshot
}

func (s *ChatState) setState(present, banned bool, vcActive *bool) {
	s.mu.Lock()
	s.StateSnapshot = StateSnapshot{
		AssistantPresent: present,
		AssistantBanned:  banned,
		VoiceChatActive:  vcActive,
	}
	s.fetched = true
	s.mu.Unlock()
}

func (s *ChatState) setLink(link string) { s.mu.Lock(); s.inviteLink = link; s.mu.Unlock() }

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
