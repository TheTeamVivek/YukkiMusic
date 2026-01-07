package ubot

import "main/ntgcalls"

func (ctx *Context) Mute(chatID int64) (bool, error) {
	return ctx.binding.Mute(chatID)
}

func (ctx *Context) Pause(chatID int64) (bool, error) {
	return ctx.binding.Pause(chatID)
}

func (ctx *Context) Resume(chatID int64) (bool, error) {
	return ctx.binding.Resume(chatID)
}

func (ctx *Context) Unmute(chatID int64) (bool, error) {
	return ctx.binding.Unmute(chatID)
}

func (ctx *Context) Play(
	chatID int64,
	mediaDescription ntgcalls.MediaDescription,
) error {
	if ctx.binding.Calls()[chatID] != nil {
		return ctx.binding.SetStreamSources(
			chatID,
			ntgcalls.CaptureStream,
			mediaDescription,
		)
	}

	err := ctx.connectCall(chatID, mediaDescription, "")
	if err != nil {
		return err
	}

	if chatID < 0 {

		err = ctx.joinPresentation(chatID, mediaDescription.Screen != nil)
		if err != nil {
			return err
		}
		return ctx.updateSources(chatID)

	}
	return nil
}

func (ctx *Context) Record(
	chatID int64,
	mediaDescription ntgcalls.MediaDescription,
) error {
	if ctx.binding.Calls()[chatID] != nil {
		return ctx.binding.SetStreamSources(
			chatID,
			ntgcalls.PlaybackStream,
			mediaDescription,
		)
	}

	return ctx.Play(chatID, ntgcalls.MediaDescription{})
}

func (ctx *Context) Stop(chatID int64) error {
	ctx.presentationsMutex.Lock()
	ctx.presentations = stdRemove(ctx.presentations, chatID)
	ctx.presentationsMutex.Unlock()

	ctx.pendingPresentationMutex.Lock()
	delete(ctx.pendingPresentation, chatID)
	ctx.pendingPresentationMutex.Unlock()

	ctx.callSourcesMutex.Lock()
	delete(ctx.callSources, chatID)
	ctx.callSourcesMutex.Unlock()

	err := ctx.binding.Stop(chatID)
	if err != nil {
		return err
	}

	ctx.inputGroupCallsMutex.RLock()
	inputGroupCall, ok := ctx.inputGroupCalls[chatID]
	ctx.inputGroupCallsMutex.RUnlock()

	if !ok {
		return nil
	}

	_, err = ctx.app.PhoneLeaveGroupCall(inputGroupCall, 0)
	if err != nil {
		return err
	}
	return nil
}
