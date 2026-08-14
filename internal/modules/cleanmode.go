package modules

import (
	"sync"
	"time"

	"yukkimusic/internal/logger"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/internal/core"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/utils"
)

const (
	cleanModeBatchWindow = 10 * time.Second
	defaultCleanDelay    = 15 * time.Minute
)

var (
	cleanModeDurationOptions = []int{15, 30, 60, 5}
	cleanScheduler           = &CleanScheduler{pending: make(map[int64][]cleanEntry)}
)

func cleanModeReadHandler(c *td.Client, upd *td.UpdateChatReadOutbox) error {
	if upd == nil || upd.LastReadOutboxMessageId == 0 {
		return nil
	}

	chatID := upd.ChatId
	if upd.LastReadOutboxMessageId == currentPlayingStatusMessageID(chatID) {
		return nil
	}

	cleanScheduler.schedule(chatID, upd.LastReadOutboxMessageId)
	return nil
}

func currentPlayingStatusMessageID(chatID int64) int64 {
	room, ok := core.RoomFor(chatID)
	if !ok || room == nil {
		return 0
	}
	status := room.StatusMsg()
	if status == nil {
		return 0
	}
	return status.Id
}

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

type cleanEntry struct {
	messageID int64
	dueAt     time.Time
}

type CleanScheduler struct {
	mu      sync.Mutex
	pending map[int64][]cleanEntry
}

func (s *CleanScheduler) start() {
	go func() {
		ticker := time.NewTicker(cleanModeBatchWindow)
		defer ticker.Stop()
		for range ticker.C {
			s.flushDue(time.Now())
		}
	}()
}

func (s *CleanScheduler) schedule(chatID int64, messageID int64) {
	if messageID == 0 {
		return
	}
	s.mu.Lock()
	s.pending[chatID] = append(s.pending[chatID], cleanEntry{
		messageID: messageID,
		dueAt:     time.Now().Add(cleanModeDelay(chatID)),
	})
	s.mu.Unlock()
}

func (s *CleanScheduler) cancel(chatID int64) {
	s.mu.Lock()
	delete(s.pending, chatID)
	s.mu.Unlock()
}

func (s *CleanScheduler) flushDue(deadline time.Time) {
	s.mu.Lock()
	batches := make(map[int64][]int64)

	for chatID, entries := range s.pending {
		statusID := currentPlayingStatusMessageID(chatID)
		keep := entries[:0]

		for _, entry := range entries {
			if entry.messageID == statusID || entry.dueAt.After(deadline) {
				keep = append(keep, entry)
				continue
			}
			batches[chatID] = append(batches[chatID], entry.messageID)
		}

		if len(keep) == 0 {
			delete(s.pending, chatID)
		} else {
			s.pending[chatID] = keep
		}
	}
	s.mu.Unlock()

	for chatID, ids := range batches {
		enabled, err := database.CleanMode(chatID)
		if err != nil || !enabled {
			continue
		}
		if err := core.Bot.DeleteMessages(chatID, ids, nil); err != nil {
			logger.Debugf("cleanmode delete failed chat=%d err=%v", chatID, err)
		}
	}
}

func scheduleOldPlayingMessage(r *core.RoomState) {
	if m := r.StatusMsg(); m != nil {
		cleanScheduler.schedule(m.ChatId, m.Id)
	}
}
