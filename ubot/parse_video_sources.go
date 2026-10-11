package ubot

import (
	tg "github.com/amarnathcjd/gogram/telegram"

	"yukkimusic/ntgcalls"
)

func parseVideoSources(
	sources []*tg.GroupCallParticipantVideoSourceGroup,
) []ntgcalls.SSRCGroup {
	var ssrcGroups []ntgcalls.SSRCGroup
	for _, source := range sources {
		var sourceIds []uint32
		for _, sourceId := range source.Sources {
			sourceIds = append(sourceIds, uint32(sourceId))
		}
		ssrcGroups = append(ssrcGroups, ntgcalls.SSRCGroup{
			Semantics: source.Semantics,
			SSRCs:     sourceIds,
		})
	}
	return ssrcGroups
}
