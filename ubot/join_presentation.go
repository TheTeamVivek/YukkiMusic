package ubot

import (
	"slices"

	tg "github.com/amarnathcjd/gogram/telegram"

	"yukkimusic/ntgcalls"
)

func (ctx *Context) joinPresentation(chatID int64, join bool) error {
	defer func() {
		ctx.waitConnectMutex.Lock()
		delete(ctx.waitConnect, chatID)
		ctx.waitConnectMutex.Unlock()
	}()

	connectionMode, err := ctx.binding.GetConnectionMode(chatID)
	if err != nil {
		return err
	}

	if connectionMode == ntgcalls.StreamConnection {
		ctx.pendingConnectionsMutex.Lock()
		if ctx.pendingConnections[chatID] != nil {
			ctx.pendingConnections[chatID].Presentation = join
		}
		ctx.pendingConnectionsMutex.Unlock()
		return nil
	}

	if connectionMode != ntgcalls.RTCConnection {
		return nil
	}

	if join {
		ctx.presentationsMutex.Lock()
		alreadyPresenting := slices.Contains(ctx.presentations, chatID)
		ctx.presentationsMutex.Unlock()

		if alreadyPresenting {
			return nil
		}

		ctx.waitConnectMutex.Lock()
		waitChan := make(chan error)
		ctx.waitConnect[chatID] = waitChan
		ctx.waitConnectMutex.Unlock()

		jsonParams, err := ctx.binding.InitPresentation(chatID)
		if err != nil {
			return err
		}

		ctx.inputGroupCallsMutex.RLock()
		inputGroupCall := ctx.inputGroupCalls[chatID]
		ctx.inputGroupCallsMutex.RUnlock()

		resultParams := "{\"transport\": null}"
		callResRaw, err := ctx.app.PhoneJoinGroupCallPresentation(
			inputGroupCall,
			&tg.DataJson{
				Data: jsonParams,
			},
		)
		if err != nil {
			return err
		}

		callRes := callResRaw.(*tg.UpdatesObj)
		for _, u := range callRes.Updates {
			switch update := u.(type) {
			case *tg.UpdateGroupCallConnection:
				resultParams = update.Params.Data
			}
		}

		err = ctx.binding.Connect(chatID, resultParams, true)
		if err != nil {
			return err
		}

		<-waitChan

		ctx.presentationsMutex.Lock()
		ctx.presentations = append(ctx.presentations, chatID)
		ctx.presentationsMutex.Unlock()

	} else {
		// Leave presentation
		ctx.presentationsMutex.Lock()
		isPresenting := slices.Contains(ctx.presentations, chatID)
		if isPresenting {
			ctx.presentations = stdRemove(ctx.presentations, chatID)
		}
		ctx.presentationsMutex.Unlock()

		if !isPresenting {
			return nil
		}

		err = ctx.binding.StopPresentation(chatID)
		if err != nil {
			return err
		}

		ctx.inputGroupCallsMutex.RLock()
		inputGroupCall := ctx.inputGroupCalls[chatID]
		ctx.inputGroupCallsMutex.RUnlock()

		_, err = ctx.app.PhoneLeaveGroupCallPresentation(inputGroupCall)
		if err != nil {
			return err
		}
	}

	return nil
}
