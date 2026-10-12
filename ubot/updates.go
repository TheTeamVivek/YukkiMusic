package ubot

import (
	"errors"
	"fmt"
	"slices"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"

	"yukkimusic/ntgcalls"
)

type participantUpdate struct {
	participantID      int64
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

var (
	ErrConnectionTimeout = errors.New("connection timeout")
	ErrConnectionFailed  = errors.New("connection failed")
)

func (ctx *Context) registerHandlers() {
	ctx.app.AddRawHandler(
		&tg.UpdatePhoneCallSignalingData{},
		func(m tg.Update, c *tg.Client) error {
			signalingData := m.(*tg.UpdatePhoneCallSignalingData)
			userID, err := ctx.convertCallID(signalingData.PhoneCallID)
			if err == nil {
				_ = ctx.binding.SendSignalingData(userID, signalingData.Data)
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
			var userID int64

			switch call := phoneCall.(type) {
			case *tg.PhoneCallAccepted:
				userID = call.ParticipantID
				ID = call.ID
				AccessHash = call.AccessHash
			case *tg.PhoneCallWaiting:
				userID = call.ParticipantID
				ID = call.ID
				AccessHash = call.AccessHash
			case *tg.PhoneCallRequested:
				userID = call.AdminID
				ID = call.ID
				AccessHash = call.AccessHash
			case *tg.PhoneCallObj:
				userID = call.AdminID
			case *tg.PhoneCallDiscarded:
				userID, _ = ctx.convertCallID(call.ID)
			}

			switch phoneCall.(type) {
			case *tg.PhoneCallAccepted, *tg.PhoneCallRequested, *tg.PhoneCallWaiting:
				ctx.inputCallsMutex.Lock()
				ctx.inputCalls[userID] = &tg.InputPhoneCall{
					ID:         ID,
					AccessHash: AccessHash,
				}
				ctx.inputCallsMutex.Unlock()
			}

			ctx.p2pConfigsMutex.RLock()
			p2pConfig := ctx.p2pConfigs[userID]
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
					reasonMessage = fmt.Sprintf("the user %d is busy", userID)
				case *tg.PhoneCallDiscardReasonHangup:
					reasonMessage = fmt.Sprintf("call declined by %d", userID)
				}
				if p2pConfig != nil {
					p2pConfig.WaitData <- errors.New(reasonMessage)
				}
				ctx.inputCallsMutex.Lock()
				delete(ctx.inputCalls, userID)
				ctx.inputCallsMutex.Unlock()

				ctx.binding.Stop(userID)

			case *tg.PhoneCallRequested:
				if p2pConfig == nil {
					p2pConfigs, err := ctx.newP2PConfig(call.GAHash)
					if err != nil {
						return err
					}
					ctx.p2pConfigsMutex.Lock()
					ctx.p2pConfigs[userID] = p2pConfigs
					ctx.p2pConfigsMutex.Unlock()

					ctx.callbacksMutex.RLock()
					callbacks := make([]func(client *Context, chatID int64), len(ctx.incomingCallCallbacks))
					copy(callbacks, ctx.incomingCallCallbacks)
					ctx.callbacksMutex.RUnlock()

					for _, callback := range callbacks {
						go callback(ctx, userID)
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
			chatID, err := ctx.convertGroupCallID(
				participantsUpdate.Call.(*tg.InputGroupCallObj).ID,
			)
			if err != nil {
				return nil
			}

			var updates []participantUpdate
			var selfUpdate *tg.GroupCallParticipant

			ctx.participantsMutex.Lock()
			if ctx.callParticipants[chatID] == nil {
				ctx.callParticipants[chatID] = &CallParticipantsCache{
					Participants: make(map[int64]*tg.GroupCallParticipant),
				}
			}

			ctx.callSourcesMutex.Lock()
			if ctx.callSources == nil {
				ctx.callSources = make(map[int64]*CallSources)
			}
			if ctx.callSources[chatID] == nil {
				ctx.callSources[chatID] = &CallSources{
					Camera: make(map[int64]string),
					Screen: make(map[int64]string),
				}
			}

			for _, participant := range participantsUpdate.Participants {
				participantID := getParticipantID(participant.Peer)

				if participant.Left {
					delete(
						ctx.callParticipants[chatID].Participants,
						participantID,
					)

					var oldCamera, oldScreen string
					oldCamera = ctx.callSources[chatID].Camera[participantID]
					oldScreen = ctx.callSources[chatID].Screen[participantID]
					delete(ctx.callSources[chatID].Camera, participantID)
					delete(ctx.callSources[chatID].Screen, participantID)

					if oldCamera != "" || oldScreen != "" {
						updates = append(updates, participantUpdate{
							participantID:     participantID,
							participant:       participant,
							oldCameraEndpoint: oldCamera,
							oldScreenEndpoint: oldScreen,
						})
					}
					continue
				}

				ctx.callParticipants[chatID].Participants[participantID] = participant

				wasCamera := ctx.callSources[chatID].Camera[participantID] != ""
				wasScreen := ctx.callSources[chatID].Screen[participantID] != ""
				hasCamera := participant.Video != nil
				hasScreen := participant.Presentation != nil

				update := participantUpdate{
					participantID: participantID,
					participant:   participant,
					wasCamera:     wasCamera,
					wasScreen:     wasScreen,
					hasCamera:     hasCamera,
					hasScreen:     hasScreen,
				}

				if hasCamera && !wasCamera {
					update.cameraEndpoint = participant.Video.Endpoint
					update.cameraSourceGroups = participant.Video.SourceGroups
					ctx.callSources[chatID].Camera[participantID] = participant.Video.Endpoint
				} else if !hasCamera && wasCamera {
					update.oldCameraEndpoint = ctx.callSources[chatID].Camera[participantID]
					delete(ctx.callSources[chatID].Camera, participantID)
				}

				if hasScreen && !wasScreen {
					update.screenEndpoint = participant.Presentation.Endpoint
					update.screenSourceGroups = participant.Presentation.SourceGroups
					ctx.callSources[chatID].Screen[participantID] = participant.Presentation.Endpoint
				} else if !hasScreen && wasScreen {
					update.oldScreenEndpoint = ctx.callSources[chatID].Screen[participantID]
					delete(ctx.callSources[chatID].Screen, participantID)
				}

				if update.cameraEndpoint != "" || update.screenEndpoint != "" ||
					update.oldCameraEndpoint != "" || update.oldScreenEndpoint != "" {
					updates = append(updates, update)
				}

				if participantID == ctx.me.ID {
					selfUpdate = participant
				}
			}

			ctx.callParticipants[chatID].lastUpdate = time.Now()
			ctx.callSourcesMutex.Unlock()
			ctx.participantsMutex.Unlock()

			for _, update := range updates {
				if update.cameraEndpoint != "" {
					_, _ = ctx.binding.AddIncomingVideo(
						chatID,
						update.participantID,
						update.cameraEndpoint,
						parseVideoSources(update.cameraSourceGroups),
					)
				}
				if update.oldCameraEndpoint != "" {
					_, _ = ctx.binding.RemoveIncomingVideo(
						chatID,
						update.oldCameraEndpoint,
					)
				}

				if update.screenEndpoint != "" {
					_, _ = ctx.binding.AddIncomingVideo(
						chatID,
						update.participantID,
						update.screenEndpoint,
						parseVideoSources(update.screenSourceGroups),
					)
				}
				if update.oldScreenEndpoint != "" {
					_, _ = ctx.binding.RemoveIncomingVideo(
						chatID,
						update.oldScreenEndpoint,
					)
				}
			}

			if selfUpdate != nil {
				connectionMode, err := ctx.binding.GetConnectionMode(chatID)
				if err != nil {
					return nil
				}

				if connectionMode == ntgcalls.StreamConnection &&
					selfUpdate.CanSelfUnmute {
					ctx.pendingConnectionsMutex.RLock()
					pending := ctx.pendingConnections[chatID]
					ctx.pendingConnectionsMutex.RUnlock()

					if pending != nil {
						ctx.connectCall(
							chatID,
							pending.MediaDescription,
							pending.Payload,
						)
					}

				} else if !selfUpdate.CanSelfUnmute {
					ctx.mutedByAdminMutex.Lock()
					if !slices.Contains(ctx.mutedByAdmin, chatID) {
						ctx.mutedByAdmin = append(ctx.mutedByAdmin, chatID)
					}
					ctx.mutedByAdminMutex.Unlock()

				} else {
					ctx.mutedByAdminMutex.Lock()
					wasMuted := slices.Contains(ctx.mutedByAdmin, chatID)
					if wasMuted {
						ctx.mutedByAdmin = stdRemove(ctx.mutedByAdmin, chatID)
					}
					ctx.mutedByAdminMutex.Unlock()

					if wasMuted {
						state, err := ctx.binding.GetState(chatID)
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
				chatID, err := ctx.parseChatID(updateGroupCall.Peer)
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

	ctx.binding.OnRequestBroadcastTimestamp(func(chatID int64) {
		ctx.inputGroupCallsMutex.RLock()
		inputGroupCall := ctx.inputGroupCalls[chatID]
		ctx.inputGroupCallsMutex.RUnlock()

		if inputGroupCall != nil {
			channels, err := ctx.app.PhoneGetGroupCallStreamChannels(
				inputGroupCall,
			)
			if err == nil {
				_ = ctx.binding.SendBroadcastTimestamp(
					chatID,
					channels.Channels[0].LastTimestampMs,
				)
			}
		}
	})

	ctx.binding.OnRequestBroadcastPart(
		func(chatID int64, segmentPartRequest ntgcalls.SegmentPartRequest) {
			ctx.inputGroupCallsMutex.RLock()
			inputGroupCall := ctx.inputGroupCalls[chatID]
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
					chatID,
					segmentPartRequest.SegmentID,
					segmentPartRequest.PartID,
					status,
					segmentPartRequest.QualityUpdate,
					data,
				)
			}
		},
	)

	ctx.binding.OnSignal(func(chatID int64, signal []byte) {
		ctx.inputCallsMutex.RLock()
		inputCall := ctx.inputCalls[chatID]
		ctx.inputCallsMutex.RUnlock()

		_, _ = ctx.app.PhoneSendSignalingData(inputCall, signal)
	})

	ctx.binding.OnConnectionChange(
		func(chatID int64, state ntgcalls.NetworkInfo) {
			ctx.waitConnectMutex.RLock()
			waitChan := ctx.waitConnect[chatID]
			ctx.waitConnectMutex.RUnlock()

			if waitChan != nil {
				switch state.State {
				case ntgcalls.Connected:
					waitChan <- nil
				case ntgcalls.Closed, ntgcalls.Failed:
					waitChan <- ErrConnectionFailed
				case ntgcalls.Timeout:
					waitChan <- ErrConnectionTimeout
				default:
				}
			}
		},
	)

	ctx.binding.OnUpgrade(func(chatID int64, state ntgcalls.MediaState) {
		ctx.inputGroupCallsMutex.RLock()
		inputGroupCall := ctx.inputGroupCalls[chatID]
		ctx.inputGroupCallsMutex.RUnlock()

		if err := ctx.setCallStatus(inputGroupCall, state); err != nil {
			fmt.Println(err)
		}
	})

	ctx.binding.OnStreamEnd(
		func(chatID int64, streamType ntgcalls.StreamType, streamDevice ntgcalls.StreamDevice) {
			ctx.callbacksMutex.RLock()
			callbacks := make(
				[]ntgcalls.StreamEndCallback,
				len(ctx.streamEndCallbacks),
			)
			copy(callbacks, ctx.streamEndCallbacks)
			ctx.callbacksMutex.RUnlock()

			for _, callback := range callbacks {
				go callback(chatID, streamType, streamDevice)
			}
		},
	)

	ctx.binding.OnFrame(
		func(chatID int64, mode ntgcalls.StreamMode, device ntgcalls.StreamDevice, frames []ntgcalls.Frame) {
			ctx.callbacksMutex.RLock()
			callbacks := make([]ntgcalls.FrameCallback, len(ctx.frameCallbacks))
			copy(callbacks, ctx.frameCallbacks)
			ctx.callbacksMutex.RUnlock()

			for _, callback := range callbacks {
				go callback(chatID, mode, device, frames)
			}
		},
	)
}
