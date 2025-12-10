package ntgcalls

//#include "ntgcalls.h"
//#include <stdlib.h>
import "C"
import "unsafe"

type MediaDescription struct {
	Microphone *AudioDescription
	Speaker    *AudioDescription
	Camera     *VideoDescription
	Screen     *VideoDescription
}

func (ctx *MediaDescription) ParseToC() C.ntg_media_description_struct {
	var x C.ntg_media_description_struct

	if ctx.Microphone != nil {
		cMic := C.malloc(C.size_t(unsafe.Sizeof(C.ntg_audio_description_struct{})))
		micStruct := ctx.Microphone.ParseToC()
		*(*C.ntg_audio_description_struct)(cMic) = micStruct
		x.microphone = (*C.ntg_audio_description_struct)(cMic)
	}

	if ctx.Speaker != nil {
		cSpeaker := C.malloc(C.size_t(unsafe.Sizeof(C.ntg_audio_description_struct{})))
		speakerStruct := ctx.Speaker.ParseToC()
		*(*C.ntg_audio_description_struct)(cSpeaker) = speakerStruct
		x.speaker = (*C.ntg_audio_description_struct)(cSpeaker)
	}

	if ctx.Camera != nil {
		cCamera := C.malloc(C.size_t(unsafe.Sizeof(C.ntg_video_description_struct{})))
		cameraStruct := ctx.Camera.ParseToC()
		*(*C.ntg_video_description_struct)(cCamera) = cameraStruct
		x.camera = (*C.ntg_video_description_struct)(cCamera)
	}

	if ctx.Screen != nil {
		cScreen := C.malloc(C.size_t(unsafe.Sizeof(C.ntg_video_description_struct{})))
		screenStruct := ctx.Screen.ParseToC()
		*(*C.ntg_video_description_struct)(cScreen) = screenStruct
		x.screen = (*C.ntg_video_description_struct)(cScreen)
	}

	return x
}

func freeMediaDescriptionC(desc C.ntg_media_description_struct) {
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
