/*
  - This file is part of YukkiMusic.
    *

  - YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
  - Copyright (C) 2025 TheTeamVivek
    *
  - This program is free software: you can redistribute it and/or modify
  - it under the terms of the GNU General Public License as published by
  - the Free Software Foundation, either version 3 of the License, or
  - (at your option) any later version.
    *
  - This program is distributed in the hope that it will be useful,
  - but WITHOUT ANY WARRANTY; without even the implied warranty of
  - MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
  - GNU General Public License for more details.
    *
  - You should have received a copy of the GNU General Public License
  - along with this program. If not, see <https://www.gnu.org/licenses/>.
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
	"os"
	"net/http"
	_ "net/http/pprof"
	"runtime"


	"github.com/Laky-64/gologging"

	"main/internal/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/modules"
)

func main() {
	initLogger()
	defer config.CloseLogging()

	checkFFmpegAndFFprobe()
	refreshCacheAndDownloads()
	startPprofServer()

	gologging.Debug("🔹 Initializing MongoDB...")
	dbCleanup := database.Init(config.MongoURI)
	defer dbCleanup()
	gologging.Info("✅ Database connected successfully")
	gologging.Debug("🔹 Initializing clients...")
	cleanup := core.Init(
		config.ApiID,
		config.ApiHash,
		config.Token,
		config.StringSessions, // list of sessions
		config.SessionType,    // pyrogram / telethon / gogram
		config.LoggerID,
	)
	defer cleanup()

	core.AssistantIndexFunc = database.GetAssistantIndex
	core.GetChatLanguage = database.GetChatLanguage

	if err := database.RebalanceAssistantIndexes(core.Assistants.Count()); err != nil {
		gologging.Fatal("Failed to rebalance Assistants: " + err.Error())
	}

	modules.Init(core.Bot, core.Assistants)
	core.Bot.Idle()
}

func initLogger() {
	gologging.SetLevel(gologging.DebugLevel)
	gologging.SetOutput(config.LogWriter)

	l := gologging.GetLogger("ntgcalls")
	l.SetLevel(gologging.ErrorLevel)
	l.SetOutput(config.LogWriter)

	l = gologging.GetLogger("webrtc")
	l.SetLevel(gologging.ErrorLevel)
	l.SetOutput(config.LogWriter)

	gologging.GetLogger("Database").SetOutput(config.LogWriter)
}

func refreshCacheAndDownloads() error {
	dirs := []string{
		"./cache",
		"./downloads",
	}

	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func startPprofServer() {
	runtime.SetMutexProfileFraction(1)
	runtime.SetBlockProfileRate(1)

	go func() {
		if err := http.ListenAndServe(":6060", nil); err != nil {
			gologging.Error("Failed to start pprof server: " + err.Error())
		}
	}()
}