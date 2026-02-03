package ubot

import (
	"fmt"
	"sync"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/ntgcalls"
)

type Context struct {
	binding *ntgcalls.Client
	app     *tg.Client
	self    *tg.UserObj

	mutedByAdminMutex sync.RWMutex
	mutedByAdmin      []int64

	presentationsMutex sync.RWMutex
	presentations      []int64

	pendingPresentationMutex sync.RWMutex
	pendingPresentation      map[int64]bool

	p2pConfigsMutex sync.RWMutex
	p2pConfigs      map[int64]*P2PConfig

	inputCallsMutex sync.RWMutex
	inputCalls      map[int64]*tg.InputPhoneCall

	inputGroupCallsMutex sync.RWMutex
	inputGroupCalls      map[int64]tg.InputGroupCall

	participantsMutex sync.Mutex
	callParticipants  map[int64]*CallParticipantsCache

	pendingConnectionsMutex sync.RWMutex
	pendingConnections      map[int64]*PendingConnection

	callSourcesMutex sync.RWMutex
	callSources      map[int64]*CallSources

	waitConnectMutex sync.RWMutex
	waitConnect      map[int64]chan error

	callbacksMutex        sync.RWMutex
	incomingCallCallbacks []func(client *Context, chatId int64)
	streamEndCallbacks    []ntgcalls.StreamEndCallback
	frameCallbacks        []ntgcalls.FrameCallback
}

func NewContext(app *tg.Client) *Context {
	client := &Context{
		binding: ntgcalls.NTgCalls(),
		app:     app,

		pendingPresentation: make(map[int64]bool),
		p2pConfigs:          make(map[int64]*P2PConfig),
		inputCalls:          make(map[int64]*tg.InputPhoneCall),
		inputGroupCalls:     make(map[int64]tg.InputGroupCall),
		pendingConnections:  make(map[int64]*PendingConnection),
		callParticipants:    make(map[int64]*CallParticipantsCache),
		callSources:         make(map[int64]*CallSources),
		waitConnect:         make(map[int64]chan error),
	}
	if app.IsConnected() {
		me := app.Me()

		if me.ID == 0 {
			var err error
			me, err = app.GetMe()
			if err != nil {
				gologging.Fatal(err)
			}
		}

		client.self = me
	}

	client.handleUpdates()
	return client
}

func (ctx *Context) Calls() map[int64]*ntgcalls.CallInfo {
	return ctx.binding.Calls()
}

func (ctx *Context) Mute(chatID int64) (bool, error) {
	return ctx.binding.Mute(chatID)
}

func (ctx *Context) Pause(chatID int64) (bool, error) {
	return ctx.binding.Pause(chatID)
}

func (ctx *Context) Resume(chatID int64) (bool, error) {
	return ctx.binding.Resume(chatID)
}

func (ctx *Context) Unmute(chatID int64) (bool, error) {
	return ctx.binding.Unmute(chatID)
}

func (ctx *Context) Play(
	chatID int64,
	mediaDescription ntgcalls.MediaDescription,
) error {
	if ctx.binding.Calls()[chatID] != nil {
		return ctx.binding.SetStreamSources(
			chatID,
			ntgcalls.CaptureStream,
			mediaDescription,
		)
	}

	err := ctx.connectCall(chatID, mediaDescription, "")
	if err != nil {
		return err
	}

	if chatID < 0 {
		err = ctx.joinPresentation(chatID, mediaDescription.Screen != nil)
		if err != nil {
			return err
		}
		return ctx.updateSources(chatID)
	}

	return nil
}

func (ctx *Context) Record(
	chatID int64,
	mediaDescription ntgcalls.MediaDescription,
) error {
	if ctx.binding.Calls()[chatID] != nil {
		return ctx.binding.SetStreamSources(
			chatID,
			ntgcalls.PlaybackStream,
			mediaDescription,
		)
	}

	return ctx.Play(chatID, ntgcalls.MediaDescription{})
}

func (ctx *Context) Stop(chatID int64) error {
	// Clean up presentations
	ctx.presentationsMutex.Lock()
	ctx.presentations = stdRemove(ctx.presentations, chatID)
	ctx.presentationsMutex.Unlock()

	ctx.callSourcesMutex.Lock()
	delete(ctx.callSources, chatID)
	ctx.callSourcesMutex.Unlock()

	ctx.waitConnectMutex.Lock()
	if waitChan, exists := ctx.waitConnect[chatID]; exists {
		select {
		case waitChan <- fmt.Errorf("call stopped"):
		default:
		}
		delete(ctx.waitConnect, chatID)
	}
	ctx.waitConnectMutex.Unlock()

	err := ctx.binding.Stop(chatID)
	if err != nil {
		return err
	}

	ctx.inputGroupCallsMutex.RLock()
	inputGroupCall, ok := ctx.inputGroupCalls[chatID]
	ctx.inputGroupCallsMutex.RUnlock()

	if !ok {
		return nil
	}

	_, err = ctx.app.PhoneLeaveGroupCall(inputGroupCall, 0)
	if err != nil {
		return err
	}
	return nil
}

func (ctx *Context) OnIncomingCall(
	callback func(client *Context, chatId int64),
) {
	ctx.callbacksMutex.Lock()
	defer ctx.callbacksMutex.Unlock()
	ctx.incomingCallCallbacks = append(ctx.incomingCallCallbacks, callback)
}

func (ctx *Context) OnStreamEnd(callback ntgcalls.StreamEndCallback) {
	ctx.callbacksMutex.Lock()
	defer ctx.callbacksMutex.Unlock()
	ctx.streamEndCallbacks = append(ctx.streamEndCallbacks, callback)
}

func (ctx *Context) OnFrame(callback ntgcalls.FrameCallback) {
	ctx.callbacksMutex.Lock()
	defer ctx.callbacksMutex.Unlock()
	ctx.frameCallbacks = append(ctx.frameCallbacks, callback)
}

func (ctx *Context) Close() {
	if ctx.binding == nil {
		return
	}

	for chatId := range ctx.binding.Calls() {
		ctx.binding.Stop(chatId)
	}

	ctx.p2pConfigsMutex.Lock()
	ctx.p2pConfigs = nil
	ctx.p2pConfigsMutex.Unlock()

	ctx.inputCallsMutex.Lock()
	ctx.inputCalls = nil
	ctx.inputCallsMutex.Unlock()

	ctx.inputGroupCallsMutex.Lock()
	ctx.inputGroupCalls = nil
	ctx.inputGroupCallsMutex.Unlock()

	ctx.pendingConnectionsMutex.Lock()
	ctx.pendingConnections = nil
	ctx.pendingConnectionsMutex.Unlock()

	ctx.participantsMutex.Lock()
	ctx.callParticipants = nil
	ctx.participantsMutex.Unlock()

	ctx.callSourcesMutex.Lock()
	ctx.callSources = nil
	ctx.callSourcesMutex.Unlock()

	ctx.waitConnectMutex.Lock()
	ctx.waitConnect = nil
	ctx.waitConnectMutex.Unlock()

	ctx.binding.Free()
	ctx.binding = nil
}
