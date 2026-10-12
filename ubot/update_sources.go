package ubot

import (
	"slices"

	tg "github.com/amarnathcjd/gogram/telegram"
)

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
