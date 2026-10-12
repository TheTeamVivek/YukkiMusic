package ubot

import (
	"fmt"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"

	"yukkimusic/ntgcalls"
)

func (ctx *Context) connectCall(
	chatID int64,
	mediaDescription ntgcalls.MediaDescription,
	jsonParams string,
) error {
	ctx.waitConnectMutex.Lock()
	waitChan := make(chan error)
	ctx.waitConnect[chatID] = waitChan
	ctx.waitConnectMutex.Unlock()

	defer func() {
		ctx.waitConnectMutex.Lock()
		delete(ctx.waitConnect, chatID)
		ctx.waitConnectMutex.Unlock()
	}()

	if chatID >= 0 {
		// P2P call handling
		defer func() {
			ctx.p2pConfigsMutex.Lock()
			delete(ctx.p2pConfigs, chatID)
			ctx.p2pConfigsMutex.Unlock()
		}()

		// Get or create P2P config
		ctx.p2pConfigsMutex.Lock()
		p2pConfig := ctx.p2pConfigs[chatID]
		ctx.p2pConfigsMutex.Unlock()

		if p2pConfig == nil {
			var err error
			p2pConfig, err = ctx.newP2PConfig(nil)
			if err != nil {
				return err
			}
			ctx.p2pConfigsMutex.Lock()
			ctx.p2pConfigs[chatID] = p2pConfig
			ctx.p2pConfigsMutex.Unlock()
		}

		err := ctx.binding.CreateP2PCall(chatID)
		if err != nil {
			return err
		}

		err = ctx.binding.SetStreamSources(
			chatID,
			ntgcalls.CaptureStream,
			mediaDescription,
		)
		if err != nil {
			return err
		}

		ctx.p2pConfigsMutex.Lock()
		dhConfig := ntgcalls.DhConfig{
			G:      p2pConfig.DHConfig.G,
			P:      p2pConfig.DHConfig.P,
			Random: p2pConfig.DHConfig.Random,
		}
		gaOrB := p2pConfig.GAorB
		ctx.p2pConfigsMutex.Unlock()

		newGAorB, err := ctx.binding.InitExchange(chatID, dhConfig, gaOrB)
		if err != nil {
			return err
		}

		ctx.p2pConfigsMutex.Lock()
		p2pConfig.GAorB = newGAorB
		ctx.p2pConfigsMutex.Unlock()

		protocolRaw, err := ntgcalls.GetProtocol()
		if err != nil {
			return err
		}
		protocol := &tg.PhoneCallProtocol{
			UdpP2P:          protocolRaw.UDPP2P,
			UdpReflector:    protocolRaw.UDPReflector,
			MinLayer:        protocolRaw.MinLayer,
			MaxLayer:        protocolRaw.MaxLayer,
			LibraryVersions: protocolRaw.Versions,
		}

		userID, err := ctx.app.GetSendableUser(chatID)
		if err != nil {
			return err
		}

		ctx.inputCallsMutex.RLock()
		inputCall := ctx.inputCalls[chatID]
		ctx.inputCallsMutex.RUnlock()

		ctx.p2pConfigsMutex.RLock()
		isOutgoing := p2pConfig.IsOutgoing
		gaOrBHash := p2pConfig.GAorB
		ctx.p2pConfigsMutex.RUnlock()

		if isOutgoing {
			_, err = ctx.app.PhoneRequestCall(
				&tg.PhoneRequestCallParams{
					Protocol: protocol,
					UserID:   userID,
					GAHash:   gaOrBHash,
					RandomID: int32(tg.GenRandInt()),
					Video: mediaDescription.Camera != nil ||
						mediaDescription.Screen != nil,
				},
			)
			if err != nil {
				return err
			}
		} else {
			_, err = ctx.app.PhoneAcceptCall(
				inputCall,
				gaOrBHash,
				protocol,
			)
			if err != nil {
				return err
			}
		}

		select {
		case err = <-p2pConfig.WaitData:
			if err != nil {
				return err
			}
		case <-time.After(10 * time.Second):
			return fmt.Errorf("timed out waiting for an answer")
		}

		ctx.p2pConfigsMutex.RLock()
		gaOrB = p2pConfig.GAorB
		fingerprint := p2pConfig.KeyFingerprint
		ctx.p2pConfigsMutex.RUnlock()

		res, err := ctx.binding.ExchangeKeys(chatID, gaOrB, fingerprint)
		if err != nil {
			return err
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
				return err
			}
			ctx.p2pConfigsMutex.Lock()
			p2pConfig.PhoneCall = confirmRes.PhoneCall.(*tg.PhoneCallObj)
			ctx.p2pConfigsMutex.Unlock()
		}

		ctx.p2pConfigsMutex.RLock()
		phoneCall := p2pConfig.PhoneCall
		ctx.p2pConfigsMutex.RUnlock()

		err = ctx.binding.ConnectP2P(
			chatID,
			parseRTCServers(phoneCall.Connections),
			phoneCall.Protocol.LibraryVersions,
			phoneCall.P2PAllowed,
			"",
		)
		if err != nil {
			return err
		}

	} else {
		// Group call handling
		var err error
		jsonParams, err = ctx.binding.CreateCall(chatID)
		if err != nil {
			ctx.binding.Stop(chatID)
			return err
		}

		err = ctx.binding.SetStreamSources(chatID, ntgcalls.CaptureStream, mediaDescription)
		if err != nil {
			ctx.binding.Stop(chatID)
			return err
		}

		inputGroupCall, err := ctx.GetInputGroupCall(chatID)
		if err != nil {
			ctx.binding.Stop(chatID)
			return err
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
					UserID:     ctx.me.ID,
					AccessHash: ctx.me.AccessHash,
				},
			},
		)
		if err != nil {
			ctx.binding.Stop(chatID)
			return err
		}

		callRes := callResRaw.(*tg.UpdatesObj)
		for _, u := range callRes.Updates {
			switch update := u.(type) {
			case *tg.UpdateGroupCallConnection:
				resultParams = update.Params.Data
			}
		}

		err = ctx.binding.Connect(chatID, resultParams, false)
		if err != nil {
			return err
		}

		connectionMode, err := ctx.binding.GetConnectionMode(chatID)
		if err != nil {
			return err
		}

		if connectionMode == ntgcalls.StreamConnection && len(jsonParams) > 0 {
			ctx.pendingConnectionsMutex.Lock()
			ctx.pendingConnections[chatID] = &PendingConnection{
				MediaDescription: mediaDescription,
				Payload:          jsonParams,
			}
			ctx.pendingConnectionsMutex.Unlock()
		}
	}

	select {
	case err := <-waitChan:
		return err
	case <-time.After(20 * time.Second):
		return fmt.Errorf("connection timeout: no response from ntgcalls")
	}
}
