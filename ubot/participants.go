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

func getParticipantID(peer tg.Peer) int64 {
	var participantID int64
	switch chatObj := peer.(type) {
	case *tg.PeerUser:
		participantID = chatObj.UserID
	case *tg.PeerChannel:
		participantID = chatObj.ChannelID
	case *tg.PeerChat:
		participantID = chatObj.ChatID
	}
	return participantID
}

type pendingVideo struct {
	participantID int64
	endpoint      string
	sourceGroups  []*tg.GroupCallParticipantVideoSourceGroup
	isCamera      bool
}

func (ctx *Context) addIncomingSources(chatID int64) error {
	participants, err := ctx.GetParticipants(chatID)
	if err != nil {
		return err
	}

	var pendingVideos []pendingVideo
	var shouldAddToMutedByAdmin bool

	ctx.callSourcesMutex.Lock()
	if ctx.callSources[chatID] == nil {
		ctx.callSources[chatID] = &CallSources{
			Camera: make(map[int64]string),
			Screen: make(map[int64]string),
		}
	}
	for _, participant := range participants {
		participantID := getParticipantID(participant.Peer)

		if participant.Video != nil &&
			ctx.callSources[chatID].Camera[participantID] == "" {
			ctx.callSources[chatID].Camera[participantID] = participant.Video.Endpoint
			pendingVideos = append(pendingVideos, pendingVideo{
				participantID: participantID,
				endpoint:      participant.Video.Endpoint,
				sourceGroups:  participant.Video.SourceGroups,
				isCamera:      true,
			})
		}

		if participant.Presentation != nil &&
			ctx.callSources[chatID].Screen[participantID] == "" {
			ctx.callSources[chatID].Screen[participantID] = participant.Presentation.Endpoint
			pendingVideos = append(pendingVideos, pendingVideo{
				participantID: participantID,
				endpoint:      participant.Presentation.Endpoint,
				sourceGroups:  participant.Presentation.SourceGroups,
				isCamera:      false,
			})
		}

		if participantID == ctx.me.ID && !participant.CanSelfUnmute {
			shouldAddToMutedByAdmin = true
		}
	}
	ctx.callSourcesMutex.Unlock()

	for _, video := range pendingVideos {
		_, err = ctx.binding.AddIncomingVideo(
			chatID,
			video.participantID,
			video.endpoint,
			parseVideoSources(video.sourceGroups),
		)
		if err != nil {
			return err
		}
	}

	if shouldAddToMutedByAdmin {
		ctx.mutedByAdminMutex.Lock()
		if !slices.Contains(ctx.mutedByAdmin, chatID) {
			ctx.mutedByAdmin = append(ctx.mutedByAdmin, chatID)
		}
		ctx.mutedByAdminMutex.Unlock()
	}

	return nil
}
