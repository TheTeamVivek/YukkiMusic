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
