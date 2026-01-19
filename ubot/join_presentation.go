package ubot

import (
	"slices"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/ntgcalls"
)

func (ctx *Context) joinPresentation(chatId int64, join bool) error {
	defer func() {
		ctx.waitConnectMutex.Lock()
		delete(ctx.waitConnect, chatId)
		ctx.waitConnectMutex.Unlock()
	}()

	connectionMode, err := ctx.binding.GetConnectionMode(chatId)
	if err != nil {
		return err
	}

	if connectionMode == ntgcalls.StreamConnection {
		ctx.pendingConnectionsMutex.Lock()
		if ctx.pendingConnections[chatId] != nil {
			ctx.pendingConnections[chatId].Presentation = join
		}
		ctx.pendingConnectionsMutex.Unlock()
		return nil
	}

	if connectionMode != ntgcalls.RtcConnection {
		return nil
	}

	if join {
		ctx.presentationsMutex.Lock()
		alreadyPresenting := slices.Contains(ctx.presentations, chatId)
		ctx.presentationsMutex.Unlock()

		if alreadyPresenting {
			return nil
		}

		ctx.waitConnectMutex.Lock()
		waitChan := make(chan error)
		ctx.waitConnect[chatId] = waitChan
		ctx.waitConnectMutex.Unlock()

		jsonParams, err := ctx.binding.InitPresentation(chatId)
		if err != nil {
			return err
		}

		ctx.inputGroupCallsMutex.RLock()
		inputGroupCall := ctx.inputGroupCalls[chatId]
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

		err = ctx.binding.Connect(chatId, resultParams, true)
		if err != nil {
			return err
		}

		<-waitChan

		ctx.presentationsMutex.Lock()
		ctx.presentations = append(ctx.presentations, chatId)
		ctx.presentationsMutex.Unlock()

	} else {
		// Leave presentation
		ctx.presentationsMutex.Lock()
		isPresenting := slices.Contains(ctx.presentations, chatId)
		if isPresenting {
			ctx.presentations = stdRemove(ctx.presentations, chatId)
		}
		ctx.presentationsMutex.Unlock()

		if !isPresenting {
			return nil
		}

		err = ctx.binding.StopPresentation(chatId)
		if err != nil {
			return err
		}

		ctx.inputGroupCallsMutex.RLock()
		inputGroupCall := ctx.inputGroupCalls[chatId]
		ctx.inputGroupCallsMutex.RUnlock()

		_, err = ctx.app.PhoneLeaveGroupCallPresentation(inputGroupCall)
		if err != nil {
			return err
		}
	}

	return nil
}
