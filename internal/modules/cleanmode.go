package modules

import (
	"context"
	"sync"
	"time"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/core"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

const (
	defaultCleanDelay = 15 * time.Minute
	cleanTick         = 5 * time.Second
	cleanChunk        = 100 // Telegram accepts at most 100 ids per deleteMessages call
)

// cleaner deletes messages once the chat's clean mode delay has passed.
//
// The player's status message is held while it is the room's active one. Its
// delay starts only when it stops being the active one, i.e. when the next
// track starts or the room ends.
var cleaner = &cleanQueue{chats: make(map[int64]map[int64]time.Time)}

type cleanQueue struct {
	mu    sync.Mutex
	chats map[int64]map[int64]time.Time // chat -> message -> due, zero time means held
}

// Add schedules a message of a group chat. Adding it twice keeps the first deadline.
func (q *cleanQueue) Add(chatID, msgID int64) {
	if chatID >= 0 || msgID == 0 {
		return
	}
	if on, _ := database.CleanMode(chatID); !on {
		return
	}
	due := time.Now().Add(cleanModeDelay(chatID))

	q.mu.Lock()
	defer q.mu.Unlock()
	msgs := q.chats[chatID]
	if msgs == nil {
		msgs = make(map[int64]time.Time)
		q.chats[chatID] = msgs
	}
	if _, ok := msgs[msgID]; !ok {
		msgs[msgID] = due
	}
}

// Forget drops messages that were deleted by someone else.
func (q *cleanQueue) Forget(chatID int64, ids []int64) {
	q.mu.Lock()
	defer q.mu.Unlock()
	msgs := q.chats[chatID]
	for _, id := range ids {
		delete(msgs, id)
	}
	if len(msgs) == 0 {
		delete(q.chats, chatID)
	}
}

// Clear drops everything pending for a chat.
func (q *cleanQueue) Clear(chatID int64) {
	q.mu.Lock()
	delete(q.chats, chatID)
	q.mu.Unlock()
}

func (q *cleanQueue) Run(ctx context.Context) {
	t := time.NewTicker(cleanTick)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			q.sweep(now)
		}
	}
}

func (q *cleanQueue) sweep(now time.Time) {
	q.mu.Lock()
	chatIDs := make([]int64, 0, len(q.chats))
	for id := range q.chats {
		chatIDs = append(chatIDs, id)
	}
	q.mu.Unlock()

	for _, chatID := range chatIDs {
		if on, _ := database.CleanMode(chatID); !on {
			q.Clear(chatID)
			continue
		}
		for ids := q.expired(chatID, now); len(ids) > 0; {
			n := min(len(ids), cleanChunk)
			// A failure is final: the message is usually gone already.
			err := core.Bot.DeleteMessages(chatID, ids[:n], &td.DeleteMessagesOpts{Revoke: true})
			if err != nil {
				utils.Log.Debugf("cleanmode delete failed chat=%d err=%v", chatID, err)
			}
			ids = ids[n:]
		}
	}
}

// expired pops the messages of a chat that are due and updates held ones.
func (q *cleanQueue) expired(chatID int64, now time.Time) []int64 {
	var keep int64
	if room, ok := core.Rooms.Get(chatID); ok && room != nil {
		if m := room.StatusMsg(); m != nil {
			keep = m.Id
		}
	}
	delay := cleanModeDelay(chatID)

	q.mu.Lock()
	defer q.mu.Unlock()
	msgs := q.chats[chatID]

	var ids []int64
	for id, due := range msgs {
		switch {
		case id == keep:
			msgs[id] = time.Time{}
		case due.IsZero():
			msgs[id] = now.Add(delay)
		case !due.After(now):
			ids = append(ids, id)
			delete(msgs, id)
		}
	}
	if len(msgs) == 0 {
		delete(q.chats, chatID)
	}
	return ids
}

var cleanModeDurationOptions = []int{15, 30, 60, 5}

func cleanModeDelay(chatID int64) time.Duration {
	settings, err := database.GetChatSettings(chatID)
	if err != nil || settings.CleanModeDurationMins <= 0 {
		return defaultCleanDelay
	}
	return time.Duration(settings.CleanModeDurationMins) * time.Minute
}

func cleanModeStatusText(chatID int64, enabled bool) string {
	settings, _ := database.GetChatSettings(chatID)
	duration := 15
	if settings != nil && settings.CleanModeDurationMins > 0 {
		duration = settings.CleanModeDurationMins
	}
	return F(
		chatID,
		"cleanmode_status",
		locales.Arg{"action": F(chatID, utils.IfElse(enabled, "enabled", "disabled")), "duration": duration},
	)
}
