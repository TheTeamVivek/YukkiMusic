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
	Format          int
	Codec           int
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
	MediaSourceDevice
	MediaSourceDesktop
	MediaSourceExternal
)

const (
	FormatPCM Format = iota
	FormatOgg
	FormatI420
	FormatMatroska
)

const (
	CodecPCM Codec = iota
	CodecOpus
	CodecI420
	CodecH264
	CodecVP8
	CodecVP9
	CodecAV1
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
	SegmentStatusDownloading
	SegmentStatusSuccess
)

type CallType int

const (
	GroupCallType CallType = iota
	OutgoingCallType
	IncomingCallType
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
	Codec                      Codec
	Keyframe                   bool
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
	Bitrate      *uint32
	Format       *Format
	MaxDelayMs   *uint32
}

type VideoDescription struct {
	MediaSource   MediaSource
	Input         string
	Width, Height int16
	FPS           uint8
	KeepOpen      bool
	Format        *Format
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

func (f Format) ParseToC() C.ntg_format {
	switch f {
	case FormatOgg:
		return C.NTG_FORMAT_OGG
	case FormatI420:
		return C.NTG_FORMAT_I420
	case FormatMatroska:
		return C.NTG_FORMAT_MATROSKA
	default:
		return C.NTG_FORMAT_PCM
	}
}

func (c Codec) ParseToC() C.ntg_codec {
	switch c {
	case CodecOpus:
		return C.NTG_CODEC_OPUS
	case CodecI420:
		return C.NTG_CODEC_I420
	case CodecH264:
		return C.NTG_CODEC_H264
	case CodecVP8:
		return C.NTG_CODEC_VP8
	case CodecVP9:
		return C.NTG_CODEC_VP9
	case CodecAV1:
		return C.NTG_CODEC_AV1
	default:
		return C.NTG_CODEC_PCM
	}
}

func parseCodec(codec C.ntg_codec) Codec {
	switch codec {
	case C.NTG_CODEC_OPUS:
		return CodecOpus
	case C.NTG_CODEC_I420:
		return CodecI420
	case C.NTG_CODEC_H264:
		return CodecH264
	case C.NTG_CODEC_VP8:
		return CodecVP8
	case C.NTG_CODEC_VP9:
		return CodecVP9
	case C.NTG_CODEC_AV1:
		return CodecAV1
	default:
		return CodecPCM
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
	case SegmentStatusDownloading:
		return C.NTG_MEDIA_SEGMENT_PART_STATUS_DOWNLOADING
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

func freeAudioDescriptionC(desc *C.ntg_audio_description) {
	if desc == nil {
		return
	}
	C.free(unsafe.Pointer(desc.input))
	C.free(unsafe.Pointer(desc.bitrate))
	C.free(unsafe.Pointer(desc.format))
	C.free(unsafe.Pointer(desc.max_delay_ms))
	C.free(unsafe.Pointer(desc))
}

func freeVideoDescriptionC(desc *C.ntg_video_description) {
	if desc == nil {
		return
	}
	C.free(unsafe.Pointer(desc.input))
	C.free(unsafe.Pointer(desc.format))
	C.free(unsafe.Pointer(desc))
}

func freeMediaDescriptionC(desc C.ntg_media_description) {
	freeAudioDescriptionC(desc.microphone)
	freeAudioDescriptionC(desc.speaker)
	freeVideoDescriptionC(desc.camera)
	freeVideoDescriptionC(desc.screen)
}

func (d *AudioDescription) ParseToC() C.ntg_audio_description {
	var x C.ntg_audio_description
	x.media_source = d.MediaSource.ParseToC()
	x.input = C.CString(d.Input)
	x.sample_rate = C.uint32_t(d.SampleRate)
	x.channel_count = C.uint8_t(d.ChannelCount)
	x.keep_open = C.bool(d.KeepOpen)
	if d.Bitrate != nil {
		if ptr := C.malloc(C.size_t(unsafe.Sizeof(C.uint32_t(0)))); ptr != nil {
			*(*C.uint32_t)(ptr) = C.uint32_t(*d.Bitrate)
			x.bitrate = (*C.uint32_t)(ptr)
		}
	}
	if d.Format != nil {
		if ptr := C.malloc(C.size_t(unsafe.Sizeof(C.ntg_format(0)))); ptr != nil {
			*(*C.ntg_format)(ptr) = d.Format.ParseToC()
			x.format = (*C.ntg_format)(ptr)
		}
	}
	if d.MaxDelayMs != nil {
		if ptr := C.malloc(C.size_t(unsafe.Sizeof(C.uint32_t(0)))); ptr != nil {
			*(*C.uint32_t)(ptr) = C.uint32_t(*d.MaxDelayMs)
			x.max_delay_ms = (*C.uint32_t)(ptr)
		}
	}
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
	if d.Format != nil {
		if ptr := C.malloc(C.size_t(unsafe.Sizeof(C.ntg_format(0)))); ptr != nil {
			*(*C.ntg_format)(ptr) = d.Format.ParseToC()
			x.format = (*C.ntg_format)(ptr)
		}
	}
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
	x.codec = d.Codec.ParseToC()
	x.keyframe = C.bool(d.Keyframe)
	return x
}
