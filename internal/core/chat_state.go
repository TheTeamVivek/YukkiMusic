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

	"yukkimusic/internal/database"
	"yukkimusic/internal/utils"

	td "github.com/AshokShau/gotdbot"
	"github.com/amarnathcjd/gogram/telegram"
)

// Flood-wait handling while the assistant joins a chat.
const (
	shortFloodWait = 10 // seconds: sleep and retry with the same assistant
	longFloodWait  = 30 // seconds: from here on the user is told to wait
	leaveBatch     = 5  // chats an assistant leaves when it hit the join limit
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

// Presence is where the assistant stands in a chat. The zero value means it
// has not been looked up yet.
type Presence uint8

const (
	PresenceUnknown Presence = iota
	PresenceAbsent           // not in the chat (never joined, or left)
	PresenceJoined           // a member, possibly muted
	PresenceBanned           // banned: cannot join until unbanned
)

// Voice is what is known about the chat's voice chat. The zero value means
// nothing is known.
type Voice uint8

const (
	VoiceUnknown Voice = iota
	VoiceInactive
	VoiceActive
)

// Status is the cached view of a chat: the assistant's presence and the
// voice chat.
type Status struct {
	Presence Presence
	Voice    Voice
}

// ChatState is the per-chat state needed to run the assistant in a voice chat:
// which assistant serves the chat, whether it is in the chat, and whether a
// voice chat is running. Get one with ChatStateFor.
//
// Typical use:
//
//	cs, err := core.ChatStateFor(chatID)
//	st, err := cs.Status()          // cached, looked up once if unknown
//	if st.Presence != core.PresenceJoined { err = cs.Join() }
//
// Event handlers keep it current through SetPresence and SetVoice; Refresh
// forces a fresh lookup from Telegram.
type ChatState struct {
	ChatID int64

	// op serializes the calls that talk to Telegram (Refresh, Join, Unban),
	// so concurrent commands in one chat never join twice.
	op sync.Mutex

	mu        sync.RWMutex // guards the fields below
	assistant *Assistant
	status    Status
	link      string // cached invite link
}

var (
	chatStatesMu sync.Mutex
	chatStates   = map[int64]*ChatState{}
)

// ChatStateFor returns the ChatState of a chat, binding the chat's assistant
// on first use.
func ChatStateFor(chatID int64) (*ChatState, error) {
	chatStatesMu.Lock()
	s := chatStates[chatID]
	if s == nil {
		s = &ChatState{ChatID: chatID}
		chatStates[chatID] = s
	}
	chatStatesMu.Unlock()

	if s.Assistant() != nil {
		return s, nil
	}
	ass, err := Assistants.ForChat(chatID)
	if err != nil {
		utils.Log.Errorf("chat_state: no assistant for %d: %v", chatID, err)
		return nil, fmt.Errorf("%w: %w", ErrAssistantNotAvailable, err)
	}
	s.mu.Lock()
	if s.assistant == nil {
		s.assistant = ass
	}
	s.mu.Unlock()
	return s, nil
}

// DropChatState forgets the state of a chat the bot no longer serves.
func DropChatState(chatID int64) {
	chatStatesMu.Lock()
	delete(chatStates, chatID)
	chatStatesMu.Unlock()
}

// Assistant returns the assistant serving the chat.
func (s *ChatState) Assistant() *Assistant {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.assistant
}

// Status returns the cached status, looking it up first if the assistant's
// presence is not known yet.
func (s *ChatState) Status() (Status, error) {
	if st := s.cached(); st.Presence != PresenceUnknown {
		return st, nil
	}
	return s.Refresh()
}

// SetPresence records the assistant's presence, e.g. from a member update.
func (s *ChatState) SetPresence(p Presence) {
	s.mu.Lock()
	s.status.Presence = p
	s.mu.Unlock()
}

// SetVoice records the voice chat state, e.g. from a service message.
func (s *ChatState) SetVoice(v Voice) {
	s.mu.Lock()
	s.status.Voice = v
	s.mu.Unlock()
}

// PresenceOf maps a TDLib chat member to the assistant's presence.
func PresenceOf(m *td.ChatMember) Presence {
	if m == nil {
		return PresenceAbsent
	}
	switch st := m.Status.(type) {
	case *td.ChatMemberStatusBanned:
		return PresenceBanned
	case *td.ChatMemberStatusRestricted:
		// A restricted non-member has left and may rejoin; only a ban blocks.
		if st.IsMember {
			return PresenceJoined
		}
	case *td.ChatMemberStatusMember,
		*td.ChatMemberStatusAdministrator,
		*td.ChatMemberStatusCreator:
		return PresenceJoined
	}
	return PresenceAbsent
}

// Refresh looks the assistant's presence and the voice chat state up from
// Telegram. A banned assistant cannot see the voice chat, so then the voice
// state is left as it was.
func (s *ChatState) Refresh() (Status, error) {
	s.op.Lock()
	defer s.op.Unlock()

	ass := s.Assistant()
	member, err := Bot.GetChatMember(s.ChatID, &td.MessageSenderUser{UserId: ass.Self.ID})
	if err != nil {
		if isAdminError(err) {
			utils.Log.Errorf("chat_state: admin permission required for GetChatMember in %d", s.ChatID)
			return Status{}, ErrAdminPermissionRequired
		}
		return Status{}, fmt.Errorf("%w: %w", ErrStateFetchFailed, err)
	}

	presence := PresenceOf(member)
	s.SetPresence(presence)

	if presence != PresenceBanned {
		active, err := voiceActive(ass, s.ChatID)
		switch {
		case err != nil:
			utils.Log.Warnf("chat_state: voice chat lookup failed for %d: %v", s.ChatID, err)
		case active:
			s.SetVoice(VoiceActive)
		default:
			s.SetVoice(VoiceInactive)
		}
	}
	return s.cached(), nil
}

// Unban lifts the assistant's ban through the bot. The assistant is not a
// member afterwards, so it still has to Join.
func (s *ChatState) Unban() error {
	s.op.Lock()
	defer s.op.Unlock()

	err := Bot.SetChatMemberStatus(
		s.ChatID,
		&td.MessageSenderUser{UserId: s.Assistant().Self.ID},
		&td.ChatMemberStatusLeft{},
	)
	if err == nil {
		s.SetPresence(PresenceAbsent)
	}
	return err
}

// Join makes sure the assistant is in the chat. A short flood wait is slept
// through, a longer one hands the chat to another assistant, and if none can
// join a *FloodWaitError tells the user how long to wait.
func (s *ChatState) Join() error {
	s.op.Lock()
	defer s.op.Unlock()

	ass := s.Assistant()
	err := s.joinWith(ass)
	if wait := telegram.GetFloodWait(err); wait > 0 {
		err = s.afterFlood(ass, wait)
	}

	switch {
	case err == nil:
		s.SetPresence(PresenceJoined)
	case matches(err, "USER_BANNED_IN_CHANNEL"):
		s.SetPresence(PresenceBanned)
	}
	return err
}

func (s *ChatState) afterFlood(ass *Assistant, wait int) error {
	if wait > shortFloodWait {
		if s.switchAssistant(ass) {
			return nil
		}
		if wait >= longFloodWait {
			return &FloodWaitError{Seconds: wait}
		}
	}

	utils.Log.Errorf("chat_state: flood wait %ds while joining %d, retrying", wait, s.ChatID)
	time.Sleep(time.Duration(wait) * time.Second)

	err := s.joinWith(ass)
	if again := telegram.GetFloodWait(err); again > 0 {
		return &FloodWaitError{Seconds: again}
	}
	return err
}

// switchAssistant joins the chat with each other assistant and keeps the first
// one that succeeds, reporting whether the chat was joined.
func (s *ChatState) switchAssistant(current *Assistant) bool {
	for idx := 1; idx <= Assistants.Count(); idx++ {
		ass, err := Assistants.Get(idx)
		if err != nil || ass == current {
			continue
		}
		if err := s.joinWith(ass); err != nil {
			continue
		}

		s.mu.Lock()
		s.assistant = ass
		s.status.Presence = PresenceJoined
		s.mu.Unlock()

		database.SaveAssistant(s.ChatID, ass.Index)
		if room, ok := Rooms.Get(s.ChatID); ok {
			room.SetAssistant(ass)
		}
		utils.Log.Infof("chat_state: switched assistant for %d to %d", s.ChatID, ass.Index)
		return true
	}
	return false
}

// joinWith makes ass join the chat. A public chat is joined by its username,
// which needs no admin rights; otherwise (or if that fails) by invite link.
func (s *ChatState) joinWith(ass *Assistant) error {
	if name := s.username(); name != "" {
		err := s.attempt(ass, "https://t.me/"+name)
		if err == nil || telegram.GetFloodWait(err) > 0 || errors.Is(err, ErrInviteRequestSent) {
			return err
		}
	}

	link, err := s.inviteLink("")
	if err != nil {
		return err
	}
	err = s.attempt(ass, link)
	if matches(err, "INVITE_HASH_EXPIRED") {
		if link, err = s.inviteLink(link); err != nil {
			return err
		}
		err = s.attempt(ass, link)
	}
	return err
}

// attempt joins through one link. Hitting the per-account channel limit makes
// the assistant leave a few inactive chats and try once more.
func (s *ChatState) attempt(ass *Assistant, link string) error {
	_, err := ass.Client.JoinChannel(link)
	if matches(err, "USER_CHANNELS_TOO_MUCH", "CHANNELS_TOO_MUCH") {
		utils.Log.Infof("chat_state: join limit reached for %d, leaving inactive chats", s.ChatID)
		s.leaveInactiveChats(ass, leaveBatch)
		time.Sleep(time.Second)
		_, err = ass.Client.JoinChannel(link)
	}

	switch {
	case err == nil, matches(err, "USER_ALREADY_PARTICIPANT"):
		return nil
	case matches(err, "INVITE_REQUEST_SENT"):
		// The chat needs approval: the bot can give it when it is an admin.
		approve := &td.ProcessChatJoinRequestOpts{Approve: true}
		if err := Bot.ProcessChatJoinRequest(s.ChatID, ass.Self.ID, approve); err != nil {
			utils.Log.Errorf("chat_state: cannot approve join request in %d: %v", s.ChatID, err)
			return ErrInviteRequestSent
		}
		return nil
	case isAdminError(err):
		return ErrAdminPermissionRequired
	default:
		return fmt.Errorf("%w: %w", ErrJoinFailed, err)
	}
}

// username returns the chat's public username, or "" for a private chat.
func (s *ChatState) username() string {
	chat, err := Bot.GetChat(s.ChatID)
	if err != nil {
		return ""
	}
	ct, ok := chat.Type.(*td.ChatTypeSupergroup)
	if !ok {
		return ""
	}
	sg, err := Bot.GetSupergroup(ct.SupergroupId)
	if err != nil || sg == nil || sg.Usernames == nil || len(sg.Usernames.ActiveUsernames) == 0 {
		return ""
	}
	return sg.Usernames.ActiveUsernames[0]
}

// inviteLink returns an invite link for the chat, creating one when the chat
// has none. stale is a link that just failed and must not be returned again.
func (s *ChatState) inviteLink(stale string) (string, error) {
	s.mu.RLock()
	link := s.link
	s.mu.RUnlock()
	if link != "" && link != stale {
		return link, nil
	}

	full, err := Bot.GetSupergroupFullInfo(-s.ChatID - 1_000_000_000_000)
	if err != nil {
		if isAdminError(err) {
			return "", ErrAdminPermissionRequired
		}
		return "", fmt.Errorf("%w: %w", ErrAssistantInviteLinkFetch, err)
	}
	if full != nil && full.InviteLink != nil &&
		full.InviteLink.InviteLink != "" && full.InviteLink.InviteLink != stale {
		return s.setLink(full.InviteLink.InviteLink), nil
	}

	inv, err := Bot.CreateChatInviteLink(s.ChatID, nil)
	if err != nil {
		if isAdminError(err) {
			return "", ErrAdminPermissionRequired
		}
		return "", fmt.Errorf("%w: %w", ErrAssistantInviteLinkFetch, err)
	}
	if inv == nil || inv.InviteLink == "" {
		return "", ErrAssistantInviteLinkFetch
	}
	return s.setLink(inv.InviteLink), nil
}

func (s *ChatState) setLink(link string) string {
	s.mu.Lock()
	s.link = link
	s.mu.Unlock()
	return link
}

func (s *ChatState) cached() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

// leaveInactiveChats makes room for a new chat by leaving a few older chats
// of the assistant that are not streaming.
func (s *ChatState) leaveInactiveChats(ass *Assistant, limit int) {
	rooms := Rooms.All()
	left := 0
	err := ass.Client.IterDialogs(func(d *telegram.TLDialog) error {
		if d == nil || d.IsUser() {
			return nil
		}
		id := d.GetChannelID()
		if id == 0 || id == s.ChatID {
			return nil
		}
		if _, ok := rooms[id]; ok {
			return nil
		}

		if err := ass.Client.LeaveChannel(id); err != nil {
			if wait := telegram.GetFloodWait(err); wait > 0 {
				time.Sleep(time.Duration(wait) * time.Second)
			}
			return nil
		}
		if left++; left >= limit {
			return telegram.ErrStopIteration
		}
		return nil
	}, &telegram.DialogOptions{Limit: int32(limit * 20)})

	if err != nil && err != telegram.ErrStopIteration {
		utils.Log.Warnf("chat_state: IterDialogs failed while leaving chats: %v", err)
	}
}

// voiceActive reports whether a voice/video chat is running in the chat. The
// bot's TDLib client cannot observe this (getGroupCall is user-only), so it is
// asked through the assistant's MTProto client: getFullChannel -> Call != nil.
func voiceActive(ass *Assistant, chatID int64) (bool, error) {
	if ass == nil || ass.Client == nil {
		return false, ErrAssistantNotAvailable
	}

	peer, err := ass.Client.ResolvePeer(chatID)
	if err != nil {
		return false, fmt.Errorf("resolve peer: %w", err)
	}

	switch p := peer.(type) {
	case *telegram.InputPeerChannel:
		full, err := ass.Client.ChannelsGetFullChannel(&telegram.InputChannelObj{
			ChannelID:  p.ChannelID,
			AccessHash: p.AccessHash,
		})
		if err != nil {
			return false, fmt.Errorf("get full channel: %w", err)
		}
		ch, ok := full.FullChat.(*telegram.ChannelFull)
		return ok && ch.Call != nil, nil

	case *telegram.InputPeerChat:
		full, err := ass.Client.MessagesGetFullChat(p.ChatID)
		if err != nil {
			return false, fmt.Errorf("get full chat: %w", err)
		}
		ch, ok := full.FullChat.(*telegram.ChatFullObj)
		return ok && ch.Call != nil, nil
	}

	return false, nil
}

// matches reports whether err is one of the given Telegram error codes.
func matches(err error, codes ...string) bool {
	if err == nil {
		return false
	}
	for _, code := range codes {
		if telegram.MatchError(err, code) {
			return true
		}
	}
	return false
}

func isAdminError(err error) bool {
	if telegram.MatchError(err, "CHAT_ID_INVALID") || telegram.MatchError(err, "CHAT_ADMIN_REQUIRED") ||
		telegram.MatchError(err, "CHANNEL_PRIVATE") || telegram.MatchError(err, "CHANNEL_INVALID") {
		return true
	}

	if tde, ok := errors.AsType[*td.Error](err); ok {
		msg := strings.ToLower(tde.Message)
		return tde.Code == 404 ||
			strings.Contains(msg, "chat not found") ||
			strings.Contains(msg, "chat admin privileges") ||
			strings.Contains(msg, "not enough rights") ||
			strings.Contains(msg, "channel private")
	}
	return false
}
