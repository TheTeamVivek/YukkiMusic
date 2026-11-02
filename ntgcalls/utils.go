package ntgcalls

//#include "ntgcalls.h"
//#include <stdlib.h>
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

func slicePtr[T any](data []T) *T {
	if len(data) == 0 {
		return nil
	}
	return &data[0]
}

func parseResult(result C.ntg_result) error {
	if result == C.NTG_OK {
		return nil
	}
	if message := C.GoString(C.ntg_last_error()); len(message) > 0 {
		return errors.New(message)
	}
	return fmt.Errorf("error code: %d", int32(result))
}

func parseConnectionState(state C.ntg_connection_state) ConnectionState {
	switch state {
	case C.NTG_CONNECTION_STATE_CONNECTING:
		return Connecting
	case C.NTG_CONNECTION_STATE_CONNECTED:
		return Connected
	case C.NTG_CONNECTION_STATE_FAILED:
		return Failed
	case C.NTG_CONNECTION_STATE_TIMEOUT:
		return Timeout
	case C.NTG_CONNECTION_STATE_CLOSED:
		return Closed
	}
	return Connecting
}

func parseStreamDevice(device C.ntg_stream_device) StreamDevice {
	switch device {
	case C.NTG_STREAM_DEVICE_MICROPHONE:
		return MicrophoneStream
	case C.NTG_STREAM_DEVICE_SPEAKER:
		return SpeakerStream
	case C.NTG_STREAM_DEVICE_CAMERA:
		return CameraStream
	case C.NTG_STREAM_DEVICE_SCREEN:
		return ScreenStream
	}
	return MicrophoneStream
}

func parseStreamStatus(status C.ntg_stream_status) StreamStatus {
	switch status {
	case C.NTG_STREAM_STATUS_ACTIVE:
		return ActiveStream
	case C.NTG_STREAM_STATUS_PAUSED:
		return PausedStream
	case C.NTG_STREAM_STATUS_IDLING:
		return IdlingStream
	}
	return ActiveStream
}

func parseMediaState(state C.ntg_media_state) MediaState {
	return MediaState{
		Muted:               bool(state.muted),
		VideoPaused:         bool(state.video_paused),
		VideoStopped:        bool(state.video_stopped),
		PresentationPaused:  bool(state.presentation_paused),
		PresentationStopped: bool(state.presentation_stopped),
	}
}

func parseBytes(data []byte) (*C.uint8_t, C.size_t) {
	if len(data) == 0 {
		return nil, 0
	}
	return (*C.uint8_t)(C.CBytes(data)), C.size_t(len(data))
}

func freeBytes(data *C.uint8_t) {
	if data != nil {
		C.free(unsafe.Pointer(data))
	}
}

func parseStringVector(data **C.char, size C.size_t) []string {
	raw := unsafe.Slice(data, int(size))
	result := make([]string, len(raw))
	for i, pointer := range raw {
		result[i] = C.GoString(pointer)
	}
	return result
}

func parseStringVectorC(data []string) []*C.char {
	if len(data) == 0 {
		return nil
	}
	raw := make([]*C.char, len(data))
	for i, value := range data {
		raw[i] = C.CString(value)
	}
	return raw
}

func freeStringVectorC(data []*C.char) {
	for _, value := range data {
		C.free(unsafe.Pointer(value))
	}
}

func parseUint32VectorC(data []uint32) (*C.uint32_t, C.size_t) {
	if len(data) == 0 {
		return nil, 0
	}
	cData := C.malloc(C.size_t(len(data)) * C.size_t(unsafe.Sizeof(C.uint32_t(0))))
	if cData == nil {
		return nil, 0
	}
	value := unsafe.Slice((*C.uint32_t)(cData), len(data))
	for i, item := range data {
		value[i] = C.uint32_t(item)
	}
	return (*C.uint32_t)(cData), C.size_t(len(data))
}

func parseRtcServers(rtcServers []RTCServer) []C.ntg_rtc_server {
	rawServers := make([]C.ntg_rtc_server, len(rtcServers))
	for i, server := range rtcServers {
		peerTagC, peerTagSize := parseBytes(server.PeerTag)
		rawServers[i] = C.ntg_rtc_server{
			id:           C.uint64_t(server.ID),
			ipv4:         C.CString(server.IPv4),
			ipv6:         C.CString(server.IPv6),
			username:     C.CString(server.Username),
			password:     C.CString(server.Password),
			port:         C.uint16_t(server.Port),
			turn:         C.bool(server.TURN),
			stun:         C.bool(server.STUN),
			tcp:          C.bool(server.TCP),
			peer_tag:     peerTagC,
			peer_tag_len: peerTagSize,
		}
	}
	return rawServers
}

func freeRtcServers(rawServers []C.ntg_rtc_server) {
	for _, server := range rawServers {
		C.free(unsafe.Pointer(server.ipv4))
		C.free(unsafe.Pointer(server.ipv6))
		C.free(unsafe.Pointer(server.username))
		C.free(unsafe.Pointer(server.password))
		freeBytes(server.peer_tag)
	}
}

func parseSSRCGroups(groups []SSRCGroup) []C.ntg_ssrc_group {
	rawGroups := make([]C.ntg_ssrc_group, len(groups))
	for i, group := range groups {
		ssrcs, ssrcsLen := parseUint32VectorC(group.SSRCs)
		rawGroups[i] = C.ntg_ssrc_group{
			semantics: C.CString(group.Semantics),
			ssrcs:     ssrcs,
			ssrcs_len: ssrcsLen,
		}
	}
	return rawGroups
}

func freeSSRCGroups(groups []C.ntg_ssrc_group) {
	for _, group := range groups {
		C.free(unsafe.Pointer(group.semantics))
		if group.ssrcs != nil {
			C.free(unsafe.Pointer(group.ssrcs))
		}
	}
}

func parseSSRCMappings(mappings []SSRCMapping) []C.ntg_ssrc_mapping {
	rawMappings := make([]C.ntg_ssrc_mapping, len(mappings))
	for i, mapping := range mappings {
		rawMappings[i] = C.ntg_ssrc_mapping{
			user_id: C.int64_t(mapping.UserID),
			ssrc:    C.int32_t(mapping.SSRC),
		}
	}
	return rawMappings
}

func parseBlocks(blocks [][]byte) []C.ntg_bytes {
	rawBlocks := make([]C.ntg_bytes, len(blocks))
	for i, block := range blocks {
		data, size := parseBytes(block)
		rawBlocks[i] = C.ntg_bytes{
			data: data,
			len:  size,
		}
	}
	return rawBlocks
}

func freeBlocks(blocks []C.ntg_bytes) {
	for _, block := range blocks {
		freeBytes(block.data)
	}
}

func parseDeviceInfoVector(devices *C.ntg_device_info, size C.size_t) []DeviceInfo {
	raw := unsafe.Slice(devices, int(size))
	result := make([]DeviceInfo, len(raw))
	for i, device := range raw {
		result[i] = DeviceInfo{
			Name:     C.GoString(device.name),
			Metadata: C.GoString(device.metadata),
		}
	}
	return result
}
