/*
 * ● YukkiMusic
 * ○ A high-performance engine for streaming music in Telegram voicechats.
 *
 * Copyright (C) 2026 TheTeamVivek
 *
 * This program is free software: you can redistribute it and/or modify it under the
 * terms of the GNU General Public License as published by the Free Software Foundation,
 * either version 3 of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
 * PARTICULAR PURPOSE. See the GNU General Public License for more details.
 *
 * Repository: https://github.com/TheTeamVivek/YukkiMusic
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
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"github.com/Laky-64/gologging"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/modules"
	"yukkimusic/internal/platforms"
)

func main() {
	cfgCleanup, err := config.Load()
	if err != nil {
		gologging.Fatal(err.Error())
		return
	}
	defer cfgCleanup()
	initLogger()

	shutdownPlatforms, err := platforms.Init()
	if err != nil {
		gologging.Fatal("Failed to initialize platforms: " + err.Error())
	}
	defer shutdownPlatforms()

	checkFFmpegAndFFprobe()

	if err := refreshDirs(); err != nil {
		gologging.Fatal("Failed to refresh directories: " + err.Error())
	}

	gologging.Debug("Initializing MongoDB...")

	closeDB, err := database.Init(config.MongoURI)
	if err != nil {
		gologging.Fatal("Failed to initialize database: " + err.Error())
	}
	defer closeDB()

	gologging.Info("Database connected successfully")

	if err := locales.Load(); err != nil {
		gologging.Fatal("Failed to load locales: " + err.Error())
	}

	gologging.Debug("Initializing clients...")

	shutdownCore, err := core.Init()
	if err != nil {
		gologging.Fatal("Failed to initialize core: " + err.Error())
	}
	defer shutdownCore()

	core.GetAssistantIndexFunc = database.AssistantIndex
	core.F = modules.F

	if err := database.RebalanceAssistantIndexes(core.Assistants.Count()); err != nil {
		gologging.Fatal("Failed to rebalance Assistants: " + err.Error())
	}

	modules.Init(core.Bot, core.Assistants)

	startHTTPServer()

	core.Bot.Idle()
}

func startHTTPServer() {
	go func() {
		addr := "0.0.0.0:" + config.Port

		gologging.InfoF("HTTP server listening on %s", addr)

		if err := http.ListenAndServe(addr, nil); err != nil {
			gologging.FatalF("HTTP server failed: %v", err)
		}
	}()
}

func initLogger() {
	gologging.SetLevel(gologging.DebugLevel)
	gologging.SetOutput(config.LogWriter)

	for _, name := range []string{
		"ntgcalls",
		"webrtc",
	} {
		l := gologging.GetLogger(name)
		l.SetLevel(gologging.ErrorLevel)
		l.SetOutput(config.LogWriter)
	}

	gologging.GetLogger("Database").SetOutput(config.LogWriter)
}

func refreshDirs() error {
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
