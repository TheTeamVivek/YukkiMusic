package ubot

import (
	"fmt"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/ntgcalls"
	"main/ubot/types"
)

func (ctx *Context) connectCall(chatId int64, mediaDescription ntgcalls.MediaDescription, jsonParams string) error {
	defer func() {
		ctx.waitConnectMutex.Lock()
		if ctx.waitConnect[chatId] != nil {
			delete(ctx.waitConnect, chatId)
		}
		ctx.waitConnectMutex.Unlock()
	}()

	ctx.waitConnectMutex.Lock()
	ctx.waitConnect[chatId] = make(chan error)
	waitChan := ctx.waitConnect[chatId]
	ctx.waitConnectMutex.Unlock()

	if chatId >= 0 {
		defer func() {
			ctx.p2pConfigsMutex.Lock()
			if ctx.p2pConfigs[chatId] != nil {
				delete(ctx.p2pConfigs, chatId)
			}
			ctx.p2pConfigsMutex.Unlock()
		}()

		ctx.p2pConfigsMutex.Lock()
		if ctx.p2pConfigs[chatId] == nil {
			ctx.p2pConfigsMutex.Unlock()
			p2pConfigs, err := ctx.getP2PConfigs(nil)
			if err != nil {
				return err
			}
			ctx.p2pConfigsMutex.Lock()
			ctx.p2pConfigs[chatId] = p2pConfigs
		}
		p2pConfig := ctx.p2pConfigs[chatId]
		ctx.p2pConfigsMutex.Unlock()

		err := ctx.binding.CreateP2PCall(chatId)
		if err != nil {
			return err
		}

		err = ctx.binding.SetStreamSources(chatId, ntgcalls.CaptureStream, mediaDescription)
		if err != nil {
			return err
		}

		ctx.p2pConfigsMutex.Lock()

		p2pConfig.GAorB, err = ctx.binding.InitExchange(chatId, ntgcalls.DhConfig{
			G:      p2pConfig.DhConfig.G,
			P:      p2pConfig.DhConfig.P,
			Random: p2pConfig.DhConfig.Random,
		}, p2pConfig.GAorB)
		ctx.p2pConfigsMutex.Unlock()

		if err != nil {
			return err
		}

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
			return err
		}
		ctx.inputCallsMutex.RLock()
		inputCall := ctx.inputCalls[chatId]
		ctx.inputCallsMutex.RUnlock()

		if p2pConfig.IsOutgoing {
			_, err = ctx.app.PhoneRequestCall(
				&tg.PhoneRequestCallParams{
					Protocol: protocol,
					UserID:   userId,
					GAHash:   p2pConfig.GAorB,
					RandomID: int32(tg.GenRandInt()),
					Video:    mediaDescription.Camera != nil || mediaDescription.Screen != nil,
				},
			)
			if err != nil {
				return err
			}
		} else {
			_, err = ctx.app.PhoneAcceptCall(
				inputCall,
				p2pConfig.GAorB,
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
		gaOrB := p2pConfig.GAorB
		fingerprint := p2pConfig.KeyFingerprint
		ctx.p2pConfigsMutex.RUnlock()

		res, err := ctx.binding.ExchangeKeys(
			chatId,
			gaOrB,
			fingerprint,
		)
		if err != nil {
			return err
		}

		if p2pConfig.IsOutgoing {
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
			chatId,
			parseRTCServers(phoneCall.Connections),
			phoneCall.Protocol.LibraryVersions,
			phoneCall.P2PAllowed,
		)
		if err != nil {
			return err
		}

	} else {
		var err error
		jsonParams, err = ctx.binding.CreateCall(chatId)
		if err != nil {
			ctx.binding.Stop(chatId)
			return err
		}

		err = ctx.binding.SetStreamSources(chatId, ntgcalls.CaptureStream, mediaDescription)
		if err != nil {
			ctx.binding.Stop(chatId)
			return err
		}

		inputGroupCall, err := ctx.GetInputGroupCall(chatId)
		if err != nil {
			ctx.binding.Stop(chatId)
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
					UserID:     ctx.self.ID,
					AccessHash: ctx.self.AccessHash,
				},
			},
		)
		if err != nil {
			return err
		}
		callRes := callResRaw.(*tg.UpdatesObj)
		for _, u := range callRes.Updates {
			switch update := u.(type) {
			case *tg.UpdateGroupCallConnection:
				resultParams = update.Params.Data
			}
		}

		err = ctx.binding.Connect(
			chatId,
			resultParams,
			false,
		)
		if err != nil {
			return err
		}

		connectionMode, err := ctx.binding.GetConnectionMode(chatId)
		if err != nil {
			return err
		}

		if connectionMode == ntgcalls.StreamConnection && len(jsonParams) > 0 {
			ctx.pendingConnectionsMutex.Lock()
			ctx.pendingConnections[chatId] = &types.PendingConnection{
				MediaDescription: mediaDescription,
				Payload:          jsonParams,
			}
			ctx.pendingConnectionsMutex.Unlock()
		}
	}
	return <-waitChan
}
