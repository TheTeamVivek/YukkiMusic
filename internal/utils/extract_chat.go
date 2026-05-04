package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func ExtractChat(m *telegram.NewMessage) (int64, error) {
	if m == nil || m.Message == nil {
		return 0, fmt.Errorf("invalid message")
	}

	parts := strings.Fields(m.Text())
	if len(parts) < 2 {
		return 0, fmt.Errorf("no chat identifier found")
	}

	target := strings.TrimSpace(parts[1])
	if target == "" {
		return 0, fmt.Errorf("empty chat identifier")
	}

	if id, err := strconv.ParseInt(target, 10, 64); err == nil {
		return id, nil
	}

	target = strings.TrimPrefix(target, "@")
	peer, err := m.Client.ResolvePeer(target)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve peer: %w", err)
	}

	switch p := peer.(type) {
	case *telegram.InputPeerChannel:
		return p.ChannelID, nil
	case *telegram.InputPeerChat:
		return p.ChatID, nil
	default:
		return 0, fmt.Errorf("resolved peer is not a chat")
	}
}
