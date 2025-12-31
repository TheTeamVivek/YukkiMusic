package ubot

import (
	"errors"
	"fmt"
	"slices"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/ntgcalls"
)

type participantUpdate struct {
	participantId      int64
	participant        *tg.GroupCallParticipant
	wasCamera          bool
	wasScreen          bool
	hasCamera          bool
	hasScreen          bool
	cameraEndpoint     string
	screenEndpoint     string
	cameraSourceGroups []*tg.GroupCallParticipantVideoSourceGroup
	screenSourceGroups []*tg.GroupCallParticipantVideoSourceGroup
	oldCameraEndpoint  string
	oldScreenEndpoint  string
}

func (ctx *Context) handleUpdates() {
	ctx.app.AddRawHandler(
		&tg.UpdatePhoneCallSignalingData{},
		func(m tg.Update, c *tg.Client) error {
			signalingData := m.(*tg.UpdatePhoneCallSignalingData)
			userId, err := ctx.convertCallId(signalingData.PhoneCallID)
			if err == nil {
				_ = ctx.binding.SendSignalingData(userId, signalingData.Data)
			}
			return nil
		},
	)

	ctx.app.AddRawHandler(
		&tg.UpdatePhoneCall{},
		func(m tg.Update, _ *tg.Client) error {
			phoneCall := m.(*tg.UpdatePhoneCall).PhoneCall

			var ID int64
			var AccessHash int64
			var userId int64

			switch call := phoneCall.(type) {
			case *tg.PhoneCallAccepted:
				userId = call.ParticipantID
				ID = call.ID
				AccessHash = call.AccessHash
			case *tg.PhoneCallWaiting:
				userId = call.ParticipantID
				ID = call.ID
				AccessHash = call.AccessHash
			case *tg.PhoneCallRequested:
				userId = call.AdminID
				ID = call.ID
				AccessHash = call.AccessHash
			case *tg.PhoneCallObj:
				userId = call.AdminID
			case *tg.PhoneCallDiscarded:
				userId, _ = ctx.convertCallId(call.ID)
			}

			switch phoneCall.(type) {
			case *tg.PhoneCallAccepted, *tg.PhoneCallRequested, *tg.PhoneCallWaiting:
				ctx.inputCallsMutex.Lock()
				ctx.inputCalls[userId] = &tg.InputPhoneCall{
					ID:         ID,
					AccessHash: AccessHash,
				}
				ctx.inputCallsMutex.Unlock()
			}

			ctx.p2pConfigsMutex.RLock()
			p2pConfig := ctx.p2pConfigs[userId]
			ctx.p2pConfigsMutex.RUnlock()

			switch call := phoneCall.(type) {
			case *tg.PhoneCallAccepted:
				if p2pConfig != nil {
					ctx.p2pConfigsMutex.Lock()
					p2pConfig.GAorB = call.GB
					ctx.p2pConfigsMutex.Unlock()
					p2pConfig.WaitData <- nil
				}
			case *tg.PhoneCallObj:
				if p2pConfig != nil {
					ctx.p2pConfigsMutex.Lock()
					p2pConfig.GAorB = call.GAOrB
					p2pConfig.KeyFingerprint = call.KeyFingerprint
					p2pConfig.PhoneCall = call
					ctx.p2pConfigsMutex.Unlock()
					p2pConfig.WaitData <- nil

				}
			case *tg.PhoneCallDiscarded:
				var reasonMessage string

				switch call.Reason.(type) {
				case *tg.PhoneCallDiscardReasonBusy:
					reasonMessage = fmt.Sprintf("the user %d is busy", userId)
				case *tg.PhoneCallDiscardReasonHangup:
					reasonMessage = fmt.Sprintf("call declined by %d", userId)
				}
				if p2pConfig != nil {
					p2pConfig.WaitData <- errors.New(reasonMessage)
				}
				ctx.inputCallsMutex.Lock()
				delete(ctx.inputCalls, userId)
				ctx.inputCallsMutex.Unlock()

				ctx.binding.Stop(userId)

			case *tg.PhoneCallRequested:
				if p2pConfig == nil {
					p2pConfigs, err := ctx.getP2PConfigs(call.GAHash)
					if err != nil {
						return err
					}
					ctx.p2pConfigsMutex.Lock()
					ctx.p2pConfigs[userId] = p2pConfigs
					ctx.p2pConfigsMutex.Unlock()

					ctx.callbacksMutex.RLock()
					callbacks := make([]func(client *Context, chatId int64), len(ctx.incomingCallCallbacks))
					copy(callbacks, ctx.incomingCallCallbacks)
					ctx.callbacksMutex.RUnlock()

					for _, callback := range callbacks {
						go callback(ctx, userId)
					}
				}
			}
			return nil
		},
	)

	ctx.app.AddRawHandler(
		&tg.UpdateGroupCallParticipants{},
		func(m tg.Update, c *tg.Client) error {
			participantsUpdate := m.(*tg.UpdateGroupCallParticipants)
			chatId, err := ctx.convertGroupCallId(
				participantsUpdate.Call.(*tg.InputGroupCallObj).ID,
			)
			if err != nil {
				return nil
			}

			var updates []participantUpdate
			var selfUpdate *tg.GroupCallParticipant

			ctx.participantsMutex.Lock()
			if ctx.callParticipants[chatId] == nil {
				ctx.callParticipants[chatId] = &CallParticipantsCache{
					CallParticipants: make(map[int64]*tg.GroupCallParticipant),
				}
			}

			ctx.callSourcesMutex.Lock()
			if ctx.callSources == nil {
				ctx.callSources = make(map[int64]*CallSources)
			}
			if ctx.callSources[chatId] == nil {
				ctx.callSources[chatId] = &CallSources{
					CameraSources: make(map[int64]string),
					ScreenSources: make(map[int64]string),
				}
			}

			for _, participant := range participantsUpdate.Participants {
				participantId := getParticipantId(participant.Peer)

				if participant.Left {
					delete(
						ctx.callParticipants[chatId].CallParticipants,
						participantId,
					)

					var oldCamera, oldScreen string
					oldCamera = ctx.callSources[chatId].CameraSources[participantId]
					oldScreen = ctx.callSources[chatId].ScreenSources[participantId]
					delete(ctx.callSources[chatId].CameraSources, participantId)
					delete(ctx.callSources[chatId].ScreenSources, participantId)

					if oldCamera != "" || oldScreen != "" {
						updates = append(updates, participantUpdate{
							participantId:     participantId,
							participant:       participant,
							oldCameraEndpoint: oldCamera,
							oldScreenEndpoint: oldScreen,
						})
					}
					continue
				}

				ctx.callParticipants[chatId].CallParticipants[participantId] = participant

				wasCamera := ctx.callSources[chatId].CameraSources[participantId] != ""
				wasScreen := ctx.callSources[chatId].ScreenSources[participantId] != ""
				hasCamera := participant.Video != nil
				hasScreen := participant.Presentation != nil

				update := participantUpdate{
					participantId: participantId,
					participant:   participant,
					wasCamera:     wasCamera,
					wasScreen:     wasScreen,
					hasCamera:     hasCamera,
					hasScreen:     hasScreen,
				}

				if hasCamera && !wasCamera {
					update.cameraEndpoint = participant.Video.Endpoint
					update.cameraSourceGroups = participant.Video.SourceGroups
					ctx.callSources[chatId].CameraSources[participantId] = participant.Video.Endpoint
				} else if !hasCamera && wasCamera {
					update.oldCameraEndpoint = ctx.callSources[chatId].CameraSources[participantId]
					delete(ctx.callSources[chatId].CameraSources, participantId)
				}

				if hasScreen && !wasScreen {
					update.screenEndpoint = participant.Presentation.Endpoint
					update.screenSourceGroups = participant.Presentation.SourceGroups
					ctx.callSources[chatId].ScreenSources[participantId] = participant.Presentation.Endpoint
				} else if !hasScreen && wasScreen {
					update.oldScreenEndpoint = ctx.callSources[chatId].ScreenSources[participantId]
					delete(ctx.callSources[chatId].ScreenSources, participantId)
				}

				if update.cameraEndpoint != "" || update.screenEndpoint != "" ||
					update.oldCameraEndpoint != "" || update.oldScreenEndpoint != "" {
					updates = append(updates, update)
				}

				if participantId == ctx.self.ID {
					selfUpdate = participant
				}
			}

			ctx.callParticipants[chatId].LastMtprotoUpdate = time.Now()
			ctx.callSourcesMutex.Unlock()
			ctx.participantsMutex.Unlock()

			for _, update := range updates {
				if update.cameraEndpoint != "" {
					_, _ = ctx.binding.AddIncomingVideo(
						chatId,
						update.cameraEndpoint,
						parseVideoSources(update.cameraSourceGroups),
					)
				}
				if update.oldCameraEndpoint != "" {
					_ = ctx.binding.RemoveIncomingVideo(
						chatId,
						update.oldCameraEndpoint,
					)
				}

				if update.screenEndpoint != "" {
					_, _ = ctx.binding.AddIncomingVideo(
						chatId,
						update.screenEndpoint,
						parseVideoSources(update.screenSourceGroups),
					)
				}
				if update.oldScreenEndpoint != "" {
					_ = ctx.binding.RemoveIncomingVideo(
						chatId,
						update.oldScreenEndpoint,
					)
				}
			}

			if selfUpdate != nil {
				connectionMode, err := ctx.binding.GetConnectionMode(chatId)
				if err != nil {
					return nil
				}

				if connectionMode == ntgcalls.StreamConnection &&
					selfUpdate.CanSelfUnmute {
					ctx.pendingConnectionsMutex.RLock()
					pending := ctx.pendingConnections[chatId]
					ctx.pendingConnectionsMutex.RUnlock()

					if pending != nil {
						ctx.connectCall(
							chatId,
							pending.MediaDescription,
							pending.Payload,
						)
					}

				} else if !selfUpdate.CanSelfUnmute {
					ctx.mutedByAdminMutex.Lock()
					if !slices.Contains(ctx.mutedByAdmin, chatId) {
						ctx.mutedByAdmin = append(ctx.mutedByAdmin, chatId)
					}
					ctx.mutedByAdminMutex.Unlock()

				} else {
					ctx.mutedByAdminMutex.Lock()
					wasMuted := slices.Contains(ctx.mutedByAdmin, chatId)
					if wasMuted {
						ctx.mutedByAdmin = stdRemove(ctx.mutedByAdmin, chatId)
					}
					ctx.mutedByAdminMutex.Unlock()

					if wasMuted {
						state, err := ctx.binding.GetState(chatId)
						if err != nil {
							return nil
						}
						if err := ctx.setCallStatus(participantsUpdate.Call, state); err != nil {
							return nil
						}
					}
				}
			}

			return nil
		},
	)

	ctx.app.AddRawHandler(
		&tg.UpdateGroupCall{},
		func(m tg.Update, c *tg.Client) error {
			updateGroupCall := m.(*tg.UpdateGroupCall)
			if groupCallRaw := updateGroupCall.Call; groupCallRaw != nil {
				chatID, err := ctx.parseChatId(updateGroupCall.Peer)
				if err != nil {
					return err
				}
				switch groupCall := groupCallRaw.(type) {
				case *tg.GroupCallObj:
					ctx.inputGroupCallsMutex.Lock()
					ctx.inputGroupCalls[chatID] = &tg.InputGroupCallObj{
						ID:         groupCall.ID,
						AccessHash: groupCall.AccessHash,
					}
					ctx.inputGroupCallsMutex.Unlock()
					return nil
				case *tg.GroupCallDiscarded:
					ctx.inputGroupCallsMutex.Lock()
					delete(ctx.inputGroupCalls, chatID)
					ctx.inputGroupCallsMutex.Unlock()
					ctx.binding.Stop(chatID)
					return nil
				}
			}
			return nil
		},
	)

	ctx.binding.OnRequestBroadcastTimestamp(func(chatId int64) {
		ctx.inputGroupCallsMutex.RLock()
		inputGroupCall := ctx.inputGroupCalls[chatId]
		ctx.inputGroupCallsMutex.RUnlock()

		if inputGroupCall != nil {
			channels, err := ctx.app.PhoneGetGroupCallStreamChannels(
				inputGroupCall,
			)
			if err == nil {
				_ = ctx.binding.SendBroadcastTimestamp(
					chatId,
					channels.Channels[0].LastTimestampMs,
				)
			}
		}
	})

	ctx.binding.OnRequestBroadcastPart(
		func(chatId int64, segmentPartRequest ntgcalls.SegmentPartRequest) {
			ctx.inputGroupCallsMutex.RLock()
			inputGroupCall := ctx.inputGroupCalls[chatId]
			ctx.inputGroupCallsMutex.RUnlock()

			if inputGroupCall != nil {
				file, err := ctx.app.UploadGetFile(
					&tg.UploadGetFileParams{
						Location: &tg.InputGroupCallStream{
							Call:         inputGroupCall,
							TimeMs:       segmentPartRequest.Timestamp,
							Scale:        0,
							VideoChannel: segmentPartRequest.ChannelID,
							VideoQuality: max(
								int32(segmentPartRequest.Quality),
								0,
							),
						},
						Offset: 0,
						Limit:  segmentPartRequest.Limit,
					},
				)

				status := ntgcalls.SegmentStatusNotReady
				var data []byte
				data = nil

				if err != nil {
					secondsWait := tg.GetFloodWait(err)
					if secondsWait == 0 {
						status = ntgcalls.SegmentStatusResyncNeeded
					}
				} else {
					data = file.(*tg.UploadFileObj).Bytes
					status = ntgcalls.SegmentStatusSuccess
				}

				_ = ctx.binding.SendBroadcastPart(
					chatId,
					segmentPartRequest.SegmentID,
					segmentPartRequest.PartID,
					status,
					segmentPartRequest.QualityUpdate,
					data,
				)
			}
		},
	)

	ctx.binding.OnSignal(func(chatId int64, signal []byte) {
		ctx.inputCallsMutex.RLock()
		inputCall := ctx.inputCalls[chatId]
		ctx.inputCallsMutex.RUnlock()

		_, _ = ctx.app.PhoneSendSignalingData(inputCall, signal)
	})

	ctx.binding.OnConnectionChange(
		func(chatId int64, state ntgcalls.NetworkInfo) {
			ctx.waitConnectMutex.RLock()
			waitChan := ctx.waitConnect[chatId]
			ctx.waitConnectMutex.RUnlock()
			if waitChan != nil {
				switch state.State {
				case ntgcalls.Connected:
					waitChan <- nil
				case ntgcalls.Closed, ntgcalls.Failed:
					waitChan <- fmt.Errorf("connection failed")
				case ntgcalls.Timeout:
					waitChan <- fmt.Errorf("connection timeout")
				default:
				}
			}
		},
	)

	ctx.binding.OnUpgrade(func(chatId int64, state ntgcalls.MediaState) {
		ctx.inputGroupCallsMutex.RLock()
		inputGroupCall := ctx.inputGroupCalls[chatId]
		ctx.inputGroupCallsMutex.RUnlock()

		if err := ctx.setCallStatus(inputGroupCall, state); err != nil {
			fmt.Println(err)
		}
	})

	ctx.binding.OnStreamEnd(
		func(chatId int64, streamType ntgcalls.StreamType, streamDevice ntgcalls.StreamDevice) {
			ctx.callbacksMutex.RLock()
			callbacks := make(
				[]ntgcalls.StreamEndCallback,
				len(ctx.streamEndCallbacks),
			)
			copy(callbacks, ctx.streamEndCallbacks)
			ctx.callbacksMutex.RUnlock()

			for _, callback := range callbacks {
				go callback(chatId, streamType, streamDevice)
			}
		},
	)

	ctx.binding.OnFrame(
		func(chatId int64, mode ntgcalls.StreamMode, device ntgcalls.StreamDevice, frames []ntgcalls.Frame) {
			ctx.callbacksMutex.RLock()
			callbacks := make([]ntgcalls.FrameCallback, len(ctx.frameCallbacks))
			copy(callbacks, ctx.frameCallbacks)
			ctx.callbacksMutex.RUnlock()

			for _, callback := range callbacks {
				go callback(chatId, mode, device, frames)
			}
		},
	)
}
