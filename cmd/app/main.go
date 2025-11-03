/*
 * This file is part of YukkiMusic.
 *
 * YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
 * Copyright (C) 2025 TheTeamVivek
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 */
package main

/*
#cgo CFLAGS: -I../../
#cgo linux LDFLAGS: -L ../../ -lntgcalls -lm -lz
#cgo darwin LDFLAGS: -L ../../ -lntgcalls -lc++ -lz -lbz2 -liconv -framework AVFoundation -framework AudioToolbox -framework CoreAudio -framework QuartzCore -framework CoreMedia -framework VideoToolbox -framework AppKit -framework Metal -framework MetalKit -framework OpenGL -framework IOSurface -framework ScreenCaptureKit

// Currently is supported only dynamically linked library on Windows due to
// https://github.com/golang/go/issues/63903
#cgo windows LDFLAGS: -L../../ -lntgcalls
#include "ntgcalls/ntgcalls.h"
#include "glibc_compatibility.h"
*/
import "C"

import (
	"github.com/Laky-64/gologging"

	"main/config"
	"main/internal/cookies"
	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/modules"
)

func main() {
	gologging.SetLevel(gologging.DebugLevel)
	gologging.GetLogger("webrtc").SetLevel(gologging.WarnLevel)

	l := gologging.GetLogger("Main")

	l.Debug("🔹 Initializing MongoDB...")
	dbCleanup := database.Init(config.MongoURI)
	defer dbCleanup()
	l.Info("✅ Database connected successfully")

	go database.MigrateData(config.MongoURI)

	l.Debug("🔹 Initializing cookies...")
	cookies.Init()
	locales.Init()
	l.Debug("🔹 Initializing clients...")
	cleanup := core.Init(config.ApiID, config.ApiHash, config.Token, config.StringSession, config.LoggerID)
	defer cleanup()
	modules.Init(core.Bot, core.UBot, core.Ntg)
	l.Info("🚀 Bot is started")
	core.Bot.Idle()
}
