package ubot

import (
	"fmt"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/ntgcalls"
)

func (ctx *Context) connectCall(chatId int64, mediaDescription ntgcalls.MediaDescription, jsonParams string) error {
	// Create wait channel and ensure cleanup
	ctx.waitConnectMutex.Lock()
	waitChan := make(chan error, 1) // Buffered to prevent goroutine leak
	ctx.waitConnect[chatId] = waitChan
	ctx.waitConnectMutex.Unlock()

	defer func() {
		ctx.waitConnectMutex.Lock()
		delete(ctx.waitConnect, chatId)
		ctx.waitConnectMutex.Unlock()
	}()

	// Helper to signal error and return
	signalError := func(err error) error {
		select {
		case waitChan <- err:
		default:
		}
		return err
	}

	if chatId >= 0 {
		// P2P call handling
		defer func() {
			ctx.p2pConfigsMutex.Lock()
			delete(ctx.p2pConfigs, chatId)
			ctx.p2pConfigsMutex.Unlock()
		}()

		// Get or create P2P config
		ctx.p2pConfigsMutex.Lock()
		p2pConfig := ctx.p2pConfigs[chatId]
		ctx.p2pConfigsMutex.Unlock()

		if p2pConfig == nil {
			var err error
			p2pConfig, err = ctx.getP2PConfigs(nil)
			if err != nil {
				return signalError(err)
			}
			ctx.p2pConfigsMutex.Lock()
			ctx.p2pConfigs[chatId] = p2pConfig
			ctx.p2pConfigsMutex.Unlock()
		}

		err := ctx.binding.CreateP2PCall(chatId)
		if err != nil {
			return signalError(err)
		}

		err = ctx.binding.SetStreamSources(chatId, ntgcalls.CaptureStream, mediaDescription)
		if err != nil {
			return signalError(err)
		}

		ctx.p2pConfigsMutex.Lock()
		dhConfig := ntgcalls.DhConfig{
			G:      p2pConfig.DhConfig.G,
			P:      p2pConfig.DhConfig.P,
			Random: p2pConfig.DhConfig.Random,
		}
		gaOrB := p2pConfig.GAorB
		ctx.p2pConfigsMutex.Unlock()

		newGAorB, err := ctx.binding.InitExchange(chatId, dhConfig, gaOrB)
		if err != nil {
			return signalError(err)
		}

		ctx.p2pConfigsMutex.Lock()
		p2pConfig.GAorB = newGAorB
		ctx.p2pConfigsMutex.Unlock()

		protocolRaw := ntgcalls.GetProtocol()
		protocol := &tg.PhoneCallProtocol{
			UdpP2P:          protocolRaw.UdpP2P,
			UdpReflector:    protocolRaw.UdpReflector,
			MinLayer:        protocolRaw.MinLayer,
			MaxLayer:        protocolRaw.MaxLayer,
			LibraryVersions: protocolRaw.Versions,
		}

		userId, err := ctx.app.GetSendableUser(chatId)
		if err != nil {
			return signalError(err)
		}

		ctx.inputCallsMutex.RLock()
		inputCall := ctx.inputCalls[chatId]
		ctx.inputCallsMutex.RUnlock()

		ctx.p2pConfigsMutex.RLock()
		isOutgoing := p2pConfig.IsOutgoing
		gaOrBHash := p2pConfig.GAorB
		ctx.p2pConfigsMutex.RUnlock()

		if isOutgoing {
			_, err = ctx.app.PhoneRequestCall(
				&tg.PhoneRequestCallParams{
					Protocol: protocol,
					UserID:   userId,
					GAHash:   gaOrBHash,
					RandomID: int32(tg.GenRandInt()),
					Video:    mediaDescription.Camera != nil || mediaDescription.Screen != nil,
				},
			)
			if err != nil {
				return signalError(err)
			}
		} else {
			_, err = ctx.app.PhoneAcceptCall(
				inputCall,
				gaOrBHash,
				protocol,
			)
			if err != nil {
				return signalError(err)
			}
		}

		select {
		case err = <-p2pConfig.WaitData:
			if err != nil {
				return signalError(err)
			}
		case <-time.After(10 * time.Second):
			return signalError(fmt.Errorf("timed out waiting for an answer"))
		}

		ctx.p2pConfigsMutex.RLock()
		gaOrB = p2pConfig.GAorB
		fingerprint := p2pConfig.KeyFingerprint
		ctx.p2pConfigsMutex.RUnlock()

		res, err := ctx.binding.ExchangeKeys(chatId, gaOrB, fingerprint)
		if err != nil {
			return signalError(err)
		}

		ctx.p2pConfigsMutex.RLock()
		isOutgoing = p2pConfig.IsOutgoing
		ctx.p2pConfigsMutex.RUnlock()

		if isOutgoing {
			confirmRes, err := ctx.app.PhoneConfirmCall(
				inputCall,
				res.GAOrB,
				res.KeyFingerprint,
				protocol,
			)
			if err != nil {
				return signalError(err)
			}
			ctx.p2pConfigsMutex.Lock()
			p2pConfig.PhoneCall = confirmRes.PhoneCall.(*tg.PhoneCallObj)
			ctx.p2pConfigsMutex.Unlock()
		}

		ctx.p2pConfigsMutex.RLock()
		phoneCall := p2pConfig.PhoneCall
		ctx.p2pConfigsMutex.RUnlock()

		err = ctx.binding.ConnectP2P(
			chatId,
			parseRTCServers(phoneCall.Connections),
			phoneCall.Protocol.LibraryVersions,
			phoneCall.P2PAllowed,
		)
		if err != nil {
			return signalError(err)
		}

	} else {
		// Group call handling
		var err error
		jsonParams, err = ctx.binding.CreateCall(chatId)
		if err != nil {
			ctx.binding.Stop(chatId)
			return signalError(err)
		}

		err = ctx.binding.SetStreamSources(chatId, ntgcalls.CaptureStream, mediaDescription)
		if err != nil {
			ctx.binding.Stop(chatId)
			return signalError(err)
		}

		inputGroupCall, err := ctx.GetInputGroupCall(chatId)
		if err != nil {
			ctx.binding.Stop(chatId)
			return signalError(err)
		}

		resultParams := "{\"transport\": null}"
		callResRaw, err := ctx.app.PhoneJoinGroupCall(
			&tg.PhoneJoinGroupCallParams{
				Muted:        false,
				VideoStopped: mediaDescription.Camera == nil,
				Call:         inputGroupCall,
				Params: &tg.DataJson{
					Data: jsonParams,
				},
				JoinAs: &tg.InputPeerUser{
					UserID:     ctx.self.ID,
					AccessHash: ctx.self.AccessHash,
				},
			},
		)
		if err != nil {
			ctx.binding.Stop(chatId)
			return signalError(err)
		}

		callRes := callResRaw.(*tg.UpdatesObj)
		for _, u := range callRes.Updates {
			switch update := u.(type) {
			case *tg.UpdateGroupCallConnection:
				resultParams = update.Params.Data
			}
		}

		err = ctx.binding.Connect(chatId, resultParams, false)
		if err != nil {
			return signalError(err)
		}

		connectionMode, err := ctx.binding.GetConnectionMode(chatId)
		if err != nil {
			return signalError(err)
		}

		if connectionMode == ntgcalls.StreamConnection && len(jsonParams) > 0 {
			ctx.pendingConnectionsMutex.Lock()
			ctx.pendingConnections[chatId] = &PendingConnection{
				MediaDescription: mediaDescription,
				Payload:          jsonParams,
			}
			ctx.pendingConnectionsMutex.Unlock()
		}
	}

	return <-waitChan
}
