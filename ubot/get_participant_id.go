package ubot

import tg "github.com/amarnathcjd/gogram/telegram"

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
