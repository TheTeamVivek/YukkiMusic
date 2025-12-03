package main

import (
	"os/exec"
)

func checkFFmpegAndFFprobe() {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		gologging.Fatal("❌ ffmpeg not found in PATH. Please install ffmpeg.")
	}

	if _, err := exec.LookPath("ffprobe"); err != nil {
		gologging.Fatal("❌ ffprobe not found in PATH. Please install ffprobe.")
	}
}