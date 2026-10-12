package ubot

import (
	"maps"
	"slices"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func (ctx *Context) GetParticipants(
	chatID int64,
) ([]*tg.GroupCallParticipant, error) {
	ctx.participantsMutex.Lock()
	defer ctx.participantsMutex.Unlock()
	if ctx.callParticipants[chatID] == nil {
		ctx.callParticipants[chatID] = &CallParticipantsCache{
			Participants: make(map[int64]*tg.GroupCallParticipant),
		}
	}
	if time.Since(
		ctx.callParticipants[chatID].lastUpdate,
	) > time.Minute {
		groupCall, err := ctx.GetInputGroupCall(chatID)
		if err != nil {
			return nil, err
		}
		ctx.callParticipants[chatID].Participants = make(
			map[int64]*tg.GroupCallParticipant,
		)
		var nextOffset string
		for {
			res, err := ctx.app.PhoneGetGroupParticipants(
				groupCall,
				[]tg.InputPeer{},
				[]int32{},
				nextOffset,
				0,
			)
			if err != nil {
				return nil, err
			}
			for _, participant := range res.Participants {
				ctx.callParticipants[chatID].Participants[getParticipantID(participant.Peer)] = participant
			}
			if res.NextOffset == "" {
				break
			}
			nextOffset = res.NextOffset
		}
		ctx.callParticipants[chatID].lastUpdate = time.Now()
	}
	return slices.Collect(
		maps.Values(ctx.callParticipants[chatID].Participants),
	), nil
}
