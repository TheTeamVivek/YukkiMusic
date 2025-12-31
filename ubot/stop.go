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

	stopErr := ctx.binding.Stop(parsedChatId)

	ctx.inputGroupCallsMutex.RLock()
	inputGroupCall, ok := ctx.inputGroupCalls[parsedChatId]
	ctx.inputGroupCallsMutex.RUnlock()

	if !ok { return stopErr }
	_, leaveErr := ctx.app.PhoneLeaveGroupCall(inputGroupCall, 0)
	if stopErr != nil {
		return stopErr
	}
	if leaveErr != nil {
		return leaveErr
	}
	return nil
}
