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

// getCallbacks returns a snapshot copy of a callback slice under the client
// lock. It is a generic method, a language feature introduced in Go 1.27.
func (c *Client) getCallbacks[T any](src *[]T) []T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(*src) == 0 {
		return nil
	}
	return append([]T(nil), (*src)...)
}

func (c *Client) getStreamEndCallbacks() []StreamEndCallback {
	return c.getCallbacks(&c.streamEndCallbacks)
}

func (c *Client) getUpgradeCallbacks() []UpgradeCallback {
	return c.getCallbacks(&c.upgradeCallbacks)
}

func (c *Client) getConnectionChangeCallbacks() []ConnectionChangeCallback {
	return c.getCallbacks(&c.connectionChangeCallbacks)
}

func (c *Client) getSignalCallbacks() []SignalCallback {
	return c.getCallbacks(&c.signalCallbacks)
}

func (c *Client) getFrameCallbacks() []FrameCallback {
	return c.getCallbacks(&c.frameCallbacks)
}

func (c *Client) getRemoteSourceCallbacks() []RemoteSourceCallback {
	return c.getCallbacks(&c.remoteSourceCallbacks)
}

func (c *Client) getBroadcastTimestampCallbacks() []BroadcastTimestampCallback {
	return c.getCallbacks(&c.broadcastTimestampCallbacks)
}

func (c *Client) getBroadcastPartCallbacks() []BroadcastPartCallback {
	return c.getCallbacks(&c.broadcastPartCallbacks)
}

func (c *Client) getEmojisCallbacks() []EmojisCallback {
	return c.getCallbacks(&c.emojisCallbacks)
}

func (c *Client) getRequestParticipantsCallbacks() []RequestParticipantsCallback {
	return c.getCallbacks(&c.requestParticipantsCallbacks)
}

func (c *Client) getOutboundBlockCallbacks() []OutboundBlockCallback {
	return c.getCallbacks(&c.outboundBlockCallbacks)
}

func (c *Client) getSubchainRequestCallbacks() []SubchainRequestCallback {
	return c.getCallbacks(&c.subchainRequestCallbacks)
}
