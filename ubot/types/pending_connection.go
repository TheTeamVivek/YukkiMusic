package types

import "github.com/TheTeamVivek/YukkiMusic/ntgcalls"

type PendingConnection struct {
	MediaDescription ntgcalls.MediaDescription
	Payload          string
	Presentation     bool
}
