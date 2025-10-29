package ubot

import (
	"fmt"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func (ctx *Context) parseChatId(chatId any) (int64, error) {
	var parsedChatId int64
	switch v := chatId.(type) {
	case int64:
		parsedChatId = v
	case int:
		parsedChatId = int64(v)
	case int32:
		parsedChatId = int64(v)
	case int16:
		parsedChatId = int64(v)
	case int8:
		parsedChatId = int64(v)
	case string:
		rawChat, err := ctx.app.ResolveUsername(chatId.(string))
		if err != nil {
			return 0, fmt.Errorf("failed to resolve username: %w", err)
		}
		switch chat := rawChat.(type) {
		case *tg.UserObj:
			parsedChatId = chat.ID
		case *tg.ChatObj:
			parsedChatId = -chat.ID
		case *tg.Channel:
			parsedChatId = -1000000000000 - chat.ID
		}
	default:
		return 0, fmt.Errorf("unsupported chatId type: %T", chatId)
	}

	switch chatId.(type) {
	case int64, int, int32, int16, int8:
		rawChat, err := ctx.app.GetInputPeer(parsedChatId)
		if err != nil {
			return 0, fmt.Errorf("failed to resolve peer: %w", err)
		}
		switch chat := rawChat.(type) {
		case *tg.InputPeerUser:
			parsedChatId = chat.UserID
		case *tg.InputPeerChat:
			parsedChatId = -chat.ChatID
		case *tg.InputPeerChannel:
			parsedChatId = -1000000000000 - chat.ChannelID
		}
	}
	return parsedChatId, nil
}
