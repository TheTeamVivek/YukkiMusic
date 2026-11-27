package ubot

import (
	"fmt"
	"time"

	tg "github.com/xAmarnath/gogram/telegram"

	"yukkimusic/ntgcalls"
)

func (ctx *Context) getCallMessages(chatID int64) *CallMessagesCache {
	if ctx.callMessages == nil {
		ctx.callMessages = make(map[int64]*CallMessagesCache)
	}
	if ctx.callMessages[chatID] == nil {
		ctx.callMessages[chatID] = &CallMessagesCache{
			SendingIDs:    make(map[int64]struct{}),
			MessageIDs:    make(map[int32]struct{}),
			ConferenceIDs: make(map[int64]int32),
		}
	}
	return ctx.callMessages[chatID]
}

func (ctx *Context) isConference(chatID int64) bool {
	callType, err := ctx.binding.GetCallType(chatID)
	return err == nil && callType == ntgcalls.ConferenceCallType
}

func (ctx *Context) chatIDFromInputCall(call tg.InputGroupCall) (int64, bool) {
	inputCall, ok := call.(*tg.InputGroupCallObj)
	if !ok {
		return 0, false
	}
	chatID, err := ctx.convertGroupCallID(inputCall.ID)
	return chatID, err == nil
}

func (ctx *Context) emitCallMessage(chatID int64, action CallMessageAction, messageID int32, message *CallMessage) {
	ctx.callbacksMutex.RLock()
	callbacks := make([]CallMessageCallback, len(ctx.callMessageCallbacks))
	copy(callbacks, ctx.callMessageCallbacks)
	ctx.callbacksMutex.RUnlock()

	for _, callback := range callbacks {
		go callback(chatID, action, messageID, message)
	}
}

func (ctx *Context) callMessageSent(randomID int64, messageID int32) {
	ctx.messagesMutex.Lock()
	defer ctx.messagesMutex.Unlock()
	for _, cache := range ctx.callMessages {
		if _, ok := cache.SendingIDs[randomID]; ok {
			delete(cache.SendingIDs, randomID)
			cache.MessageIDs[messageID] = struct{}{}
			return
		}
	}
}

func (ctx *Context) conferenceMessageSent(chatID int64, randomID int64) int32 {
	ctx.messagesMutex.Lock()
	defer ctx.messagesMutex.Unlock()
	cache := ctx.getCallMessages(chatID)
	cache.ConferenceAutoIncrement++
	cache.ConferenceIDs[randomID] = cache.ConferenceAutoIncrement
	return cache.ConferenceAutoIncrement
}

func parseCallMessage(message *tg.GroupCallMessage) *CallMessage {
	if message == nil {
		return nil
	}
	var text string
	var entities []tg.MessageEntity
	if message.Message != nil {
		text = message.Message.Text
		entities = message.Message.Entities
	}
	return &CallMessage{
		MessageID: message.ID,
		SenderID:  getParticipantID(message.FromID),
		Text:      text,
		HTML:      unparseHTML(text, entities),
		Date:      message.Date,
		Stars:     message.PaidMessageStars,
		FromAdmin: message.FromAdmin,
	}
}

func (ctx *Context) callMessageReceived(chatID int64, message *tg.GroupCallMessage) {
	if message == nil || ctx.isConference(chatID) {
		return
	}
	var text string
	if message.Message != nil {
		text = message.Message.Text
	}
	if text == "" && message.PaidMessageStars == 0 {
		return
	}
	ctx.messagesMutex.Lock()
	cache := ctx.getCallMessages(chatID)
	if _, ok := cache.MessageIDs[message.ID]; ok {
		ctx.messagesMutex.Unlock()
		return
	}
	cache.MessageIDs[message.ID] = struct{}{}
	ctx.messagesMutex.Unlock()
	ctx.emitCallMessage(chatID, CallMessageReceived, message.ID, parseCallMessage(message))
}

func (ctx *Context) callMessagesDeleted(chatID int64, messageIDs []int32) {
	ctx.messagesMutex.Lock()
	cache := ctx.getCallMessages(chatID)
	var deleted []int32
	for _, messageID := range messageIDs {
		if _, ok := cache.MessageIDs[messageID]; ok {
			delete(cache.MessageIDs, messageID)
			deleted = append(deleted, messageID)
		}
	}
	ctx.messagesMutex.Unlock()
	for _, messageID := range deleted {
		ctx.emitCallMessage(chatID, CallMessageDeleted, messageID, nil)
	}
}

func (ctx *Context) encryptedCallMessageReceived(chatID int64, fromID tg.Peer, data []byte) {
	if !ctx.isConference(chatID) {
		return
	}
	userID := getParticipantID(fromID)
	decrypted, err := ctx.binding.DecryptMessage(chatID, userID, data)
	if err != nil {
		fmt.Printf("decrypt call message (%d, %d): %v\n", chatID, userID, err)
		return
	}
	randomID, text, entities, ok := deserializeConferenceMessage(decrypted)
	if !ok {
		return
	}
	ctx.messagesMutex.Lock()
	cache := ctx.getCallMessages(chatID)
	if _, exists := cache.ConferenceIDs[randomID]; exists {
		ctx.messagesMutex.Unlock()
		return
	}
	ctx.messagesMutex.Unlock()
	messageID := ctx.conferenceMessageSent(chatID, randomID)
	ctx.emitCallMessage(chatID, CallMessageReceived, messageID, &CallMessage{
		MessageID: messageID,
		SenderID:  userID,
		Text:      text,
		HTML:      unparseHTML(text, entities),
		Date:      int32(time.Now().Unix()),
	})
}

func (ctx *Context) handleCallMessageUpdate(update tg.Update) {
	switch u := update.(type) {
	case *tg.UpdateMessageID:
		ctx.callMessageSent(u.RandomID, u.ID)
	case *tg.UpdateGroupCallMessage:
		if chatID, ok := ctx.chatIDFromInputCall(u.Call); ok {
			ctx.callMessageReceived(chatID, u.Message)
		}
	case *tg.UpdateDeleteGroupCallMessages:
		if chatID, ok := ctx.chatIDFromInputCall(u.Call); ok {
			ctx.callMessagesDeleted(chatID, u.Messages)
		}
	case *tg.UpdateGroupCallEncryptedMessage:
		if chatID, ok := ctx.chatIDFromInputCall(u.Call); ok {
			ctx.encryptedCallMessageReceived(chatID, u.FromID, u.EncryptedMessage)
		}
	}
}
