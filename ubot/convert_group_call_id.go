package ubot

import (
	"fmt"

	tg "github.com/amarnathcjd/gogram/telegram"
)

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
