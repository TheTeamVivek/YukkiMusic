package ubot

import (
	"fmt"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func (ctx *Context) GetInputGroupCall(chatID int64) (tg.InputGroupCall, error) {
	ctx.inputGroupCallsMutex.RLock()
	call, ok := ctx.inputGroupCalls[chatID]
	ctx.inputGroupCallsMutex.RUnlock()

	if ok {
		if call == nil {
			return nil, fmt.Errorf("group call for chatID %d is closed", chatID)
		}
		return call, nil
	}

	peer, err := ctx.app.ResolvePeer(chatID)
	if err != nil {
		return nil, err
	}

	var retrievedCall tg.InputGroupCall

	switch chatPeer := peer.(type) {
	case *tg.InputPeerChannel:
		fullChat, err := ctx.app.ChannelsGetFullChannel(
			&tg.InputChannelObj{
				ChannelID:  chatPeer.ChannelID,
				AccessHash: chatPeer.AccessHash,
			},
		)
		if err != nil {
			return nil, err
		}
		retrievedCall = fullChat.FullChat.(*tg.ChannelFull).Call
	case *tg.InputPeerChat:
		fullChat, err := ctx.app.MessagesGetFullChat(chatPeer.ChatID)
		if err != nil {
			return nil, err
		}
		retrievedCall = fullChat.FullChat.(*tg.ChatFullObj).Call
	default:
		return nil, fmt.Errorf("chatID %d is not a group call", chatID)
	}

	ctx.inputGroupCallsMutex.Lock()
	ctx.inputGroupCalls[chatID] = retrievedCall
	ctx.inputGroupCallsMutex.Unlock()

	if retrievedCall == nil {
		return nil, fmt.Errorf("group call for chatID %d is closed", chatID)
	}
	return retrievedCall, nil
}

func (ctx *Context) convertCallID(callID int64) (int64, error) {
	ctx.inputCallsMutex.RLock()
	defer ctx.inputCallsMutex.RUnlock()

	for chatID, inputCall := range ctx.inputCalls {
		if inputCall.ID == callID {
			return chatID, nil
		}
	}
	return 0, fmt.Errorf("call id %d not found", callID)
}

func (ctx *Context) convertGroupCallID(callID int64) (int64, error) {
	ctx.inputGroupCallsMutex.RLock()
	defer ctx.inputGroupCallsMutex.RUnlock()

	for chatID, inputCallInterface := range ctx.inputGroupCalls {
		if inputCall, ok := inputCallInterface.(*tg.InputGroupCallObj); ok {
			if inputCall.ID == callID {
				return chatID, nil
			}
		}
	}
	return 0, fmt.Errorf("group call id %d not found", callID)
}

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
