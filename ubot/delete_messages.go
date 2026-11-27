package ubot

import "fmt"

// DeleteMessages deletes specific in-call messages in a group call.
func (ctx *Context) DeleteMessages(chatID any, messageIDs []int32, reportSpam bool) error {
	parsedChatID, err := ctx.parseChatID(chatID)
	if err != nil {
		return err
	}
	if ctx.isConference(parsedChatID) {
		return fmt.Errorf("deleting messages is not supported in conference calls")
	}
	inputCall, err := ctx.GetInputGroupCall(parsedChatID)
	if err != nil {
		return err
	}
	_, err = ctx.app.PhoneDeleteGroupCallMessages(reportSpam, inputCall, messageIDs)
	return err
}

// DeleteParticipantMessages deletes all in-call messages sent by a participant in a group call.
func (ctx *Context) DeleteParticipantMessages(chatID any, participantID any, reportSpam bool) error {
	parsedChatID, err := ctx.parseChatID(chatID)
	if err != nil {
		return err
	}
	if ctx.isConference(parsedChatID) {
		return fmt.Errorf("deleting messages is not supported in conference calls")
	}
	inputCall, err := ctx.GetInputGroupCall(parsedChatID)
	if err != nil {
		return err
	}
	parsedParticipantID, err := ctx.parseChatID(participantID)
	if err != nil {
		return err
	}
	participant, err := ctx.app.ResolvePeer(parsedParticipantID)
	if err != nil {
		return err
	}
	_, err = ctx.app.PhoneDeleteGroupCallParticipantMessages(reportSpam, inputCall, participant)
	return err
}
