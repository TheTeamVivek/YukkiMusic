package utils

import (
	"sync"
	"time"
)

var (
	floodMap = make(map[string]time.Time)
	floodMu  sync.Mutex
)

// GetFlood returns the remaining cooldown time for a key.
// If zero or negative, the action is allowed.
func GetFlood(key string) time.Duration {
	floodMu.Lock()
	defer floodMu.Unlock()

	if t, exists := floodMap[key]; exists {
		return time.Until(t)
	}
	return 0
}

// SetFlood sets a flood timeout for the key.
// 'duration' specifies how long the key should be blocked.
func SetFlood(key string, duration time.Duration) {
	floodMu.Lock()
	defer floodMu.Unlock()

	floodMap[key] = time.Now().Add(duration)
}
