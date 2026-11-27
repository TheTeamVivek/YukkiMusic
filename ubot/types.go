package ubot

import (
	"time"

	tg "github.com/xAmarnath/gogram/telegram"

	"yukkimusic/ntgcalls"
)

const (
	CallMessageReceived CallMessageAction = 1 << iota
	CallMessageDeleted
)

type (
	CallMessageAction int

	CallMessage struct {
		MessageID int32
		SenderID  int64
		Text      string
		HTML      string
		Date      int32
		Stars     int64
		FromAdmin bool
	}

	CallMessagesCache struct {
		SendingIDs              map[int64]struct{}
		MessageIDs              map[int32]struct{}
		ConferenceIDs           map[int64]int32
		ConferenceAutoIncrement int32
	}

	CallMessageCallback func(chatID int64, action CallMessageAction, messageID int32, message *CallMessage)

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
