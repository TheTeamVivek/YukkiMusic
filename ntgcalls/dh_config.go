package ntgcalls

//#include "ntgcalls.h"
//#include <stdlib.h>
import "C"

import "unsafe"

type DhConfig struct {
	G      int32
	P      []byte
	Random []byte
}

func (ctx *DhConfig) ParseToC() C.ntg_dh_config_struct {
	var x C.ntg_dh_config_struct
	x.g = C.int32_t(ctx.G)
	pC, pSize := parseBytes(ctx.P)
	rC, rSize := parseBytes(ctx.Random)
	x.p = pC
	x.sizeP = pSize
	x.random = rC
	x.sizeRandom = rSize
	return x
}

func freeDhConfig(config *C.ntg_dh_config_struct) {
	if config.p != nil {
		C.free(unsafe.Pointer(config.p))
	}
	if config.random != nil {
		C.free(unsafe.Pointer(config.random))
	}
}
