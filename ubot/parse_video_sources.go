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
		var sourceIDs []uint32
		for _, sourceID := range source.Sources {
			sourceIDs = append(sourceIDs, uint32(sourceID))
		}
		ssrcGroups = append(ssrcGroups, ntgcalls.SSRCGroup{
			Semantics: source.Semantics,
			SSRCs:     sourceIDs,
		})
	}
	return ssrcGroups
}
