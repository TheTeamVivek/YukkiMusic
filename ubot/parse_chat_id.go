package ubot

import (
	"fmt"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func (ctx *Context) parseChatID(chatID any) (int64, error) {
	var parsedChatID int64
	switch v := chatID.(type) {
	case int64:
		parsedChatID = v
	case int:
		parsedChatID = int64(v)
	case int32:
		parsedChatID = int64(v)
	case int16:
		parsedChatID = int64(v)
	case int8:
		parsedChatID = int64(v)
	case string:
		rawChat, err := ctx.app.ResolveUsername(chatID.(string))
		if err != nil {
			return 0, fmt.Errorf("failed to resolve username: %w", err)
		}
		switch chat := rawChat.(type) {
		case *tg.UserObj:
			parsedChatID = chat.ID
		case *tg.ChatObj:
			parsedChatID = -chat.ID
		case *tg.Channel:
			parsedChatID = -1000000000000 - chat.ID
		}
	case tg.Peer:
		switch p := v.(type) {
		case *tg.PeerUser:
			parsedChatID = p.UserID
		case *tg.PeerChat:
			parsedChatID = -p.ChatID
		case *tg.PeerChannel:
			parsedChatID = -1000000000000 - p.ChannelID
		}
	default:
		return 0, fmt.Errorf("unsupported chatID type: %T", chatID)
	}

	switch chatID.(type) {
	case int64, int, int32, int16, int8:
		rawChat, err := ctx.app.GetInputPeer(parsedChatID)
		if err != nil {
			return 0, fmt.Errorf("failed to resolve peer: %w", err)
		}
		switch chat := rawChat.(type) {
		case *tg.InputPeerUser:
			parsedChatID = chat.UserID
		case *tg.InputPeerChat:
			parsedChatID = -chat.ChatID
		case *tg.InputPeerChannel:
			parsedChatID = -1000000000000 - chat.ChannelID
		}
	}
	return parsedChatID, nil
}
