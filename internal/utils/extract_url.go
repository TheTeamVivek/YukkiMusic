package utils

import (
	"fmt"

	"github.com/amarnathcjd/gogram/telegram"
)

func ExtractURLs(m *telegram.NewMessage) ([]string, error) {
	if m == nil || m.Message == nil {
		return nil, fmt.Errorf("invalid message")
	}

	urls := make([]string, 0, estimateCapacity(m))
	urls = append(urls, collectURLs(m.Message)...)

	if !m.IsReply() {
		return finalizeURLs(urls)
	}

	r, err := m.GetReplyMessage()
	if err != nil {
		if len(urls) > 0 {
			return urls, fmt.Errorf("failed to fetch reply message: %w", err)
		}
		return nil, fmt.Errorf("failed to fetch reply message: %w", err)
	}

	urls = append(urls, collectURLs(r.Message)...)
	return finalizeURLs(urls)
}

// --- Sub Functions ---

func estimateCapacity(m *telegram.NewMessage) int {
	capacity := len(m.Message.Entities)
	if m.IsReply() {
		if r, err := m.GetReplyMessage(); err == nil && r.Message != nil {
			capacity += len(r.Message.Entities)
		}
	}
	return capacity
}

func collectURLs(msg *telegram.MessageObj) []string {
	if msg == nil {
		return nil
	}

	text := msg.Message
	urls := make([]string, 0, len(msg.Entities))

	for _, ent := range msg.Entities {
		switch e := ent.(type) {
		case *telegram.MessageEntityURL:
			if int(e.Offset+e.Length) <= len(text) {
				urls = append(urls, text[e.Offset:e.Offset+e.Length])
			}
		case *telegram.MessageEntityTextURL:
			if e.URL != "" {
				urls = append(urls, e.URL)
			}
		}
	}
	return urls
}

func finalizeURLs(urls []string) ([]string, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("no URLs found")
	}
	return urls, nil
}
