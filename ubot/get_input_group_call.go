package ubot

import (
	"fmt"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func (ctx *Context) GetInputGroupCall(chatId int64) (tg.InputGroupCall, error) {
	ctx.inputGroupCallsMutex.RLock()
	call, ok := ctx.inputGroupCalls[chatId]
	ctx.inputGroupCallsMutex.RUnlock()

	if ok {
		if call == nil {
			return nil, fmt.Errorf("group call for chatId %d is closed", chatId)
		}
		return call, nil
	}

	peer, err := ctx.app.ResolvePeer(chatId)
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
		return nil, fmt.Errorf("chatId %d is not a group call", chatId)
	}

	ctx.inputGroupCallsMutex.Lock()
	ctx.inputGroupCalls[chatId] = retrievedCall
	ctx.inputGroupCallsMutex.Unlock()

	if retrievedCall == nil {
		return nil, fmt.Errorf("group call for chatId %d is closed", chatId)
	}
	return retrievedCall, nil
}
