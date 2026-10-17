package ntgcalls

/*
#include "ntgcalls.h"
#include <stdlib.h>
extern void handleStreamEnd(ntg_instance* handle, int64_t chatID, ntg_stream_type streamType, ntg_stream_device streamDevice, void* userData);
extern void handleUpgrade(ntg_instance* handle, int64_t chatID, ntg_media_state state, void* userData);
extern void handleConnectionChange(ntg_instance* handle, int64_t chatID, ntg_connection_info info, void* userData);
extern void handleSignal(ntg_instance* handle, int64_t chatID, uint8_t* data, size_t size, void* userData);
extern void handleFrames(ntg_instance* handle, int64_t chatID, ntg_stream_mode streamMode, ntg_stream_device streamDevice, ntg_frame* frames, size_t size, void* userData);
extern void handleRemoteSourceChange(ntg_instance* handle, int64_t chatID, ntg_remote_source remoteSource, void* userData);
extern void handleRequestBroadcastTimestamp(ntg_instance* handle, int64_t chatID, void* userData);
extern void handleRequestBroadcastPart(ntg_instance* handle, int64_t chatID, ntg_segment_part_request request, void* userData);
extern void handleUpdateEmojis(ntg_instance* handle, int64_t chatID, char* emojis, void* userData);
extern void handleRequestParticipants(ntg_instance* handle, int64_t chatID, ntg_participants_request request, void* userData);
extern void handleOutboundBlock(ntg_instance* handle, int64_t chatID, uint8_t* block, size_t size, void* userData);
extern void handleSubchainRequest(ntg_instance* handle, int64_t chatID, ntg_subchain_request request, void* userData);
extern void handleLogs(ntg_log_message message, void* userData);
*/
import "C"
import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"yukkimusic/internal/utils"
)

var clientRegistry = struct {
	sync.RWMutex
	clients map[*C.ntg_instance]*Client
}{clients: make(map[*C.ntg_instance]*Client)}

func lookupClient(handle *C.ntg_instance) *Client {
	clientRegistry.RLock()
	defer clientRegistry.RUnlock()
	return clientRegistry.clients[handle]
}

func init() {
	C.ntg_set_log_callback((C.ntg_log_cb)(unsafe.Pointer(C.handleLogs)), nil)
}

// Package-level utility functions
func Version() string {
	return C.GoString(C.ntg_get_version())
}

func GetProtocol() (Protocol, error) {
	var buffer C.ntg_protocol
	if err := parseResult(C.ntg_get_protocol(&buffer)); err != nil {
		return Protocol{}, err
	}
	defer C.ntg_protocol_free(&buffer)
	return Protocol{
		MinLayer:     int32(buffer.min_layer),
		MaxLayer:     int32(buffer.max_layer),
		UDPP2P:       bool(buffer.udp_p2p),
		UDPReflector: bool(buffer.udp_reflector),
		Versions:     parseStringVector(buffer.library_versions, buffer.library_versions_len),
	}, nil
}

func GetMediaDevices() (MediaDevices, error) {
	var buffer C.ntg_media_devices
	if err := parseResult(C.ntg_get_media_devices(&buffer)); err != nil {
		return MediaDevices{}, err
	}
	defer C.ntg_media_devices_free(&buffer)
	return MediaDevices{
		Microphone: parseDeviceInfoVector(buffer.microphone, buffer.microphone_len),
		Speaker:    parseDeviceInfoVector(buffer.speaker, buffer.speaker_len),
		Camera:     parseDeviceInfoVector(buffer.camera, buffer.camera_len),
		Screen:     parseDeviceInfoVector(buffer.screen, buffer.screen_len),
	}, nil
}

func EnableGLibLoop(enable bool) error {
	return parseResult(C.ntg_enable_glib_loop(C.bool(enable)))
}

func Ping() (string, error) {
	var buffer *C.char
	if err := parseResult(C.ntg_ping(&buffer)); err != nil {
		return "", err
	}
	defer C.ntg_string_free(buffer)
	return C.GoString(buffer), nil
}

func NewClient() *Client {
	client := &Client{
		handle: C.ntg_instance_create(),
	}

	clientRegistry.Lock()
	clientRegistry.clients[client.handle] = client
	clientRegistry.Unlock()
	C.ntg_on_stream_end_callback(client.handle, (C.ntg_stream_end_callback_cb)(unsafe.Pointer(C.handleStreamEnd)), nil)
	C.ntg_on_upgrade_callback(client.handle, (C.ntg_upgrade_callback_cb)(unsafe.Pointer(C.handleUpgrade)), nil)
	C.ntg_on_signaling_data_callback(client.handle, (C.ntg_signaling_data_callback_cb)(unsafe.Pointer(C.handleSignal)), nil)
	C.ntg_on_connection_change_callback(client.handle, (C.ntg_connection_change_callback_cb)(unsafe.Pointer(C.handleConnectionChange)), nil)
	C.ntg_on_frames_callback(client.handle, (C.ntg_frames_callback_cb)(unsafe.Pointer(C.handleFrames)), nil)
	C.ntg_on_remote_source_change_callback(client.handle, (C.ntg_remote_source_change_callback_cb)(unsafe.Pointer(C.handleRemoteSourceChange)), nil)
	C.ntg_on_request_broadcast_timestamp_callback(client.handle, (C.ntg_request_broadcast_timestamp_callback_cb)(unsafe.Pointer(C.handleRequestBroadcastTimestamp)), nil)
	C.ntg_on_request_broadcast_part_callback(client.handle, (C.ntg_request_broadcast_part_callback_cb)(unsafe.Pointer(C.handleRequestBroadcastPart)), nil)
	C.ntg_on_update_emojis_callback(client.handle, (C.ntg_update_emojis_callback_cb)(unsafe.Pointer(C.handleUpdateEmojis)), nil)
	C.ntg_on_request_participants_callback(client.handle, (C.ntg_request_participants_callback_cb)(unsafe.Pointer(C.handleRequestParticipants)), nil)
	C.ntg_on_outbound_block_callback(client.handle, (C.ntg_outbound_block_callback_cb)(unsafe.Pointer(C.handleOutboundBlock)), nil)
	C.ntg_on_subchain_request_callback(client.handle, (C.ntg_subchain_request_callback_cb)(unsafe.Pointer(C.handleSubchainRequest)), nil)

	runtime.SetFinalizer(client, func(c *Client) {
		if c.handle != nil {
			c.Free()
		}
	})
	return client
}

//export handleLogs
func handleLogs(message C.ntg_log_message, _ unsafe.Pointer) {
	text := fmt.Sprintf(
		"(%s:%d) %s",
		C.GoString(message.file),
		uint32(message.line),
		C.GoString(message.message),
	)

	switch message.level {
	case C.NTG_LOG_DEBUG:
		utils.Log.Debug(text)
	case C.NTG_LOG_INFO:
		utils.Log.Info(text)
	case C.NTG_LOG_WARNING:
		utils.Log.Warn(text)
	case C.NTG_LOG_ERROR:
		utils.Log.Error(text)
	}
}

//export handleStreamEnd
func handleStreamEnd(handle *C.ntg_instance, chatID C.int64_t, streamType C.ntg_stream_type, streamDevice C.ntg_stream_device, _ unsafe.Pointer) {
	client := lookupClient(handle)
	if client == nil {
		return
	}
	goType := VideoStream
	if streamType == C.NTG_STREAM_TYPE_AUDIO {
		goType = AudioStream
	}
	goDevice := parseStreamDevice(streamDevice)
	for _, callback := range client.getStreamEndCallbacks() {
		go callback(int64(chatID), goType, goDevice)
	}
}

//export handleUpgrade
func handleUpgrade(handle *C.ntg_instance, chatID C.int64_t, state C.ntg_media_state, _ unsafe.Pointer) {
	client := lookupClient(handle)
	if client == nil {
		return
	}
	goState := parseMediaState(state)
	for _, callback := range client.getUpgradeCallbacks() {
		go callback(int64(chatID), goState)
	}
}

//export handleSignal
func handleSignal(handle *C.ntg_instance, chatID C.int64_t, data *C.uint8_t, size C.size_t, _ unsafe.Pointer) {
	client := lookupClient(handle)
	if client == nil {
		return
	}
	goData := C.GoBytes(unsafe.Pointer(data), C.int(size))
	for _, callback := range client.getSignalCallbacks() {
		go callback(int64(chatID), goData)
	}
}

//export handleConnectionChange
func handleConnectionChange(handle *C.ntg_instance, chatID C.int64_t, info C.ntg_connection_info, _ unsafe.Pointer) {
	client := lookupClient(handle)
	if client == nil {
		return
	}
	goInfo := NetworkInfo{
		State: parseConnectionState(info.state),
	}
	switch info.kind {
	case C.NTG_CONNECTION_KIND_NORMAL:
		goInfo.Kind = NormalConnection
	case C.NTG_CONNECTION_KIND_PRESENTATION:
		goInfo.Kind = PresentationConnection
	}
	for _, callback := range client.getConnectionChangeCallbacks() {
		go callback(int64(chatID), goInfo)
	}
}

//export handleFrames
func handleFrames(handle *C.ntg_instance, chatID C.int64_t, streamMode C.ntg_stream_mode, streamDevice C.ntg_stream_device, frames *C.ntg_frame, size C.size_t, _ unsafe.Pointer) {
	client := lookupClient(handle)
	if client == nil {
		return
	}
	goMode := PlaybackStream
	if streamMode == C.NTG_STREAM_MODE_CAPTURE {
		goMode = CaptureStream
	}
	raw := unsafe.Slice(frames, int(size))
	goFrames := make([]Frame, len(raw))
	for i, rawFrame := range raw {
		goFrames[i] = Frame{
			SSRC: uint32(rawFrame.ssrc),
			Data: C.GoBytes(unsafe.Pointer(rawFrame.data), C.int(rawFrame.data_len)),
			FrameData: FrameData{
				AbsoluteCaptureTimestampMs: int64(rawFrame.frame_data.absolute_capture_timestamp_ms),
				Width:                      uint16(rawFrame.frame_data.width),
				Height:                     uint16(rawFrame.frame_data.height),
				Rotation:                   uint16(rawFrame.frame_data.rotation),
			},
		}
	}
	goDevice := parseStreamDevice(streamDevice)
	for _, callback := range client.getFrameCallbacks() {
		go callback(int64(chatID), goMode, goDevice, goFrames)
	}
}

//export handleRemoteSourceChange
func handleRemoteSourceChange(handle *C.ntg_instance, chatID C.int64_t, remoteSource C.ntg_remote_source, _ unsafe.Pointer) {
	client := lookupClient(handle)
	if client == nil {
		return
	}
	goSource := RemoteSource{
		SSRC:   uint32(remoteSource.ssrc),
		State:  parseStreamStatus(remoteSource.state),
		Device: parseStreamDevice(remoteSource.device),
	}
	for _, callback := range client.getRemoteSourceCallbacks() {
		go callback(int64(chatID), goSource)
	}
}

//export handleRequestBroadcastTimestamp
func handleRequestBroadcastTimestamp(handle *C.ntg_instance, chatID C.int64_t, _ unsafe.Pointer) {
	client := lookupClient(handle)
	if client == nil {
		return
	}
	for _, callback := range client.getBroadcastTimestampCallbacks() {
		go callback(int64(chatID))
	}
}

//export handleRequestBroadcastPart
func handleRequestBroadcastPart(handle *C.ntg_instance, chatID C.int64_t, request C.ntg_segment_part_request, _ unsafe.Pointer) {
	client := lookupClient(handle)
	if client == nil {
		return
	}
	var goQuality MediaSegmentQuality
	switch request.quality {
	case C.NTG_MEDIA_SEGMENT_QUALITY_NONE:
		goQuality = SegmentQualityNone
	case C.NTG_MEDIA_SEGMENT_QUALITY_THUMBNAIL:
		goQuality = SegmentQualityThumbnail
	case C.NTG_MEDIA_SEGMENT_QUALITY_MEDIUM:
		goQuality = SegmentQualityMedium
	case C.NTG_MEDIA_SEGMENT_QUALITY_FULL:
		goQuality = SegmentQualityFull
	}
	goRequest := SegmentPartRequest{
		SegmentID:     int64(request.segment_id),
		PartID:        int32(request.part_id),
		Limit:         int32(request.limit),
		Timestamp:     int64(request.timestamp),
		QualityUpdate: bool(request.quality_update),
		ChannelID:     int32(request.channel_id),
		Quality:       goQuality,
	}
	for _, callback := range client.getBroadcastPartCallbacks() {
		go callback(int64(chatID), goRequest)
	}
}

//export handleUpdateEmojis
func handleUpdateEmojis(handle *C.ntg_instance, chatID C.int64_t, emojis *C.char, _ unsafe.Pointer) {
	client := lookupClient(handle)
	if client == nil {
		return
	}
	goEmojis := C.GoString(emojis)
	for _, callback := range client.getEmojisCallbacks() {
		go callback(int64(chatID), goEmojis)
	}
}

//export handleRequestParticipants
func handleRequestParticipants(handle *C.ntg_instance, chatID C.int64_t, request C.ntg_participants_request, _ unsafe.Pointer) {
	client := lookupClient(handle)
	if client == nil {
		return
	}
	raw := unsafe.Slice(request.ssrcs, int(request.ssrcs_len))
	goRequest := ParticipantsRequest{
		SSRCs: make([]uint32, len(raw)),
	}
	for i, ssrc := range raw {
		goRequest.SSRCs[i] = uint32(ssrc)
	}
	for _, callback := range client.getRequestParticipantsCallbacks() {
		go callback(int64(chatID), goRequest)
	}
}

//export handleOutboundBlock
func handleOutboundBlock(handle *C.ntg_instance, chatID C.int64_t, block *C.uint8_t, size C.size_t, _ unsafe.Pointer) {
	client := lookupClient(handle)
	if client == nil {
		return
	}
	goBlock := C.GoBytes(unsafe.Pointer(block), C.int(size))
	for _, callback := range client.getOutboundBlockCallbacks() {
		go callback(int64(chatID), goBlock)
	}
}

//export handleSubchainRequest
func handleSubchainRequest(handle *C.ntg_instance, chatID C.int64_t, request C.ntg_subchain_request, _ unsafe.Pointer) {
	client := lookupClient(handle)
	if client == nil {
		return
	}
	goRequest := SubchainRequest{
		Subchain: int32(request.subchain),
		Height:   int32(request.height),
		Limit:    int32(request.limit),
	}
	for _, callback := range client.getSubchainRequestCallbacks() {
		go callback(int64(chatID), goRequest)
	}
}

func (c *Client) GetState(chatID int64) (MediaState, error) {
	var buffer C.ntg_media_state
	if err := parseResult(C.ntg_get_state(c.handle, C.int64_t(chatID), &buffer)); err != nil {
		return MediaState{}, err
	}
	defer C.ntg_media_state_free(&buffer)
	return parseMediaState(buffer), nil
}

func (c *Client) GetConnectionMode(chatID int64) (ConnectionMode, error) {
	var buffer C.ntg_connection_mode
	if err := parseResult(C.ntg_get_connection_mode(c.handle, C.int64_t(chatID), &buffer)); err != nil {
		return 0, err
	}
	switch buffer {
	case C.NTG_CONNECTION_MODE_RTC:
		return RTCConnection, nil
	case C.NTG_CONNECTION_MODE_STREAM:
		return StreamConnection, nil
	case C.NTG_CONNECTION_MODE_RTMP:
		return RTMPConnection, nil
	default:
		return ConnectionMode(0), fmt.Errorf("unknown connection mode")
	}
}

func (c *Client) GetEmojisFingerprint(chatID int64) (string, error) {
	var buffer *C.char
	if err := parseResult(C.ntg_get_emojis_fingerprint(c.handle, C.int64_t(chatID), &buffer)); err != nil {
		return "", err
	}
	defer C.ntg_string_free(buffer)
	return C.GoString(buffer), nil
}

func (c *Client) GetCallType(chatID int64) (CallType, error) {
	var buffer C.ntg_call_type
	if err := parseResult(C.ntg_get_call_type(c.handle, C.int64_t(chatID), &buffer)); err != nil {
		return CallType(0), err
	}
	switch buffer {
	case C.NTG_CALL_TYPE_GROUP:
		return GroupCallType, nil
	case C.NTG_CALL_TYPE_P2P:
		return P2PCallType, nil
	case C.NTG_CALL_TYPE_CONFERENCE:
		return ConferenceCallType, nil
	default:
		return CallType(0), fmt.Errorf("unknown call type")
	}
}

func (c *Client) CreateCall(chatID int64) (string, error) {
	var buffer *C.char
	if err := parseResult(C.ntg_create_call(c.handle, C.int64_t(chatID), &buffer)); err != nil {
		return "", err
	}
	defer C.ntg_string_free(buffer)
	return C.GoString(buffer), nil
}

func (c *Client) InitPresentation(chatID int64) (string, error) {
	var buffer *C.char
	if err := parseResult(C.ntg_init_presentation(c.handle, C.int64_t(chatID), &buffer)); err != nil {
		return "", err
	}
	defer C.ntg_string_free(buffer)
	return C.GoString(buffer), nil
}

func (c *Client) InitConference(chatID int64, userID int64, lastBlock []byte) (ConferenceJoinParams, error) {
	lastBlockC, lastBlockSize := parseBytes(lastBlock)
	defer freeBytes(lastBlockC)
	var buffer C.ntg_conference_join_params
	if err := parseResult(C.ntg_init_conference(c.handle, C.int64_t(chatID), C.int64_t(userID), lastBlockC, lastBlockSize, &buffer)); err != nil {
		return ConferenceJoinParams{}, err
	}
	defer C.ntg_conference_join_params_free(&buffer)
	return ConferenceJoinParams{
		Payload:   C.GoString(buffer.payload),
		PublicKey: C.GoBytes(unsafe.Pointer(buffer.public_key), C.int(buffer.public_key_len)),
		Block:     C.GoBytes(unsafe.Pointer(buffer.block), C.int(buffer.block_len)),
	}, nil
}

func (c *Client) StopPresentation(chatID int64) error {
	return parseResult(C.ntg_stop_presentation(c.handle, C.int64_t(chatID)))
}

func (c *Client) AddIncomingVideo(chatID, userID int64, endpoint string, ssrcGroups []SSRCGroup) (uint32, error) {
	endpointC := C.CString(endpoint)
	defer C.free(unsafe.Pointer(endpointC))
	rawGroups := parseSSRCGroups(ssrcGroups)
	defer freeSSRCGroups(rawGroups)
	var buffer C.uint32_t
	if err := parseResult(C.ntg_add_incoming_video(c.handle, C.int64_t(chatID), C.int64_t(userID), endpointC, slicePtr(rawGroups), C.size_t(len(rawGroups)), &buffer)); err != nil {
		return 0, err
	}
	return uint32(buffer), nil
}

func (c *Client) RemoveIncomingVideo(chatID int64, endpoint string) (bool, error) {
	endpointC := C.CString(endpoint)
	defer C.free(unsafe.Pointer(endpointC))
	var buffer C.bool
	if err := parseResult(C.ntg_remove_incoming_video(c.handle, C.int64_t(chatID), endpointC, &buffer)); err != nil {
		return false, err
	}
	return bool(buffer), nil
}

func (c *Client) CreateP2PCall(userID int64) error {
	return parseResult(C.ntg_create_p2p_call(c.handle, C.int64_t(userID)))
}

func (c *Client) InitExchange(userID int64, dhConfig DhConfig, gAHash []byte) ([]byte, error) {
	dhConfigC := dhConfig.ParseToC()
	defer freeDhConfig(&dhConfigC)
	gAHashC, gAHashSize := parseBytes(gAHash)
	defer freeBytes(gAHashC)
	var buffer *C.uint8_t
	var size C.size_t
	if err := parseResult(C.ntg_init_exchange(c.handle, C.int64_t(userID), dhConfigC, gAHashC, gAHashSize, &buffer, &size)); err != nil {
		return nil, err
	}
	defer C.ntg_bytes_free(unsafe.Pointer(buffer))
	return C.GoBytes(unsafe.Pointer(buffer), C.int(size)), nil
}

func (c *Client) ExchangeKeys(userID int64, gAB []byte, fingerprint int64) (AuthParams, error) {
	gABC, gABSize := parseBytes(gAB)
	defer freeBytes(gABC)
	var buffer C.ntg_auth_params
	if err := parseResult(C.ntg_exchange_keys(c.handle, C.int64_t(userID), gABC, gABSize, C.int64_t(fingerprint), &buffer)); err != nil {
		return AuthParams{}, err
	}
	defer C.ntg_auth_params_free(&buffer)
	return AuthParams{
		GAOrB:          C.GoBytes(unsafe.Pointer(buffer.g_a_or_b), C.int(buffer.g_a_or_b_len)),
		KeyFingerprint: int64(buffer.key_fingerprint),
	}, nil
}

func (c *Client) SkipExchange(userID int64, encryptionKey []byte, isOutgoing bool) error {
	encryptionKeyC, encryptionKeySize := parseBytes(encryptionKey)
	defer freeBytes(encryptionKeyC)
	return parseResult(C.ntg_skip_exchange(c.handle, C.int64_t(userID), encryptionKeyC, encryptionKeySize, C.bool(isOutgoing)))
}

func (c *Client) ConnectP2P(userID int64, rtcServers []RTCServer, versions []string, p2pAllowed bool, customParameters string) error {
	rawServers := parseRtcServers(rtcServers)
	defer freeRtcServers(rawServers)
	rawVersions := parseStringVectorC(versions)
	defer freeStringVectorC(rawVersions)
	var customParametersC *C.char
	if len(customParameters) > 0 {
		customParametersC = C.CString(customParameters)
		defer C.free(unsafe.Pointer(customParametersC))
	}
	return parseResult(C.ntg_connect_p2p(c.handle, C.int64_t(userID), slicePtr(rawServers), C.size_t(len(rawServers)), slicePtr(rawVersions), C.size_t(len(rawVersions)), C.bool(p2pAllowed), customParametersC))
}

func (c *Client) SendSignalingData(chatID int64, data []byte) error {
	dataC, dataSize := parseBytes(data)
	defer freeBytes(dataC)
	return parseResult(C.ntg_send_signaling_data(c.handle, C.int64_t(chatID), dataC, dataSize))
}

func (c *Client) Connect(chatID int64, params string, isPresentation bool) error {
	paramsC := C.CString(params)
	defer C.free(unsafe.Pointer(paramsC))
	return parseResult(C.ntg_connect(c.handle, C.int64_t(chatID), paramsC, C.bool(isPresentation)))
}

func (c *Client) SetStreamSources(chatID int64, streamMode StreamMode, desc MediaDescription) error {
	descC := desc.ParseToC()
	defer freeMediaDescriptionC(descC)
	return parseResult(C.ntg_set_stream_sources(c.handle, C.int64_t(chatID), streamMode.ParseToC(), descC))
}

func (c *Client) SendExternalFrame(chatID int64, streamDevice StreamDevice, data []byte, frameData FrameData) error {
	dataC, dataSize := parseBytes(data)
	defer freeBytes(dataC)
	return parseResult(C.ntg_send_external_frame(c.handle, C.int64_t(chatID), streamDevice.ParseToC(), dataC, dataSize, frameData.ParseToC()))
}

func (c *Client) SendBroadcastTimestamp(chatID int64, timestamp int64) error {
	return parseResult(C.ntg_send_broadcast_timestamp(c.handle, C.int64_t(chatID), C.int64_t(timestamp)))
}

func (c *Client) SendBroadcastPart(chatID int64, segmentID int64, partID int32, status MediaSegmentStatus, qualityUpdate bool, data []byte) error {
	dataC, dataSize := parseBytes(data)
	defer freeBytes(dataC)
	return parseResult(C.ntg_send_broadcast_part(c.handle, C.int64_t(chatID), C.int64_t(segmentID), C.int32_t(partID), status.ParseToC(), C.bool(qualityUpdate), dataC, dataSize))
}

func (c *Client) UpdateAudioSSRCMappings(chatID int64, mappings []SSRCMapping) error {
	rawMappings := parseSSRCMappings(mappings)
	return parseResult(C.ntg_update_audio_ssrc_mappings(c.handle, C.int64_t(chatID), slicePtr(rawMappings), C.size_t(len(rawMappings))))
}

func (c *Client) ApplyBlocks(chatID int64, subchain int32, nextOffset int32, blocks [][]byte, fromShortPoll bool) error {
	rawBlocks := parseBlocks(blocks)
	defer freeBlocks(rawBlocks)
	return parseResult(C.ntg_apply_blocks(c.handle, C.int64_t(chatID), C.int32_t(subchain), C.int32_t(nextOffset), slicePtr(rawBlocks), C.size_t(len(rawBlocks)), C.bool(fromShortPoll)))
}

func (c *Client) FinishSubchainRequest(chatID int64, subchain int32) error {
	return parseResult(C.ntg_finish_subchain_request(c.handle, C.int64_t(chatID), C.int32_t(subchain)))
}

func (c *Client) Pause(chatID int64) (bool, error) {
	var buffer C.bool
	if err := parseResult(C.ntg_pause(c.handle, C.int64_t(chatID), &buffer)); err != nil {
		return false, err
	}
	return bool(buffer), nil
}

func (c *Client) Resume(chatID int64) (bool, error) {
	var buffer C.bool
	if err := parseResult(C.ntg_resume(c.handle, C.int64_t(chatID), &buffer)); err != nil {
		return false, err
	}
	return bool(buffer), nil
}

func (c *Client) Mute(chatID int64) (bool, error) {
	var buffer C.bool
	if err := parseResult(C.ntg_mute(c.handle, C.int64_t(chatID), &buffer)); err != nil {
		return false, err
	}
	return bool(buffer), nil
}

func (c *Client) Unmute(chatID int64) (bool, error) {
	var buffer C.bool
	if err := parseResult(C.ntg_unmute(c.handle, C.int64_t(chatID), &buffer)); err != nil {
		return false, err
	}
	return bool(buffer), nil
}

func (c *Client) Stop(chatID int64) error {
	return parseResult(C.ntg_stop(c.handle, C.int64_t(chatID)))
}

func (c *Client) Time(chatID int64, streamMode StreamMode) (uint64, error) {
	var buffer C.uint64_t
	if err := parseResult(C.ntg_time(c.handle, C.int64_t(chatID), streamMode.ParseToC(), &buffer)); err != nil {
		return 0, err
	}
	return uint64(buffer), nil
}

func (c *Client) CpuUsage() (float64, error) {
	var buffer C.double
	if err := parseResult(C.ntg_cpu_usage(c.handle, &buffer)); err != nil {
		return 0, err
	}
	return float64(buffer), nil
}

func (c *Client) Calls() (map[int64]CallInfo, error) {
	var buffer *C.ntg_call_info_entry
	var size C.size_t
	if err := parseResult(C.ntg_calls(c.handle, &buffer, &size)); err != nil {
		return nil, err
	}
	defer C.ntg_call_info_entry_free(buffer, size)
	raw := unsafe.Slice(buffer, int(size))
	result := make(map[int64]CallInfo, len(raw))
	for _, entry := range raw {
		result[int64(entry.key)] = CallInfo{
			Playback: parseStreamStatus(entry.value.playback),
			Capture:  parseStreamStatus(entry.value.capture),
		}
	}
	return result, nil
}

func (c *Client) Free() {
	clientRegistry.Lock()
	delete(clientRegistry.clients, c.handle)
	clientRegistry.Unlock()
	C.ntg_instance_destroy(c.handle)
	c.handle = nil
}
