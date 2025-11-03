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
package modules

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"

	"main/config"
	"main/internal/database"
)

func statsHandler(m *telegram.NewMessage) error {
	var sb strings.Builder
	sb.Grow(512)

	sb.WriteString(getSystemStats())
	sb.WriteString(getGoMemStats())
	sb.WriteString(getServerStats())
	sb.WriteString(getServedStats())

	m.Reply(sb.String())
	return telegram.EndGroup
}

// ---- Sub Functions ----

func getSystemStats() string {
	var sb strings.Builder
	sb.WriteString("🔧 <b>System:</b>\n")
	fmt.Fprintf(&sb, "• OS: <code>%s</code>, Arch: <code>%s</code>\n", runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&sb, "• CPUs: <code>%d</code>, Goroutines: <code>%d</code>\n\n", runtime.NumCPU(), runtime.NumGoroutine())
	return sb.String()
}

func getGoMemStats() string {
	var sb strings.Builder
	var memStats runtime.MemStats

	runtime.ReadMemStats(&memStats)
	uptime := time.Since(config.StartTime).Minutes()
	gcPerMin := float64(memStats.NumGC) / uptime

	gcEmoji := "🟢"
	switch {
	case gcPerMin > 20:
		gcEmoji = "🔴"
	case gcPerMin > 10:
		gcEmoji = "🟠"
	}

	sb.WriteString("📦 <b>Internal Memory (Go):</b>\n")
	fmt.Fprintf(&sb, "• Alloc: <code>%d MB</code>\n", memStats.Alloc/1024/1024)
	fmt.Fprintf(&sb, "• Sys: <code>%d MB</code>\n", memStats.Sys/1024/1024)
	fmt.Fprintf(&sb, "• NumGC: <code>%d</code> (%s %.1f/min)\n\n", memStats.NumGC, gcEmoji, gcPerMin)
	return sb.String()
}

func getServerStats() string {
	var sb strings.Builder
	sysMem, _ := mem.VirtualMemory()
	cpuPercent, _ := cpu.Percent(0, false)
	diskStat, _ := disk.Usage("/")

	cpuEmoji := "🟢"
	if len(cpuPercent) > 0 {
		switch {
		case cpuPercent[0] > 70:
			cpuEmoji = "🔴"
		case cpuPercent[0] > 40:
			cpuEmoji = "🟡"
		}
	}

	ramUsagePercent := (float64(sysMem.Used) / float64(sysMem.Total)) * 100
	ramEmoji := "🟢"
	switch {
	case ramUsagePercent > 80:
		ramEmoji = "🔴"
	case ramUsagePercent > 50:
		ramEmoji = "🟡"
	}

	sb.WriteString("💻 <b>Server Stats:</b>\n")
	fmt.Fprintf(&sb, "• CPU Usage: %s <code>%.2f%%</code>\n", cpuEmoji, cpuPercent[0])
	fmt.Fprintf(&sb, "• RAM Usage: %s <code>%.2f GiB</code> | <code>%.2f GiB</code>\n",
		ramEmoji,
		float64(sysMem.Used)/1073741824,
		float64(sysMem.Total)/1073741824,
	)
	fmt.Fprintf(&sb, "• Storage: <code>%.2f GiB</code> | <code>%.2f GiB</code>\n\n",
		float64(diskStat.Used)/1073741824,
		float64(diskStat.Total)/1073741824,
	)
	return sb.String()
}

func getServedStats() string {
	var sb strings.Builder
	servedChats, err1 := database.GetServed()
	servedUsers, err2 := database.GetServed(true)

	sb.WriteString("📊 <b>Served:</b>\n")

	if err1 != nil {
		fmt.Fprintf(&sb, "• Chats: <code>Error: %v</code>\n", err1)
	} else {
		fmt.Fprintf(&sb, "• Chats: <code>%d</code>\n", len(servedChats))
	}

	if err2 != nil {
		fmt.Fprintf(&sb, "• Users: <code>Error: %v</code>\n", err2)
	} else {
		fmt.Fprintf(&sb, "• Users: <code>%d</code>\n", len(servedUsers))
	}
	return sb.String()
}
