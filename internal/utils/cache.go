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
package utils

import (
	"sync"
	"time"
)

type CacheItem[V any] struct {
	Value      V
	Expiration int64
}

func (i CacheItem[V]) Expired() bool {
	return i.Expiration > 0 && time.Now().UnixMilli() > i.Expiration
}

type Cache[K comparable, V any] struct {
	mu         sync.RWMutex
	items      map[K]CacheItem[V]
	defaultTTL int64
}

func NewCache[K comparable, V any](defaultTTL time.Duration) *Cache[K, V] {
	return &Cache[K, V]{
		items:      make(map[K]CacheItem[V]),
		defaultTTL: defaultTTL.Milliseconds(),
	}
}

func (c *Cache[K, V]) Set(key K, value V, ttl ...time.Duration) {
	var exp int64
	now := time.Now().UnixMilli()

	if len(ttl) > 0 && ttl[0] > 0 {
		exp = now + ttl[0].Milliseconds()
	} else if c.defaultTTL > 0 {
		exp = now + c.defaultTTL
	}

	c.mu.Lock()
	c.items[key] = CacheItem[V]{Value: value, Expiration: exp}
	c.mu.Unlock()
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	item, ok := c.items[key]
	if !ok || item.Expired() {
		if ok {
			delete(c.items, key)
		}
		c.mu.Unlock()
		var zero V
		return zero, false
	}
	c.mu.Unlock()
	return item.Value, true
}

func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}
