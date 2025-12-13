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
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"

	"main/internal/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

func formatUptime(d time.Duration) string {
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute
	d -= minutes * time.Minute
	seconds := d / time.Second

	result := ""
	if days > 0 {
		result += fmt.Sprintf("%dd ", days)
	}
	if hours > 0 {
		result += fmt.Sprintf("%dh ", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%dm ", minutes)
	}
	result += fmt.Sprintf("%ds", seconds)
	return result
}

func pingHandler(m *tg.NewMessage) error {
	if m.IsPrivate() {
		m.Delete()
		database.AddServed(m.ChannelID(), true)
	} else {
		database.AddServed(m.ChannelID())
	}

	start := time.Now()
	reply, err := m.Respond(F(m.ChatID(), "ping_start"))
	if err != nil {
		return err
	}

	latency := time.Since(start).Milliseconds()
	uptime := time.Since(config.StartTime)
	uptimeStr := formatUptime(uptime)
	ramInfo := "N/A"
	cpuUsage := "N/A"
	diskUsage := "N/A"

	opt := &tg.SendOptions{
		ReplyMarkup: core.SuppMarkup(),
	}
	if config.PingImage != "" {
		opt.Media = config.PingImage
	}

	v, err := mem.VirtualMemory()
	if err == nil {
		usedGB := float64(v.Used) / 1024 / 1024 / 1024
		totalGB := float64(v.Total) / 1024 / 1024 / 1024

		ramInfo = fmt.Sprintf("%.2f / %.2f GB", usedGB, totalGB)
	}

	if percentages, err := cpu.Percent(time.Second, false); err == nil && len(percentages) > 0 {
		cpuUsage = fmt.Sprintf("%.2f%%", percentages[0])
	}

	if d, err := disk.Usage("/"); err == nil {
		usedGB := float64(d.Used) / 1024 / 1024 / 1024
		totalGB := float64(d.Total) / 1024 / 1024 / 1024
		diskUsage = fmt.Sprintf("%.2f / %.2f GB", usedGB, totalGB)
	}

	msg := F(m.ChatID(), "ping_result", locales.Arg{
		"latency":    latency,
		"bot":        utils.MentionHTML(core.BUser),
		"uptime":     uptimeStr,
		"ram_info":   ramInfo,
		"cpu_usage":  cpuUsage,
		"disk_usage": diskUsage,
	})

	reply.Edit(msg, opt)
	return tg.ErrEndGroup
}
