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
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	td "github.com/AshokShau/gotdbot"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"

	"yukkimusic/config"
	"yukkimusic/internal/database"
	"yukkimusic/internal/utils"
	"yukkimusic/ntgcalls"
)

func init() {
	helpTexts["/stats"] = `<i>View detailed bot statistics.</i>

<u>Usage:</u>
<b>/stats</b> — Show statistics

<b>📊 Information Shown:</b>
• System stats (OS, CPU, RAM, disk)
• Go runtime stats (memory, GC)
• Server resources
• MongoDB stats
• Served chats & users

<b>🔒 Restrictions:</b>
• <b>Sudo users</b> only`
}

func resolveTdVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == "github.com/AshokShau/gotdbot" {
				if dep.Replace != nil && dep.Replace.Version != "" {
					return dep.Replace.Version
				}
				if dep.Version != "" {
					return dep.Version
				}
			}
		}
	}

	return td.Version
}

func statsHandler(c *td.Client, m *td.Message) error {
	if !checkSudo(c, m) {
		return nil
	}

	_, err := m.ReplyRichMessage(c, &td.InputRichMessage{
		Source: &td.RichMessageSourceHtml{Text: buildStats()},
	}, nil)
	return err
}

type statRow struct{ label, value string }

// buildStats renders the bot statistics as a rich HTML message.
func buildStats() string {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	var b strings.Builder
	b.WriteString("<h3>📊 Bot Statistics</h3>\n")
	b.WriteString("<aside><i>Live overview of the host, resources and database.</i></aside>\n\n")

	// System
	kernel, hostname, hostUptime := "-", "-", "-"
	if h, err := host.Info(); err == nil && h != nil {
		kernel, hostname = h.KernelVersion, h.Hostname
		hostUptime = utils.FormatDuration(int(h.Uptime))
	}
	section(&b, "🔧 System",
		statRow{"OS", code(runtime.GOOS + " · " + runtime.GOARCH)},
		statRow{"Kernel", code(kernel)},
		statRow{"Host", code(hostname)},
		statRow{"CPUs", code(strconv.Itoa(runtime.NumCPU()))},
		statRow{"Goroutines", code(strconv.Itoa(runtime.NumGoroutine()))},
		statRow{"Uptime", code(hostUptime)},
	)

	// Memory (Go)
	gcPerMin := float64(memStats.NumGC) / time.Since(config.StartTime).Minutes()
	section(&b, "📦 Memory (Go)",
		statRow{"Alloc", code(strconv.FormatUint(memStats.Alloc/1024/1024, 10) + " MB")},
		statRow{"Heap", code(strconv.FormatUint(memStats.HeapInuse/1024/1024, 10) + " MB")},
		statRow{"Stack", code(strconv.FormatUint(memStats.StackInuse/1024/1024, 10) + " MB")},
		statRow{"Sys", code(strconv.FormatUint(memStats.Sys/1024/1024, 10) + " MB")},
		statRow{"GC runs", code(strconv.FormatUint(uint64(memStats.NumGC), 10)) + " " + emoji(gcPerMin, 10, 20) + " " + code(fmt.Sprintf("%.1f/min", gcPerMin))},
	)

	// Server
	var (
		ramUsed, ramTotal    uint64
		swapUsed, swapTotal  uint64
		diskUsed, diskTotal  uint64
		cpuPct               float64
		load1, load5, load15 float64
	)
	if v, err := mem.VirtualMemory(); err == nil && v != nil {
		ramUsed, ramTotal = v.Used, v.Total
	}
	if v, err := mem.SwapMemory(); err == nil && v != nil {
		swapUsed, swapTotal = v.Used, v.Total
	}
	if d, err := disk.Usage("/"); err == nil && d != nil {
		diskUsed, diskTotal = d.Used, d.Total
	}
	if p, err := cpu.Percent(0, false); err == nil && len(p) > 0 {
		cpuPct = p[0]
	}
	if l, err := load.Avg(); err == nil && l != nil {
		load1, load5, load15 = l.Load1, l.Load5, l.Load15
	}

	var procRSS uint64
	var procPct float64
	var threads int32
	if proc, err := process.NewProcess(int32(os.Getpid())); err == nil && proc != nil {
		if mi, err := proc.MemoryInfo(); err == nil && mi != nil {
			procRSS = mi.RSS
		}
		if p, err := proc.MemoryPercent(); err == nil {
			procPct = float64(p)
		}
		if t, err := proc.NumThreads(); err == nil {
			threads = t
		}
	}

	ramPct := percent(ramUsed, ramTotal)
	section(&b, "💻 Server",
		statRow{"Load Avg", code(fmt.Sprintf("%.2f · %.2f · %.2f", load1, load5, load15))},
		statRow{"CPU", emoji(cpuPct, 40, 70) + " " + code(fmt.Sprintf("%.2f%%", cpuPct))},
		statRow{"RAM", emoji(ramPct, 50, 80) + " " + code(fmt.Sprintf("%s / %s GiB", gib(ramUsed), gib(ramTotal)))},
		statRow{"Swap", code(fmt.Sprintf("%s / %s GiB", gib(swapUsed), gib(swapTotal)))},
		statRow{"Process RAM", emoji(procPct, 20, 50) + " " + code(fmt.Sprintf("%s (%.1f%%)", utils.FormatBytes(int64(procRSS)), procPct))},
		statRow{"Threads", code(strconv.Itoa(int(threads)))},
		statRow{"Disk", code(fmt.Sprintf("%s / %s GiB", gib(diskUsed), gib(diskTotal)))},
	)

	// MongoDB
	db := database.GetMongoDBStats()

	b.WriteString("<details>\n<summary>🌐 MongoDB</summary>\n<table bordered striped compact>\n")
	fmt.Fprintf(&b, "<tr><td>Databases</td><td>%s</td></tr>\n", code(strconv.Itoa(db.TotalDatabases)))
	for _, d := range db.Databases {
		fmt.Fprintf(&b, "<tr><td>%s</td><td>%s</td></tr>\n", code(d.Name), code(utils.FormatBytes(d.SizeOnDisk)))
	}
	b.WriteString("</table>\n</details>\n\n")

	b.WriteString("<details>\n<summary>📂 Collections · " + code(config.DBName) + "</summary>\n<table bordered striped compact>\n")
	for _, coll := range db.CollectionStats {
		fmt.Fprintf(&b, "<tr><td>%s</td><td>%s</td><td>%s</td></tr>\n", code(coll.Name), code(utils.FormatBytes(coll.StorageSize)), code(strconv.FormatInt(coll.Objects, 10)))
	}
	b.WriteString("</table>\n</details>\n\n")

	// Bot Overview
	servedChats, e1 := database.ServedChats()
	servedUsers, e2 := database.ServedUsers()
	sudoers, e3 := database.Sudoers()
	blockedUsers, e4 := database.BlacklistedUsers()
	blockedChats, e5 := database.BlacklistedChats()
	autoLeave, _ := database.AutoLeave()
	loggerEnabled, _ := database.IsLoggerEnabled()
	maintenance, _ := database.GetMaintenance()

	section(&b, "👥 Bot Overview",
		statRow{"Served Chats", code(count(len(servedChats), e1))},
		statRow{"Served Users", code(count(len(servedUsers), e2))},
		statRow{"Sudoers", code(count(len(sudoers), e3))},
		statRow{"Blocked Users", code(count(len(blockedUsers), e4))},
		statRow{"Blocked Chats", code(count(len(blockedChats), e5))},
		statRow{"Auto Leave", boolEmoji(autoLeave)},
		statRow{"Logger", boolEmoji(loggerEnabled)},
		statRow{"Maintenance", boolEmoji(maintenance)},
	)

	// Versions
	section(&b, "🧩 Versions",
		statRow{"Go", code(runtime.Version())},
		statRow{"gotdbot", code(resolveTdVersion())},
		statRow{"TDLib", code(td.TDLibVersion)},
		statRow{"NTgCalls", code(ntgcalls.Version())},
	)

	return b.String()
}

// section appends a collapsible stats block (details/summary + table).
func section(b *strings.Builder, title string, rows ...statRow) {
	fmt.Fprintf(b, "<details>\n<summary>%s</summary>\n<table bordered striped compact>\n", title)
	for _, r := range rows {
		fmt.Fprintf(b, "<tr><td>%s</td><td>%s</td></tr>\n", r.label, r.value)
	}
	b.WriteString("</table>\n</details>\n\n")
}

func code(v string) string { return "<code>" + v + "</code>" }

// count returns n, or "N/A" when the lookup failed.
func count(n int, err error) string {
	if err != nil {
		return "N/A"
	}
	return strconv.Itoa(n)
}

// emoji returns a traffic-light emoji for a value against warn/crit thresholds.
func emoji(v, warn, crit float64) string {
	switch {
	case v > crit:
		return "🔴"
	case v > warn:
		return "🟡"
	default:
		return "🟢"
	}
}

func percent(used, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return float64(used) / float64(total) * 100
}

func gib(v uint64) string { return fmt.Sprintf("%.2f", float64(v)/1073741824) }

func boolEmoji(v bool) string {
	if v {
		return "✅"
	}
	return "❌"
}
