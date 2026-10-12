package ubot

import "fmt"

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
