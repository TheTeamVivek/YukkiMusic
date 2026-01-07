package ubot

func (ctx *Context) Stop(chatId any) error {
	parsedChatId, err := ctx.parseChatId(chatId)
	if err != nil {
		return err
	}

	ctx.presentationsMutex.Lock()
	ctx.presentations = stdRemove(ctx.presentations, parsedChatId)
	ctx.presentationsMutex.Unlock()

	ctx.pendingPresentationMutex.Lock()
	delete(ctx.pendingPresentation, parsedChatId)
	ctx.pendingPresentationMutex.Unlock()

	ctx.callSourcesMutex.Lock()
	delete(ctx.callSources, parsedChatId)
	ctx.callSourcesMutex.Unlock()

	err := ctx.binding.Stop(parsedChatId)

	if err != nil {
		return err
	}

	ctx.inputGroupCallsMutex.RLock()
	inputGroupCall, ok := ctx.inputGroupCalls[parsedChatId]
	ctx.inputGroupCallsMutex.RUnlock()

	_, err = ctx.app.PhoneLeaveGroupCall(inputGroupCall, 0)

	if err != nil {
		return err
	}
	return nil
}
