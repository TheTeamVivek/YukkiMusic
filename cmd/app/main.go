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
	_ "net/http/pprof"
	"os"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
	"yukkimusic/internal/modules"
	"yukkimusic/internal/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Log.Errorf("fatal: %v", err)
		os.Exit(1)
	}
}

func run() error {
	defer utils.CloseLog()

	if err := config.Load(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := checkDependencies(); err != nil {
		return fmt.Errorf("dependency check failed: %w", err)
	}

	if err := refreshDirs(); err != nil {
		return fmt.Errorf("failed to refresh directories: %w", err)
	}

	utils.Log.Debug("Initializing MongoDB...")

	closeDB, err := database.Init(config.MongoURI)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer closeDB()

	utils.Log.Info("Database connected successfully")

	if err := locales.Load(); err != nil {
		return fmt.Errorf("failed to load locales: %w", err)
	}

	utils.Log.Debug("Initializing clients...")

	shutdownCore, err := core.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize core: %w", err)
	}
	defer shutdownCore()

	core.F = modules.F

	if err := database.InitAssistantIndexes(core.Assistants.Count()); err != nil {
		return fmt.Errorf("failed to init assistants: %w", err)
	}

	modules.Init(core.Bot, core.Assistants)

	startHTTPServer()

	core.Bot.Idle()
	return nil
}

func startHTTPServer() {
	go func() {
		addr := "0.0.0.0:" + config.Port

		utils.Log.Infof("HTTP server listening on %s", addr)

		if err := http.ListenAndServe(addr, nil); err != nil {
			utils.Log.Errorf("HTTP server failed: %v", err)
		}
	}()
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
