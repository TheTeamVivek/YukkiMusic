package utils

import (
	"fmt"

	"github.com/amarnathcjd/gogram/telegram"
)

func ExtractURLs(m *telegram.NewMessage) ([]string, error) {
	if m == nil || m.Message == nil {
		return nil, fmt.Errorf("invalid message")
	}
	capacity := len(m.Message.Entities)
	if m.IsReply() {
		if r, err := m.GetReplyMessage(); err == nil {
			capacity += len(r.Message.Entities)
		}
	}

	urls := make([]string, 0, capacity)

	collect := func(msg *telegram.MessageObj) {
		text := msg.Message
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
	}

	collect(m.Message)

	if m.IsReply() {
		r, err := m.GetReplyMessage()
		if err != nil {
			if len(urls) > 0 {
				return urls, fmt.Errorf("failed to fetch reply message: %w", err)
			}
			return nil, fmt.Errorf("failed to fetch reply message: %w", err)
		}
		collect(r.Message)
	}

	if len(urls) == 0 {
		return nil, fmt.Errorf("no URLs found")
	}

	return urls, nil
}
