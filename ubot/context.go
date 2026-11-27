package ubot

import (
	"fmt"
	"sync"

	tg "github.com/xAmarnath/gogram/telegram"

	"yukkimusic/ntgcalls"
)

type Context struct {
	binding *ntgcalls.Client
	app     *tg.Client
	me      *tg.UserObj

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
	incomingCallCallbacks []func(client *Context, chatID int64)
	streamEndCallbacks    []ntgcalls.StreamEndCallback
	frameCallbacks        []ntgcalls.FrameCallback
	emojisCallbacks       []ntgcalls.EmojisCallback
	callMessageCallbacks  []CallMessageCallback

	messagesMutex sync.Mutex
	callMessages  map[int64]*CallMessagesCache
}

func NewContext(app *tg.Client) (*Context, error) {
	client := &Context{
		binding: ntgcalls.NewClient(),
		app:     app,

		pendingPresentation: make(map[int64]bool),
		p2pConfigs:          make(map[int64]*P2PConfig),
		inputCalls:          make(map[int64]*tg.InputPhoneCall),
		inputGroupCalls:     make(map[int64]tg.InputGroupCall),
		pendingConnections:  make(map[int64]*PendingConnection),
		callParticipants:    make(map[int64]*CallParticipantsCache),
		callSources:         make(map[int64]*CallSources),
		waitConnect:         make(map[int64]chan error),
		callMessages:        make(map[int64]*CallMessagesCache),
	}
	if app.IsConnected() {
		me := app.Me()

		if me.ID == 0 {
			var err error
			me, err = app.GetMe()
			if err != nil {
				return nil, fmt.Errorf("fetching own identity: %w", err)
			}
		}

		client.me = me
	}

	client.registerHandlers()
	return client, nil
}

func (ctx *Context) OnIncomingCall(
	callback func(client *Context, chatID int64),
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

func (ctx *Context) OnUpdateEmojis(callback ntgcalls.EmojisCallback) {
	ctx.callbacksMutex.Lock()
	defer ctx.callbacksMutex.Unlock()
	ctx.emojisCallbacks = append(ctx.emojisCallbacks, callback)
}

func (ctx *Context) OnCallMessage(callback CallMessageCallback) {
	ctx.callbacksMutex.Lock()
	defer ctx.callbacksMutex.Unlock()
	ctx.callMessageCallbacks = append(ctx.callMessageCallbacks, callback)
}

func (ctx *Context) getSelf() *tg.UserObj {
	if ctx.me != nil && ctx.me.ID != 0 {
		return ctx.me
	}
	if ctx.app != nil && ctx.app.IsConnected() {
		if me := ctx.app.Me(); me != nil && me.ID != 0 {
			ctx.me = me
			return me
		}
	}
	return nil
}

func (ctx *Context) selfInputPeer() tg.InputPeer {
	if me := ctx.getSelf(); me != nil {
		return &tg.InputPeerUser{
			UserID:     me.ID,
			AccessHash: me.AccessHash,
		}
	}
	return &tg.InputPeerUser{}
}

func (ctx *Context) Close() {
	if ctx.binding == nil {
		return
	}

	if calls, err := ctx.binding.Calls(); err == nil {
		for chatID := range calls {
			ctx.binding.Stop(chatID)
		}
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

	ctx.messagesMutex.Lock()
	ctx.callMessages = nil
	ctx.messagesMutex.Unlock()

	ctx.binding.Free()
	ctx.binding = nil
}
