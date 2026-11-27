package ubot

import (
	tg "github.com/xAmarnath/gogram/telegram"

	"yukkimusic/ntgcalls"
)

func (ctx *Context) getConferenceLastBlock(chatID int64) ([]byte, error) {
	inputCall, err := ctx.GetInputGroupCall(chatID)
	if err != nil {
		return nil, err
	}
	res, err := ctx.app.PhoneGetGroupCallChainBlocks(inputCall, 0, -1, 1)
	if err != nil {
		return nil, err
	}
	if updatesObj, ok := res.(*tg.UpdatesObj); ok {
		for _, update := range updatesObj.Updates {
			if chainBlocks, ok := update.(*tg.UpdateGroupCallChainBlocks); ok {
				if len(chainBlocks.Blocks) > 0 {
					return chainBlocks.Blocks[len(chainBlocks.Blocks)-1], nil
				}
			}
		}
	}
	return nil, nil
}

func (ctx *Context) sendConferenceCallBroadcast(chatID int64, block []byte) {
	inputCall, err := ctx.GetInputGroupCall(chatID)
	if err != nil {
		return
	}
	_, _ = ctx.app.PhoneSendConferenceCallBroadcast(inputCall, block)
}

func (ctx *Context) getSubchainBlocks(chatID int64, request ntgcalls.SubchainRequest) ([][]byte, int32, error) {
	inputCall, err := ctx.GetInputGroupCall(chatID)
	if err != nil {
		return nil, 0, err
	}
	res, err := ctx.app.PhoneGetGroupCallChainBlocks(
		inputCall,
		request.Subchain,
		request.Height,
		request.Limit,
	)
	if err != nil {
		return nil, 0, err
	}
	if updatesObj, ok := res.(*tg.UpdatesObj); ok {
		for _, update := range updatesObj.Updates {
			if chainBlocks, ok := update.(*tg.UpdateGroupCallChainBlocks); ok {
				return chainBlocks.Blocks, chainBlocks.NextOffset, nil
			}
		}
	}
	return nil, 0, nil
}

// JoinConference creates or joins an E2E conference call for the given chat.
func (ctx *Context) JoinConference(chatID any, mediaDescription ntgcalls.MediaDescription) error {
	parsedChatID, err := ctx.parseChatID(chatID)
	if err != nil {
		return err
	}
	return ctx.connectCall(parsedChatID, mediaDescription, "", true, nil)
}
