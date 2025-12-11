package ubot

import (
	"slices"

	tg "github.com/amarnathcjd/gogram/telegram"
)

type videoToAdd struct {
	participantId int64
	endpoint      string
	sourceGroups  []*tg.GroupCallParticipantVideoSourceGroup
	isCamera      bool
}

func (ctx *Context) updateSources(chatId int64) error {
	participants, err := ctx.GetParticipants(chatId)
	if err != nil {
		return err
	}

	var videosToAdd []videoToAdd
	var shouldAddToMutedByAdmin bool

	ctx.callSourcesMutex.Lock()
	if ctx.callSources[chatId] == nil {
		ctx.callSources[chatId] = &CallSources{
			CameraSources: make(map[int64]string),
			ScreenSources: make(map[int64]string),
		}
	}
	for _, participant := range participants {
		participantId := getParticipantId(participant.Peer)

		if participant.Video != nil && ctx.callSources[chatId].CameraSources[participantId] == "" {
			ctx.callSources[chatId].CameraSources[participantId] = participant.Video.Endpoint
			videosToAdd = append(videosToAdd, videoToAdd{
				participantId: participantId,
				endpoint:      participant.Video.Endpoint,
				sourceGroups:  participant.Video.SourceGroups,
				isCamera:      true,
			})
		}

		if participant.Presentation != nil && ctx.callSources[chatId].ScreenSources[participantId] == "" {
			ctx.callSources[chatId].ScreenSources[participantId] = participant.Presentation.Endpoint
			videosToAdd = append(videosToAdd, videoToAdd{
				participantId: participantId,
				endpoint:      participant.Presentation.Endpoint,
				sourceGroups:  participant.Presentation.SourceGroups,
				isCamera:      false,
			})
		}

		if participantId == ctx.self.ID && !participant.CanSelfUnmute {
			shouldAddToMutedByAdmin = true
		}
	}
	ctx.callSourcesMutex.Unlock()

	for _, video := range videosToAdd {
		_, err = ctx.binding.AddIncomingVideo(
			chatId,
			video.endpoint,
			parseVideoSources(video.sourceGroups),
		)
		if err != nil {
			return err
		}
	}

	if shouldAddToMutedByAdmin {
		ctx.mutedByAdminMutex.Lock()
		if !slices.Contains(ctx.mutedByAdmin, chatId) {
			ctx.mutedByAdmin = append(ctx.mutedByAdmin, chatId)
		}
		ctx.mutedByAdminMutex.Unlock()
	}

	return nil
}
