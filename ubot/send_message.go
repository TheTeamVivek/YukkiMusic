package ubot

import (
	"crypto/rand"
	"encoding/binary"
	"time"

	tg "github.com/xAmarnath/gogram/telegram"
)

func randomMessageID() int64 {
	var buffer [8]byte
	_, _ = rand.Read(buffer[:])
	return int64(binary.LittleEndian.Uint64(buffer[:]))
}

// SendMessage sends an HTML-formatted message (and optional Stars tip) in an active group or conference call.
func (ctx *Context) SendMessage(chatID any, text string, stars int64) (*CallMessage, error) {
	parsedChatID, err := ctx.parseChatID(chatID)
	if err != nil {
		return nil, err
	}
	inputCall, err := ctx.GetInputGroupCall(parsedChatID)
	if err != nil {
		return nil, err
	}
	if text == "" && stars == 0 {
		return nil, nil
	}
	message, entities := parseHTML(text)
	messageRandomID := randomMessageID()
	if ctx.isConference(parsedChatID) {
		return ctx.sendConferenceMessage(parsedChatID, inputCall, messageRandomID, message, conferenceEntities(entities))
	}

	ctx.messagesMutex.Lock()
	ctx.getCallMessages(parsedChatID).SendingIDs[messageRandomID] = struct{}{}
	ctx.messagesMutex.Unlock()

	result, err := ctx.app.PhoneSendGroupCallMessage(
		&tg.PhoneSendGroupCallMessageParams{
			Call:     inputCall,
			RandomID: messageRandomID,
			Message: &tg.TextWithEntities{
				Text:     message,
				Entities: entities,
			},
			AllowPaidStars: stars,
			SendAs:         ctx.selfInputPeer(),
		},
	)
	if err != nil {
		ctx.messagesMutex.Lock()
		delete(ctx.getCallMessages(parsedChatID).SendingIDs, messageRandomID)
		ctx.messagesMutex.Unlock()
		return nil, err
	}

	var sent *CallMessage
	if updates, ok := result.(*tg.UpdatesObj); ok {
		for _, update := range updates.Updates {
			if u, ok := update.(*tg.UpdateGroupCallMessage); ok {
				sent = parseCallMessage(u.Message)
				if sent != nil {
					ctx.callMessageSent(messageRandomID, sent.MessageID)
				}
			}
		}
	}
	return sent, nil
}

func (ctx *Context) sendConferenceMessage(
	chatID int64,
	inputCall tg.InputGroupCall,
	messageRandomID int64,
	message string,
	entities []tg.MessageEntity,
) (*CallMessage, error) {
	serialized, err := serializeConferenceMessage(messageRandomID, message, entities)
	if err != nil {
		return nil, err
	}
	encrypted, err := ctx.binding.EncryptMessage(chatID, serialized)
	if err != nil {
		return nil, err
	}
	messageID := ctx.conferenceMessageSent(chatID, messageRandomID)
	if _, err = ctx.app.PhoneSendGroupCallEncryptedMessage(inputCall, encrypted); err != nil {
		return nil, err
	}
	var senderID int64
	if self := ctx.getSelf(); self != nil {
		senderID = self.ID
	}
	return &CallMessage{
		MessageID: messageID,
		SenderID:  senderID,
		Text:      message,
		HTML:      unparseHTML(message, entities),
		Date:      int32(time.Now().Unix()),
	}, nil
}
