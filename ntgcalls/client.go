package ntgcalls

import "sync"

type Client struct {
	ptr                         uintptr
	mu                          sync.RWMutex // Protects all callback slices
	connectionChangeCallbacks   []ConnectionChangeCallback
	streamEndCallbacks          []StreamEndCallback
	upgradeCallbacks            []UpgradeCallback
	signalCallbacks             []SignalCallback
	frameCallbacks              []FrameCallback
	remoteSourceCallbacks       []RemoteSourceCallback
	broadcastTimestampCallbacks []BroadcastTimestampCallback
	broadcastPartCallbacks      []BroadcastPartCallback
}

func (ctx *Client) OnStreamEnd(callback StreamEndCallback) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.streamEndCallbacks = append(ctx.streamEndCallbacks, callback)
}

func (ctx *Client) OnUpgrade(callback UpgradeCallback) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.upgradeCallbacks = append(ctx.upgradeCallbacks, callback)
}

func (ctx *Client) OnConnectionChange(callback ConnectionChangeCallback) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.connectionChangeCallbacks = append(
		ctx.connectionChangeCallbacks,
		callback,
	)
}

func (ctx *Client) OnSignal(callback SignalCallback) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.signalCallbacks = append(ctx.signalCallbacks, callback)
}

func (ctx *Client) OnFrame(callback FrameCallback) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.frameCallbacks = append(ctx.frameCallbacks, callback)
}

func (ctx *Client) OnRemoteSourceChange(callback RemoteSourceCallback) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.remoteSourceCallbacks = append(ctx.remoteSourceCallbacks, callback)
}

func (ctx *Client) OnRequestBroadcastTimestamp(
	callback BroadcastTimestampCallback,
) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.broadcastTimestampCallbacks = append(
		ctx.broadcastTimestampCallbacks,
		callback,
	)
}

func (ctx *Client) OnRequestBroadcastPart(callback BroadcastPartCallback) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.broadcastPartCallbacks = append(ctx.broadcastPartCallbacks, callback)
}

func (ctx *Client) getStreamEndCallbacks() []StreamEndCallback {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	if len(ctx.streamEndCallbacks) == 0 {
		return nil
	}
	cbs := make([]StreamEndCallback, len(ctx.streamEndCallbacks))
	copy(cbs, ctx.streamEndCallbacks)
	return cbs
}

func (ctx *Client) getUpgradeCallbacks() []UpgradeCallback {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	if len(ctx.upgradeCallbacks) == 0 {
		return nil
	}
	cbs := make([]UpgradeCallback, len(ctx.upgradeCallbacks))
	copy(cbs, ctx.upgradeCallbacks)
	return cbs
}

func (ctx *Client) getConnectionChangeCallbacks() []ConnectionChangeCallback {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	if len(ctx.connectionChangeCallbacks) == 0 {
		return nil
	}
	cbs := make([]ConnectionChangeCallback, len(ctx.connectionChangeCallbacks))
	copy(cbs, ctx.connectionChangeCallbacks)
	return cbs
}

func (ctx *Client) getSignalCallbacks() []SignalCallback {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	if len(ctx.signalCallbacks) == 0 {
		return nil
	}
	cbs := make([]SignalCallback, len(ctx.signalCallbacks))
	copy(cbs, ctx.signalCallbacks)
	return cbs
}

func (ctx *Client) getFrameCallbacks() []FrameCallback {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	if len(ctx.frameCallbacks) == 0 {
		return nil
	}
	cbs := make([]FrameCallback, len(ctx.frameCallbacks))
	copy(cbs, ctx.frameCallbacks)
	return cbs
}

func (ctx *Client) getRemoteSourceCallbacks() []RemoteSourceCallback {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	if len(ctx.remoteSourceCallbacks) == 0 {
		return nil
	}
	cbs := make([]RemoteSourceCallback, len(ctx.remoteSourceCallbacks))
	copy(cbs, ctx.remoteSourceCallbacks)
	return cbs
}

func (ctx *Client) getBroadcastTimestampCallbacks() []BroadcastTimestampCallback {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	if len(ctx.broadcastTimestampCallbacks) == 0 {
		return nil
	}
	cbs := make(
		[]BroadcastTimestampCallback,
		len(ctx.broadcastTimestampCallbacks),
	)
	copy(cbs, ctx.broadcastTimestampCallbacks)
	return cbs
}

func (ctx *Client) getBroadcastPartCallbacks() []BroadcastPartCallback {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	if len(ctx.broadcastPartCallbacks) == 0 {
		return nil
	}
	cbs := make([]BroadcastPartCallback, len(ctx.broadcastPartCallbacks))
	copy(cbs, ctx.broadcastPartCallbacks)
	return cbs
}
