package ntgcalls

//#include "ntgcalls.h"
import "C"
import "sync"

type Client struct {
	handle                       *C.ntg_instance
	mu                           sync.RWMutex // Protects all callback slices
	connectionChangeCallbacks    []ConnectionChangeCallback
	streamEndCallbacks           []StreamEndCallback
	upgradeCallbacks             []UpgradeCallback
	signalCallbacks              []SignalCallback
	frameCallbacks               []FrameCallback
	remoteSourceCallbacks        []RemoteSourceCallback
	broadcastTimestampCallbacks  []BroadcastTimestampCallback
	broadcastPartCallbacks       []BroadcastPartCallback
	emojisCallbacks              []EmojisCallback
	requestParticipantsCallbacks []RequestParticipantsCallback
	outboundBlockCallbacks       []OutboundBlockCallback
	subchainRequestCallbacks     []SubchainRequestCallback
}

func (c *Client) OnStreamEnd(callback StreamEndCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.streamEndCallbacks = append(c.streamEndCallbacks, callback)
}

func (c *Client) OnUpgrade(callback UpgradeCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.upgradeCallbacks = append(c.upgradeCallbacks, callback)
}

func (c *Client) OnConnectionChange(callback ConnectionChangeCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connectionChangeCallbacks = append(c.connectionChangeCallbacks, callback)
}

func (c *Client) OnSignal(callback SignalCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.signalCallbacks = append(c.signalCallbacks, callback)
}

func (c *Client) OnFrame(callback FrameCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.frameCallbacks = append(c.frameCallbacks, callback)
}

func (c *Client) OnRemoteSourceChange(callback RemoteSourceCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.remoteSourceCallbacks = append(c.remoteSourceCallbacks, callback)
}

func (c *Client) OnRequestBroadcastTimestamp(callback BroadcastTimestampCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.broadcastTimestampCallbacks = append(c.broadcastTimestampCallbacks, callback)
}

func (c *Client) OnRequestBroadcastPart(callback BroadcastPartCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.broadcastPartCallbacks = append(c.broadcastPartCallbacks, callback)
}

func (c *Client) OnUpdateEmojis(callback EmojisCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.emojisCallbacks = append(c.emojisCallbacks, callback)
}

func (c *Client) OnRequestParticipants(callback RequestParticipantsCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestParticipantsCallbacks = append(c.requestParticipantsCallbacks, callback)
}

func (c *Client) OnOutboundBlock(callback OutboundBlockCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.outboundBlockCallbacks = append(c.outboundBlockCallbacks, callback)
}

func (c *Client) OnSubchainRequest(callback SubchainRequestCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subchainRequestCallbacks = append(c.subchainRequestCallbacks, callback)
}

func (c *Client) getStreamEndCallbacks() []StreamEndCallback {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.streamEndCallbacks) == 0 {
		return nil
	}
	return append([]StreamEndCallback(nil), c.streamEndCallbacks...)
}

func (c *Client) getUpgradeCallbacks() []UpgradeCallback {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.upgradeCallbacks) == 0 {
		return nil
	}
	return append([]UpgradeCallback(nil), c.upgradeCallbacks...)
}

func (c *Client) getConnectionChangeCallbacks() []ConnectionChangeCallback {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.connectionChangeCallbacks) == 0 {
		return nil
	}
	return append([]ConnectionChangeCallback(nil), c.connectionChangeCallbacks...)
}

func (c *Client) getSignalCallbacks() []SignalCallback {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.signalCallbacks) == 0 {
		return nil
	}
	return append([]SignalCallback(nil), c.signalCallbacks...)
}

func (c *Client) getFrameCallbacks() []FrameCallback {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.frameCallbacks) == 0 {
		return nil
	}
	return append([]FrameCallback(nil), c.frameCallbacks...)
}

func (c *Client) getRemoteSourceCallbacks() []RemoteSourceCallback {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.remoteSourceCallbacks) == 0 {
		return nil
	}
	return append([]RemoteSourceCallback(nil), c.remoteSourceCallbacks...)
}

func (c *Client) getBroadcastTimestampCallbacks() []BroadcastTimestampCallback {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.broadcastTimestampCallbacks) == 0 {
		return nil
	}
	return append([]BroadcastTimestampCallback(nil), c.broadcastTimestampCallbacks...)
}

func (c *Client) getBroadcastPartCallbacks() []BroadcastPartCallback {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.broadcastPartCallbacks) == 0 {
		return nil
	}
	return append([]BroadcastPartCallback(nil), c.broadcastPartCallbacks...)
}

func (c *Client) getEmojisCallbacks() []EmojisCallback {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.emojisCallbacks) == 0 {
		return nil
	}
	return append([]EmojisCallback(nil), c.emojisCallbacks...)
}

func (c *Client) getRequestParticipantsCallbacks() []RequestParticipantsCallback {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.requestParticipantsCallbacks) == 0 {
		return nil
	}
	return append([]RequestParticipantsCallback(nil), c.requestParticipantsCallbacks...)
}

func (c *Client) getOutboundBlockCallbacks() []OutboundBlockCallback {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.outboundBlockCallbacks) == 0 {
		return nil
	}
	return append([]OutboundBlockCallback(nil), c.outboundBlockCallbacks...)
}

func (c *Client) getSubchainRequestCallbacks() []SubchainRequestCallback {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.subchainRequestCallbacks) == 0 {
		return nil
	}
	return append([]SubchainRequestCallback(nil), c.subchainRequestCallbacks...)
}
