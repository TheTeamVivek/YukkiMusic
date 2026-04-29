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

package modules

import (
	"fmt"
	"runtime"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"

	"main/internal/config"
	"main/internal/database"
	"main/internal/locales"
)

func init() {
	helpTexts["/stats"] = `<i>View detailed bot statistics.</i>

<u>Usage:</u>
<b>/stats</b> — Show statistics

<b>📊 Information Shown:</b>
• System stats (OS, CPU, RAM, disk)
• Go runtime stats (memory, GC)
• Server resources
• Served chats count
• Served users count

<b>🔒 Restrictions:</b>
• <b>Sudo users</b> only`
}

func statsHandler(m *telegram.NewMessage) error {
	chatID := m.ChannelID()
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

	servedChats, err1 := database.ServedChats()
	servedUsers, err2 := database.ServedUsers()

	chatsLine := ""
	if err1 != nil {
		chatsLine = F(chatID, "stats_served_chats_line_err", locales.Arg{
			"error": err1.Error(),
		})
	} else {
		chatsLine = F(chatID, "stats_served_chats_line", locales.Arg{
			"count": len(servedChats),
		})
	}

	usersLine := ""
	if err2 != nil {
		usersLine = F(chatID, "stats_served_users_line_err", locales.Arg{
			"error": err2.Error(),
		})
	} else {
		usersLine = F(chatID, "stats_served_users_line", locales.Arg{
			"count": len(servedUsers),
		})
	}

	m.Reply(F(chatID, "stats_overview", locales.Arg{
		"os":                runtime.GOOS,
		"arch":              runtime.GOARCH,
		"cpus":              runtime.NumCPU(),
		"goroutines":        runtime.NumGoroutine(),
		"alloc":             memStats.Alloc / 1024 / 1024,
		"sys":               memStats.Sys / 1024 / 1024,
		"gc_count":          memStats.NumGC,
		"gc_emoji":          gcEmoji,
		"gc_rate":           fmt.Sprintf("%.1f", gcPerMin),
		"cpu_emoji":         cpuEmoji,
		"cpu":               fmt.Sprintf("%.2f", cpuPercent[0]),
		"ram_emoji":         ramEmoji,
		"ram_used_gib":      fmt.Sprintf("%.2f", float64(sysMem.Used)/1073741824),
		"ram_total_gib":     fmt.Sprintf("%.2f", float64(sysMem.Total)/1073741824),
		"storage_used_gib":  fmt.Sprintf("%.2f", float64(diskStat.Used)/1073741824),
		"storage_total_gib": fmt.Sprintf("%.2f", float64(diskStat.Total)/1073741824),
		"chats_line":        chatsLine,
		"users_line":        usersLine,
	}))
	return telegram.ErrEndGroup
}
