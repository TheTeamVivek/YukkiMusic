package ntgcalls

//#include "ntgcalls.h"
//#include <stdlib.h>
import "C"
import "unsafe"

type (
	StreamType      int
	ConnectionMode  int
	ConnectionKind  int
	ConnectionState int
	StreamStatus    int
	StreamMode      int
	StreamDevice    int
	MediaSource     int
)

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
	RTCConnection ConnectionMode = iota
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

type MediaSegmentQuality int

const (
	SegmentQualityNone MediaSegmentQuality = iota - 1
	SegmentQualityThumbnail
	SegmentQualityMedium
	SegmentQualityFull
)

type MediaSegmentStatus int

const (
	SegmentStatusNotReady MediaSegmentStatus = iota
	SegmentStatusResyncNeeded
	SegmentStatusSuccess
)

type CallType int

const (
	GroupCallType CallType = iota
	P2PCallType
	ConferenceCallType
)

type (
	ConnectionChangeCallback    func(chatID int64, state NetworkInfo)
	StreamEndCallback           func(chatID int64, streamType StreamType, streamDevice StreamDevice)
	UpgradeCallback             func(chatID int64, state MediaState)
	SignalCallback              func(chatID int64, data []byte)
	FrameCallback               func(chatID int64, mode StreamMode, device StreamDevice, frames []Frame)
	RemoteSourceCallback        func(chatID int64, source RemoteSource)
	BroadcastTimestampCallback  func(chatID int64)
	BroadcastPartCallback       func(chatID int64, request SegmentPartRequest)
	EmojisCallback              func(chatID int64, emojis string)
	RequestParticipantsCallback func(chatID int64, request ParticipantsRequest)
	OutboundBlockCallback       func(chatID int64, block []byte)
	SubchainRequestCallback     func(chatID int64, request SubchainRequest)
)

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
	SSRC      uint32
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
	FPS           uint8
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
	UDPP2P       bool
	UDPReflector bool
	Versions     []string
}

type RemoteSource struct {
	SSRC   uint32
	State  StreamStatus
	Device StreamDevice
}

type RTCServer struct {
	ID                 int64
	IPv4, IPv6         string
	Username, Password string
	Port               int32
	TURN, STUN, TCP    bool
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

type SSRCGroup struct {
	Semantics string
	SSRCs     []uint32
}

type ConferenceJoinParams struct {
	Payload   string
	PublicKey []byte
	Block     []byte
}

type ParticipantsRequest struct {
	SSRCs []uint32
}

type SSRCMapping struct {
	UserID int64
	SSRC   int32
}

type SubchainRequest struct {
	Subchain int32
	Height   int32
	Limit    int32
}

func (m MediaSource) ParseToC() C.ntg_media_source {
	switch m {
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

func (m StreamMode) ParseToC() C.ntg_stream_mode {
	switch m {
	case CaptureStream:
		return C.NTG_STREAM_MODE_CAPTURE
	case PlaybackStream:
		return C.NTG_STREAM_MODE_PLAYBACK
	default:
		return C.NTG_STREAM_MODE_CAPTURE
	}
}

func (d StreamDevice) ParseToC() C.ntg_stream_device {
	switch d {
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

func (s MediaSegmentStatus) ParseToC() C.ntg_media_segment_part_status {
	switch s {
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

func (d *MediaDescription) ParseToC() C.ntg_media_description {
	var x C.ntg_media_description

	if d.Microphone != nil {
		if cMic := C.malloc(C.size_t(unsafe.Sizeof(C.ntg_audio_description{}))); cMic != nil {
			*(*C.ntg_audio_description)(cMic) = d.Microphone.ParseToC()
			x.microphone = (*C.ntg_audio_description)(cMic)
		}
	}

	if d.Speaker != nil {
		if cSpeaker := C.malloc(C.size_t(unsafe.Sizeof(C.ntg_audio_description{}))); cSpeaker != nil {
			*(*C.ntg_audio_description)(cSpeaker) = d.Speaker.ParseToC()
			x.speaker = (*C.ntg_audio_description)(cSpeaker)
		}
	}

	if d.Camera != nil {
		if cCamera := C.malloc(C.size_t(unsafe.Sizeof(C.ntg_video_description{}))); cCamera != nil {
			*(*C.ntg_video_description)(cCamera) = d.Camera.ParseToC()
			x.camera = (*C.ntg_video_description)(cCamera)
		}
	}

	if d.Screen != nil {
		if cScreen := C.malloc(C.size_t(unsafe.Sizeof(C.ntg_video_description{}))); cScreen != nil {
			*(*C.ntg_video_description)(cScreen) = d.Screen.ParseToC()
			x.screen = (*C.ntg_video_description)(cScreen)
		}
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

func (d *AudioDescription) ParseToC() C.ntg_audio_description {
	var x C.ntg_audio_description
	x.media_source = d.MediaSource.ParseToC()
	x.input = C.CString(d.Input)
	x.sample_rate = C.uint32_t(d.SampleRate)
	x.channel_count = C.uint8_t(d.ChannelCount)
	x.keep_open = C.bool(d.KeepOpen)
	return x
}

func (d *VideoDescription) ParseToC() C.ntg_video_description {
	var x C.ntg_video_description
	x.media_source = d.MediaSource.ParseToC()
	x.input = C.CString(d.Input)
	x.width = C.int16_t(d.Width)
	x.height = C.int16_t(d.Height)
	x.fps = C.uint8_t(d.FPS)
	x.keep_open = C.bool(d.KeepOpen)
	return x
}

func (c *DhConfig) ParseToC() C.ntg_dh_config {
	var x C.ntg_dh_config
	x.g = C.int32_t(c.G)
	x.p, x.p_len = parseBytes(c.P)
	x.random, x.random_len = parseBytes(c.Random)
	return x
}

func freeDhConfig(config *C.ntg_dh_config) {
	freeBytes(config.p)
	freeBytes(config.random)
}

func (d FrameData) ParseToC() C.ntg_frame_data {
	var x C.ntg_frame_data
	x.absolute_capture_timestamp_ms = C.int64_t(d.AbsoluteCaptureTimestampMs)
	x.width = C.uint16_t(d.Width)
	x.height = C.uint16_t(d.Height)
	x.rotation = C.ntg_video_rotation(d.Rotation)
	return x
}
