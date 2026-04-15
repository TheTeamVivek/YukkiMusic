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

	"main/internal/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/modules"
	"main/internal/platforms"
)

func main() {
	cfgCleanup, err := config.Load()

	if err != nil {
		gologging.FatalF(err.Error())
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
		mux := http.NewServeMux()

		mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})

		if config.EnablePprof {
			gologging.Warn("pprof endpoints enabled - do not expose publicly")
			mux.HandleFunc("/debug/pprof/", pprof.Index)
			mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
			mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
			mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
			mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
			mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
			mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
			mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
			mux.Handle("/debug/pprof/block", pprof.Handler("block"))
			mux.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
		}

		server := &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		}

		gologging.Info(fmt.Sprintf("HTTP server listening on %s", addr))
		if err := server.ListenAndServe(); err != nil {
			gologging.Error("HTTP server error: " + err.Error())
		}
	}()
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
