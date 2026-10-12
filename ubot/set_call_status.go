package ubot

import (
	tg "github.com/amarnathcjd/gogram/telegram"

	"yukkimusic/ntgcalls"
)

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
