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
	"time"

	td "github.com/AshokShau/gotdbot"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"yukkimusic/internal/logger"

	"yukkimusic/config"
	"yukkimusic/internal/core"
	"yukkimusic/internal/database"
	"yukkimusic/internal/locales"
)

func init() {
	helpTexts["/ping"] = `<i>Check bot responsiveness and system stats.</i>

<u>Usage:</u>
<b>/ping</b> — Get bot status

<b>📊 Information Shown:</b>
• Response latency (ms)
• Uptime
• RAM usage
• CPU usage
• Disk usage

<b>💡 Use Case:</b>
Check if bot is responsive and view system health.`
}

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

func pingHandler(c *td.Client, m *td.Message) error {
	if isChannel(c, m) {
		return nil
	}

	chatID := m.ChatID()

	if m.IsPrivate() {
		m.Delete(c, true)
		database.AddServedUser(chatID)
	} else {
		database.AddServedChat(chatID)
	}

	start := time.Now()
	reply, err := m.ReplyText(c, F(chatID, "ping_start"), nil)
	if err != nil {
		return err
	}

	latency := time.Since(start).Milliseconds()
	uptime := time.Since(config.StartTime)
	uptimeStr := formatUptime(uptime)
	ramInfo := "N/A"
	cpuUsage := "N/A"
	diskUsage := "N/A"

	v, err := mem.VirtualMemory()
	if err == nil {
		usedGB := float64(v.Used) / 1024 / 1024 / 1024
		totalGB := float64(v.Total) / 1024 / 1024 / 1024

		ramInfo = fmt.Sprintf("%.2f / %.2f GB", usedGB, totalGB)
	}

	if percentages, err := cpu.Percent(time.Second, false); err == nil &&
		len(percentages) > 0 {
		cpuUsage = fmt.Sprintf("%.2f%%", percentages[0])
	}

	if d, err := disk.Usage("/"); err == nil {
		usedGB := float64(d.Used) / 1024 / 1024 / 1024
		totalGB := float64(d.Total) / 1024 / 1024 / 1024
		diskUsage = fmt.Sprintf("%.2f / %.2f GB", usedGB, totalGB)
	}

	msg := F(chatID, "ping_result", locales.Arg{
		"latency":    latency,
		"bot":        td.Mention(c.Me.FirstName, c.Me.Id, true, true),
		"uptime":     uptimeStr,
		"ram_info":   ramInfo,
		"cpu_usage":  cpuUsage,
		"disk_usage": diskUsage,
	})

	markup := core.SuppMarkup(chatID)

	edit := func() error {
		if config.PingImage == "" {
			_, err := reply.EditText(c, msg, &td.EditTextMessageOpts{
				ParseMode:   td.ParseModeHTML,
				ReplyMarkup: markup,
			})
			return err
		}

		caption, err := c.GetFormattedText(msg, nil, td.ParseModeHTML)
		if err != nil {
			return err
		}

		_, err = reply.EditMedia(c, &td.InputMessagePhoto{
			Photo: &td.InputPhoto{
				Photo: td.InputFileRemote{Id: config.PingImage},
			},
			Caption: caption,
		}, &td.EditMessageMediaOpts{
			ReplyMarkup: markup,
		})
		return err
	}

	if err := edit(); err != nil {
		logger.Errorf("[ping] edit failed: %v", err)

		if config.PingImage != "" {
			_, err = reply.EditText(c, msg, &td.EditTextMessageOpts{
				ParseMode:   td.ParseModeHTML,
				ReplyMarkup: markup,
			})
			if err != nil {
				logger.Errorf("[ping] fallback text edit failed: %v", err)
				return err
			}
		}
	}
	return nil
}
