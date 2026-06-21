package core

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/amarnathcjd/gortc/groupcall"
	"github.com/amarnathcjd/gortc/media"
	"github.com/amarnathcjd/gogram/telegram"
)

var ErrCallNotJoined = errors.New("voice call: not joined")

// VoiceCall wraps a single chat's gortc group-call connection and active
// player. One VoiceCall belongs to one RoomState.
type VoiceCall struct {
	mu sync.RWMutex

	gc     *groupcall.GroupCall
	player *media.Player

	joinCancel context.CancelFunc
	playCancel context.CancelFunc

	chatID int64
}

// NewVoiceCall builds a VoiceCall bound to client, but does not join yet.
func NewVoiceCall(client *telegram.Client, chatID int64, opts ...groupcall.Option) *VoiceCall {
	return &VoiceCall{
		gc:     groupcall.New(client, opts...),
		chatID: chatID,
	}
}

// Join connects to the group call. Blocking; retries internally.
func (v *VoiceCall) Join() error {
	v.mu.Lock()
	if v.gc == nil {
		v.mu.Unlock()
		return ErrCallNotJoined
	}
	ctx, cancel := context.WithCancel(context.Background())
	v.joinCancel = cancel
	gc := v.gc
	chatID := v.chatID
	v.mu.Unlock()

	return gc.JoinCall(ctx, chatID)
}

// Play starts streaming src in the background. Any previous player is
// stopped first.
func (v *VoiceCall) Play(src media.Source) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.gc == nil {
		return ErrCallNotJoined
	}
	v.stopPlayerLocked()

	ctx, cancel := context.WithCancel(context.Background())
	v.playCancel = cancel
	v.player = v.gc.Play(ctx, src)
	return nil
}

// Pause pauses the active player, if any.
func (v *VoiceCall) Pause() {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.player == nil {
		return
	}
	v.player.Pause()
}

// Resume resumes the active player, if any.
func (v *VoiceCall) Resume() {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.player == nil {
		return
	}
	v.player.Resume()
}

// Stop halts the current playback only; the call connection stays alive.
func (v *VoiceCall) Stop() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.stopPlayerLocked()
}

func (v *VoiceCall) stopPlayerLocked() {
	if v.player != nil {
		v.player.Stop()
		v.player = nil
	}
	if v.playCancel != nil {
		v.playCancel()
		v.playCancel = nil
	}
}

// Leave stops playback (if any), leaves the group call, and releases the
// join context. Safe to call even if nothing was ever joined.
func (v *VoiceCall) Leave() error {
	v.mu.Lock()
	v.stopPlayerLocked()
	gc := v.gc
	v.gc = nil
	if v.joinCancel != nil {
		v.joinCancel()
		v.joinCancel = nil
	}
	v.mu.Unlock()

	if gc == nil {
		return nil
	}
	return gc.Leave()
}

// IsJoined reports whether the underlying group-call connection still exists.
func (v *VoiceCall) IsJoined() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.gc != nil
}

// IsPlaying reports whether a player is currently active.
func (v *VoiceCall) IsPlaying() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.player != nil
}

func (v *VoiceCall) Paused() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.player == nil {
		return false
	}
	return v.player.Paused()
}

// Position returns current playback position.
func (v *VoiceCall) Position() time.Duration {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.player == nil {
		return 0
	}
	return v.player.Position()
}

// Duration returns the total duration of the current track.
func (v *VoiceCall) Duration() time.Duration {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.player == nil {
		return 0
	}
	return v.player.Duration()
}

// Done returns the active player's completion channel, or nil if nothing
// is playing.
func (v *VoiceCall) Done() <-chan error {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.player == nil {
		return nil
	}
	return v.player.Done()
}
