package ubot

import (
	"fmt"

	tg "github.com/amarnathcjd/gogram/telegram"

	"yukkimusic/ntgcalls"
)

func (ctx *Context) Calls() (map[int64]ntgcalls.CallInfo, error) {
	return ctx.binding.Calls()
}

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
	if calls, err := ctx.binding.Calls(); err == nil {
		if _, ok := calls[chatID]; ok {
			return ctx.binding.SetStreamSources(
				chatID,
				ntgcalls.CaptureStream,
				mediaDescription,
			)
		}
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
		return ctx.addIncomingSources(chatID)
	}

	return nil
}

func (ctx *Context) Record(
	chatID int64,
	mediaDescription ntgcalls.MediaDescription,
) error {
	if calls, err := ctx.binding.Calls(); err == nil {
		if _, ok := calls[chatID]; ok {
			return ctx.binding.SetStreamSources(
				chatID,
				ntgcalls.PlaybackStream,
				mediaDescription,
			)
		}
	}

	return ctx.Play(chatID, ntgcalls.MediaDescription{})
}

func (ctx *Context) Stop(chatID int64) error {
	ctx.presentationsMutex.Lock()
	ctx.presentations = stdRemove(ctx.presentations, chatID)
	ctx.presentationsMutex.Unlock()

	ctx.callSourcesMutex.Lock()
	delete(ctx.callSources, chatID)
	ctx.callSourcesMutex.Unlock()

	ctx.waitConnectMutex.Lock()
	if waitChan, exists := ctx.waitConnect[chatID]; exists {
		select {
		case waitChan <- fmt.Errorf("call stopped"):
		default:
		}
		delete(ctx.waitConnect, chatID)
	}
	ctx.waitConnectMutex.Unlock()

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

func (ctx *Context) setCallStatus(
	call tg.InputGroupCall,
	state ntgcalls.MediaState,
) error {
	_, err := ctx.app.PhoneEditGroupCallParticipant(
		&tg.PhoneEditGroupCallParticipantParams{
			Call: call,
			Participant: &tg.InputPeerUser{
				UserID:     ctx.me.ID,
				AccessHash: ctx.me.AccessHash,
			},
			Muted:              state.Muted,
			VideoPaused:        state.VideoPaused,
			VideoStopped:       state.VideoStopped,
			PresentationPaused: state.PresentationPaused,
		},
	)
	return err
}
