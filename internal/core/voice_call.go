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
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"
	"github.com/amarnathcjd/gortc/groupcall"
	"github.com/amarnathcjd/gortc/media"
)

var (
	ErrCallNotJoined     = errors.New("voice call: not joined")
	ErrConnectionTimeout = errors.New("voice call: connection timeout")
)

type VoiceCall struct {
	mu sync.RWMutex

	gc     *groupcall.GroupCall
	player *media.Player
	joined bool

	joinCancel context.CancelFunc
	playCancel context.CancelFunc

	playID atomic.Uint64

	OnStreamEnd func(chatID int64)

	chatID int64
}

func NewVoiceCall(client *telegram.Client, chatID int64, opts ...groupcall.Option) *VoiceCall {
	return &VoiceCall{
		gc:     groupcall.New(client, opts...),
		chatID: chatID,
	}
}

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

	err := gc.JoinCall(ctx, chatID)

	v.mu.Lock()
	v.joined = err == nil
	v.mu.Unlock()

	return err
}

func (v *VoiceCall) Play(src media.Source) error {
	return v.PlayAt(src, 0)
}

func (v *VoiceCall) PlayAt(src media.Source, offset time.Duration) error {
	v.mu.RLock()
	hasGC := v.gc != nil
	joined := v.joined
	v.mu.RUnlock()

	if !hasGC {
		return ErrCallNotJoined
	}
	if !joined {
		if err := v.Join(); err != nil {
			return err
		}
	}

	v.mu.Lock()
	if v.gc == nil {
		v.mu.Unlock()
		return ErrCallNotJoined
	}
	v.stopPlayerLocked()

	ctx, cancel := context.WithCancel(context.Background())
	v.playCancel = cancel

	player := v.gc.Play(ctx, src)
	if offset > 0 {
		if err := player.Seek(offset); err != nil && err != media.ErrNotSeekable {
			cancel()
			v.mu.Unlock()
			return err
		}
	}
	v.player = player
	id := v.playID.Add(1)
	v.mu.Unlock()

	v.watch(player, id)
	return nil
}

func (v *VoiceCall) watch(player *media.Player, id uint64) {
	go func() {
		done := player.Done()
		if done == nil {
			return
		}
		err := <-done
		if err != nil {
			gologging.Error(err)
			return
		}
		v.mu.RLock()
		callback := v.OnStreamEnd
		chatID := v.chatID
		v.mu.RUnlock()
		if callback != nil && v.playID.Load() == id {
			callback(chatID)
		}
	}()
}

func (v *VoiceCall) SeekTo(offset time.Duration) error {
	v.mu.Lock()
	player := v.player
	if player == nil {
		v.mu.Unlock()
		return ErrCallNotJoined
	}
	v.playID.Add(1)
	v.mu.Unlock()

	if err := player.Seek(offset); err != nil {
		return err
	}

	v.mu.Lock()
	id := v.playID.Add(1)
	v.mu.Unlock()

	v.watch(player, id)
	return nil
}

func (v *VoiceCall) Pause() {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.player == nil {
		return
	}
	v.player.Pause()
}

func (v *VoiceCall) Resume() {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.player == nil {
		return
	}
	v.player.Resume()
}

func (v *VoiceCall) Stop() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.stopPlayerLocked()
	v.playID.Add(1)
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

func (v *VoiceCall) Leave() error {
	v.mu.Lock()
	v.stopPlayerLocked()
	v.playID.Add(1)
	v.joined = false
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

func (v *VoiceCall) IsJoined() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.joined
}

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

func (v *VoiceCall) Position() time.Duration {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.player == nil {
		return 0
	}
	return v.player.Position()
}

func (v *VoiceCall) Duration() time.Duration {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.player == nil {
		return 0
	}
	return v.player.Duration()
}

func (v *VoiceCall) Done() <-chan error {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.player == nil {
		return nil
	}
	return v.player.Done()
}