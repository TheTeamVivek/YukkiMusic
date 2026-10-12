package ubot

import (
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"

	"yukkimusic/ntgcalls"
)

type (
	CallParticipantsCache struct {
		Participants map[int64]*tg.GroupCallParticipant
		lastUpdate   time.Time
	}

	CallSources struct {
		Camera, Screen map[int64]string
	}

	P2PConfig struct {
		DHConfig       *tg.MessagesDhConfigObj
		PhoneCall      *tg.PhoneCallObj
		IsOutgoing     bool
		KeyFingerprint int64
		GAorB          []byte
		WaitData       chan error
	}

	PendingConnection struct {
		MediaDescription ntgcalls.MediaDescription
		Payload          string
		Presentation     bool
	}
)
