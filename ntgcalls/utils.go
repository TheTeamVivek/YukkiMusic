package ntgcalls

//#include "ntgcalls.h"
//#include <stdlib.h>
import "C"

import (
        "fmt"
        "unsafe"
)

func parseConnectionState(state C.ntg_connection_state_enum) ConnectionState {
        switch state {
        case C.NTG_STATE_CONNECTING:
                return Connecting
        case C.NTG_STATE_CONNECTED:
                return Connected
        case C.NTG_STATE_FAILED:
                return Failed
        case C.NTG_STATE_TIMEOUT:
                return Timeout
        case C.NTG_STATE_CLOSED:
                return Closed
        }
        return Connecting
}

func parseStreamDevice(device C.ntg_stream_device_enum) StreamDevice {
        var goDevice StreamDevice
        switch device {
        case C.NTG_STREAM_MICROPHONE:
                goDevice = MicrophoneStream
        case C.NTG_STREAM_SPEAKER:
                goDevice = SpeakerStream
        case C.NTG_STREAM_CAMERA:
                goDevice = CameraStream
        case C.NTG_STREAM_SCREEN:
                goDevice = ScreenStream
        }
        return goDevice
}

func parseBool(futureResult *Future) (bool, error) {
        return *futureResult.errCode == 0, parseErrorCode(futureResult)
}

func parseBytes(data []byte) (*C.uint8_t, C.int) {
        if len(data) > 0 {
                return (*C.uint8_t)(C.CBytes(data)), C.int(len(data))
        }
        return nil, 0
}

func parseStringVector(data unsafe.Pointer, size C.int) []string {
        result := make([]string, size)
        for i := 0; i < int(size); i++ {
                pointer := *(**C.char)(unsafe.Pointer(uintptr(data) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
                result[i] = C.GoString(pointer)
                C.free(unsafe.Pointer(pointer))
        }
        defer C.free(data)
        return result
}

func parseUint32VectorC(data []uint32) (*C.uint32_t, C.int) {
        if len(data) > 0 {
                cData := C.malloc(C.size_t(len(data)) * C.size_t(unsafe.Sizeof(C.uint32_t(0))))
                if cData == nil {
                        return nil, 0
                }
                ssrcs := (*C.uint32_t)(cData)
                for i, v := range data {
                        *(*C.uint32_t)(unsafe.Pointer(uintptr(unsafe.Pointer(ssrcs)) + uintptr(i)*unsafe.Sizeof(C.uint32_t(0)))) = C.uint32_t(v)
                }
                return ssrcs, C.int(len(data))
        }
        return nil, 0
}

func parseStringVectorC(data []string) (**C.char, C.int) {
        if len(data) == 0 {
                return nil, 0
        }
        cArray := C.malloc(C.size_t(len(data)) * C.size_t(unsafe.Sizeof(uintptr(0))))
        goSlice := (*[1 << 30]*C.char)(cArray)[:len(data):len(data)]
        for i, v := range data {
                goSlice[i] = C.CString(v)
        }
        return (**C.char)(cArray), C.int(len(data))
}

func freeStringVectorC(data **C.char, size C.int) {
        if data == nil {
                return
        }
        goSlice := (*[1 << 30]*C.char)(unsafe.Pointer(data))[:size:size]
        for i := 0; i < int(size); i++ {
                C.free(unsafe.Pointer(goSlice[i]))
        }
        C.free(unsafe.Pointer(data))
}

func parseErrorCode(futureResult *Future) error {
        errorCode := int32(*futureResult.errCode)
        if errorCode < 0 {
                var message string
                if futureResult.errMessage != nil {
                        cMessage := *futureResult.errMessage
                        if cMessage != nil {
                                defer C.free(unsafe.Pointer(cMessage))
                                message = C.GoString(cMessage)
                        }
                }
                if len(message) == 0 {
                        message = fmt.Sprintf("Error code: %d", errorCode)
                }
                return fmt.Errorf("%s", message)
        }
        return nil
}

func parseStreamStatus(status C.ntg_stream_status_enum) StreamStatus {
        switch status {
        case C.NTG_ACTIVE:
                return ActiveStream
        case C.NTG_PAUSED:
                return PausedStream
        case C.NTG_IDLING:
                return IdlingStream
        }
        return ActiveStream
}

func parseRtcServers(rtcServers []RTCServer) *C.ntg_rtc_server_struct {
        if len(rtcServers) == 0 {
                return nil
        }
        cArray := C.malloc(C.size_t(len(rtcServers)) * C.size_t(unsafe.Sizeof(C.ntg_rtc_server_struct{})))
        goSlice := (*[1 << 30]C.ntg_rtc_server_struct)(cArray)[:len(rtcServers):len(rtcServers)]
        for i, server := range rtcServers {
                goSlice[i] = C.ntg_rtc_server_struct{
                        ipv4:        C.CString(server.Ipv4),
                        ipv6:        C.CString(server.Ipv6),
                        username:    C.CString(server.Username),
                        password:    C.CString(server.Password),
                        port:        C.uint16_t(server.Port),
                        turn:        C.bool(server.Turn),
                        stun:        C.bool(server.Stun),
                        tcp:         C.bool(server.Tcp),
                        peerTag:     nil,
                        peerTagSize: 0,
                }
                if len(server.PeerTag) > 0 {
                        peerTagC, peerTagSize := parseBytes(server.PeerTag)
                        goSlice[i].peerTag = peerTagC
                        goSlice[i].peerTagSize = peerTagSize
                }
        }
        return (*C.ntg_rtc_server_struct)(cArray)
}

func freeRtcServers(servers *C.ntg_rtc_server_struct, size C.int) {
        if servers == nil {
                return
        }
        goSlice := (*[1 << 30]C.ntg_rtc_server_struct)(unsafe.Pointer(servers))[:size:size]
        for i := 0; i < int(size); i++ {
                C.free(unsafe.Pointer(goSlice[i].ipv4))
                C.free(unsafe.Pointer(goSlice[i].ipv6))
                C.free(unsafe.Pointer(goSlice[i].username))
                C.free(unsafe.Pointer(goSlice[i].password))
                if goSlice[i].peerTag != nil {
                        C.free(unsafe.Pointer(goSlice[i].peerTag))
                }
        }
        C.free(unsafe.Pointer(servers))
}

func parseSsrcGroups(ssrcGroups []SsrcGroup) *C.ntg_ssrc_group_struct {
        if len(ssrcGroups) == 0 {
                return nil
        }
        cArray := C.malloc(C.size_t(len(ssrcGroups)) * C.size_t(unsafe.Sizeof(C.ntg_ssrc_group_struct{})))
        goSlice := (*[1 << 30]C.ntg_ssrc_group_struct)(cArray)[:len(ssrcGroups):len(ssrcGroups)]
        for i, group := range ssrcGroups {
                ssrcsC, sizeSsrcs := parseUint32VectorC(group.Ssrcs)
                goSlice[i] = C.ntg_ssrc_group_struct{
                        semantics: C.CString(group.Semantics),
                        ssrcs:     ssrcsC,
                        sizeSsrcs: sizeSsrcs,
                }
        }
        return (*C.ntg_ssrc_group_struct)(cArray)
}

func freeSsrcGroups(groups *C.ntg_ssrc_group_struct, size C.int) {
        if groups == nil {
                return
        }
        goSlice := (*[1 << 30]C.ntg_ssrc_group_struct)(unsafe.Pointer(groups))[:size:size]
        for i := 0; i < int(size); i++ {
                C.free(unsafe.Pointer(goSlice[i].semantics))
                if goSlice[i].ssrcs != nil {
                        C.free(unsafe.Pointer(goSlice[i].ssrcs))
                }
        }
        C.free(unsafe.Pointer(groups))
}

func parseDeviceInfoVector(devices unsafe.Pointer, size C.int) []DeviceInfo {
        rawDevices := make([]DeviceInfo, size)
        for i := 0; i < int(size); i++ {
                device := *(*C.ntg_device_info_struct)(unsafe.Pointer(uintptr(devices) + uintptr(i)*unsafe.Sizeof(C.ntg_device_info_struct{})))
                rawDevices[i] = DeviceInfo{
                        Name:     C.GoString(device.name),
                        Metadata: C.GoString(device.metadata),
                }
                C.free(unsafe.Pointer(device.name))
                C.free(unsafe.Pointer(device.metadata))
        }
        defer C.free(devices)
        return rawDevices
}