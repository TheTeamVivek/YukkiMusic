package ntgcalls

//#include "ntgcalls.h"
//#include <stdlib.h>
import "C"
import "unsafe"

type StreamType int
type ConnectionMode int
type ConnectionKind int
type ConnectionState int
type StreamStatus int
type StreamMode int
type StreamDevice int
type MediaSource int
type MediaSegmentQuality int
type MediaSegmentStatus int
type CallType int

type StreamEndCallback func(chatId int64, streamType StreamType, streamDevice StreamDevice)
type UpgradeCallback func(chatId int64, state MediaState)
type ConnectionChangeCallback func(chatId int64, state NetworkInfo)
type SignalCallback func(chatId int64, signal []byte)
type FrameCallback func(chatId int64, mode StreamMode, device StreamDevice, frames []Frame)
type RemoteSourceCallback func(chatId int64, source RemoteSource)
type BroadcastTimestampCallback func(chatId int64)

type BroadcastPartCallback func(chatId int64, segmentPartRequest SegmentPartRequest)

type EmojisCallback func(chatId int64, emojis string)
type RequestParticipantsCallback func(chatId int64, request ParticipantsRequest)
type OutboundBlockCallback func(chatId int64, block []byte)
type SubchainRequestCallback func(chatId int64, subchainRequest SubchainRequest)

const (
	MicrophoneStream StreamDevice = iota
	SpeakerStream
	CameraStream
	ScreenStream
)

const (
	AudioStream StreamType = iota
	VideoStream
)

const (
	MediaSourceFile MediaSource = 1 << iota
	MediaSourceShell
	MediaSourceFFmpeg
	MediaSourceDevice
	MediaSourceDesktop
	MediaSourceExternal
)

const (
	ActiveStream StreamStatus = iota
	PausedStream
	IdlingStream
)

const (
	RtcConnection ConnectionMode = iota
	StreamConnection
	RTMPConnection
)

const (
	Connecting ConnectionState = iota
	Connected
	Failed
	Timeout
	Closed
)

const (
	NormalConnection ConnectionKind = iota
	PresentationConnection
)

const (
	CaptureStream StreamMode = iota
	PlaybackStream
)

const (
	SegmentQualityNone MediaSegmentQuality = iota - 1
	SegmentQualityThumbnail
	SegmentQualityMedium
	SegmentQualityFull
)

const (
	SegmentStatusNotReady MediaSegmentStatus = iota
	SegmentStatusResyncNeeded
	SegmentStatusSuccess
)

const (
	GroupCallType CallType = iota
	P2PCallType
	ConferenceCallType
)

func (ctx MediaSource) ParseToC() C.ntg_media_source {
	switch ctx {
	case MediaSourceFile:
		return C.NTG_MEDIA_SOURCE_FILE
	case MediaSourceShell:
		return C.NTG_MEDIA_SOURCE_SHELL
	case MediaSourceFFmpeg:
		return C.NTG_MEDIA_SOURCE_FFMPEG
	case MediaSourceDevice:
		return C.NTG_MEDIA_SOURCE_DEVICE
	case MediaSourceDesktop:
		return C.NTG_MEDIA_SOURCE_DESKTOP
	case MediaSourceExternal:
		return C.NTG_MEDIA_SOURCE_EXTERNAL
	default:
		return C.NTG_MEDIA_SOURCE_FILE
	}
}

func (ctx StreamMode) ParseToC() C.ntg_stream_mode {
	switch ctx {
	case CaptureStream:
		return C.NTG_STREAM_MODE_CAPTURE
	case PlaybackStream:
		return C.NTG_STREAM_MODE_PLAYBACK
	default:
		return C.NTG_STREAM_MODE_CAPTURE
	}
}

func (ctx StreamDevice) ParseToC() C.ntg_stream_device {
	switch ctx {
	case MicrophoneStream:
		return C.NTG_STREAM_DEVICE_MICROPHONE
	case SpeakerStream:
		return C.NTG_STREAM_DEVICE_SPEAKER
	case CameraStream:
		return C.NTG_STREAM_DEVICE_CAMERA
	case ScreenStream:
		return C.NTG_STREAM_DEVICE_SCREEN
	default:
		return C.NTG_STREAM_DEVICE_MICROPHONE
	}
}

func (ctx MediaSegmentStatus) ParseToC() C.ntg_media_segment_part_status {
	switch ctx {
	case SegmentStatusNotReady:
		return C.NTG_MEDIA_SEGMENT_PART_STATUS_NOT_READY
	case SegmentStatusResyncNeeded:
		return C.NTG_MEDIA_SEGMENT_PART_STATUS_RESYNC_NEEDED
	case SegmentStatusSuccess:
		return C.NTG_MEDIA_SEGMENT_PART_STATUS_SUCCESS
	default:
		return C.NTG_MEDIA_SEGMENT_PART_STATUS_NOT_READY
	}
}

type AuthParams struct {
	GAOrB          []byte
	KeyFingerprint int64
}

type CallInfo struct {
	Playback, Capture StreamStatus
}

type DeviceInfo struct {
	Name     string
	Metadata string
}

type Frame struct {
	Ssrc      uint32
	Data      []byte
	FrameData FrameData
}

type FrameData struct {
	AbsoluteCaptureTimestampMs int64
	Width                      uint16
	Height                     uint16
	Rotation                   uint16
}

type MediaDescription struct {
	Microphone *AudioDescription
	Speaker    *AudioDescription
	Camera     *VideoDescription
	Screen     *VideoDescription
}

type AudioDescription struct {
	MediaSource  MediaSource
	Input        string
	SampleRate   uint32
	ChannelCount uint8
	KeepOpen     bool
}

type VideoDescription struct {
	MediaSource   MediaSource
	Input         string
	Width, Height int16
	Fps           uint8
	KeepOpen      bool
}

type DhConfig struct {
	G      int32
	P      []byte
	Random []byte
}

type MediaDevices struct {
	Microphone []DeviceInfo
	Speaker    []DeviceInfo
	Camera     []DeviceInfo
	Screen     []DeviceInfo
}

type MediaState struct {
	Muted               bool
	VideoPaused         bool
	VideoStopped        bool
	PresentationPaused  bool
	PresentationStopped bool
}

type NetworkInfo struct {
	Kind  ConnectionKind
	State ConnectionState
}

type Protocol struct {
	MinLayer     int32
	MaxLayer     int32
	UdpP2P       bool
	UdpReflector bool
	Versions     []string
}

type RemoteSource struct {
	Ssrc   uint32
	State  StreamStatus
	Device StreamDevice
}

type RTCServer struct {
	ID                 int64
	Ipv4, Ipv6         string
	Username, Password string
	Port               int32
	Turn, Stun, Tcp    bool
	PeerTag            []byte
}

type SegmentPartRequest struct {
	SegmentID     int64
	PartID        int32
	Limit         int32
	Timestamp     int64
	QualityUpdate bool
	ChannelID     int32
	Quality       MediaSegmentQuality
}

type SsrcGroup struct {
	Semantics string
	Ssrcs     []uint32
}

type ConferenceJoinParams struct {
	Payload   string
	PublicKey []byte
	Block     []byte
}

type ParticipantsRequest struct {
	Ssrcs []uint32
}

type SsrcMapping struct {
	UserID int64
	Ssrc   int32
}

type SubchainRequest struct {
	Subchain int32
	Height   int32
	Limit    int32
}

func (ctx *MediaDescription) ParseToC() C.ntg_media_description {
	var x C.ntg_media_description

	if ctx.Microphone != nil {
		cMic := C.malloc(
			C.size_t(unsafe.Sizeof(C.ntg_audio_description{})),
		)
		micStruct := ctx.Microphone.ParseToC()
		*(*C.ntg_audio_description)(cMic) = micStruct
		x.microphone = (*C.ntg_audio_description)(cMic)
	}

	if ctx.Speaker != nil {
		cSpeaker := C.malloc(
			C.size_t(unsafe.Sizeof(C.ntg_audio_description{})),
		)
		speakerStruct := ctx.Speaker.ParseToC()
		*(*C.ntg_audio_description)(cSpeaker) = speakerStruct
		x.speaker = (*C.ntg_audio_description)(cSpeaker)
	}

	if ctx.Camera != nil {
		cCamera := C.malloc(
			C.size_t(unsafe.Sizeof(C.ntg_video_description{})),
		)
		cameraStruct := ctx.Camera.ParseToC()
		*(*C.ntg_video_description)(cCamera) = cameraStruct
		x.camera = (*C.ntg_video_description)(cCamera)
	}

	if ctx.Screen != nil {
		cScreen := C.malloc(
			C.size_t(unsafe.Sizeof(C.ntg_video_description{})),
		)
		screenStruct := ctx.Screen.ParseToC()
		*(*C.ntg_video_description)(cScreen) = screenStruct
		x.screen = (*C.ntg_video_description)(cScreen)
	}

	return x
}

func freeMediaDescriptionC(desc C.ntg_media_description) {
	if desc.microphone != nil {
		C.free(unsafe.Pointer(desc.microphone.input))
		C.free(unsafe.Pointer(desc.microphone))
	}
	if desc.speaker != nil {
		C.free(unsafe.Pointer(desc.speaker.input))
		C.free(unsafe.Pointer(desc.speaker))
	}
	if desc.camera != nil {
		C.free(unsafe.Pointer(desc.camera.input))
		C.free(unsafe.Pointer(desc.camera))
	}
	if desc.screen != nil {
		C.free(unsafe.Pointer(desc.screen.input))
		C.free(unsafe.Pointer(desc.screen))
	}
}

func (ctx *AudioDescription) ParseToC() C.ntg_audio_description {
	var x C.ntg_audio_description
	x.media_source = ctx.MediaSource.ParseToC()
	x.input = C.CString(ctx.Input)
	x.sample_rate = C.uint32_t(ctx.SampleRate)
	x.channel_count = C.uint8_t(ctx.ChannelCount)
	x.keep_open = C.bool(ctx.KeepOpen)
	return x
}

func (ctx *VideoDescription) ParseToC() C.ntg_video_description {
	var x C.ntg_video_description
	x.media_source = ctx.MediaSource.ParseToC()
	x.input = C.CString(ctx.Input)
	x.width = C.int16_t(ctx.Width)
	x.height = C.int16_t(ctx.Height)
	x.fps = C.uint8_t(ctx.Fps)
	x.keep_open = C.bool(ctx.KeepOpen)
	return x
}

func (ctx *DhConfig) ParseToC() C.ntg_dh_config {
	var x C.ntg_dh_config
	x.g = C.int32_t(ctx.G)
	pC, pSize := parseBytes(ctx.P)
	rC, rSize := parseBytes(ctx.Random)
	x.p = pC
	x.p_len = pSize
	x.random = rC
	x.random_len = rSize
	return x
}

func freeDhConfig(config *C.ntg_dh_config) {
	if config.p != nil {
		C.free(unsafe.Pointer(config.p))
	}
	if config.random != nil {
		C.free(unsafe.Pointer(config.random))
	}
}

func (ctx *FrameData) ParseToC() C.ntg_frame_data {
	var x C.ntg_frame_data
	x.absolute_capture_timestamp_ms = C.int64_t(ctx.AbsoluteCaptureTimestampMs)
	x.width = C.uint16_t(ctx.Width)
	x.height = C.uint16_t(ctx.Height)
	x.rotation = C.ntg_video_rotation(ctx.Rotation)
	return x
}
