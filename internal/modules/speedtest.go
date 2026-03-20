/*
 * ● YukkiMusic
 * ○ A high-performance engine for streaming music in Telegram voicechats.
 * ________________________________________________________________________________________
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
 * ________________________________________________________________________________________
 */

package modules

import (
	"fmt"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/showwin/speedtest-go/speedtest"

	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/speedtest"] = `<i>Run server network speed test.</i>

<u>Usage:</u>
<b>/speedtest</b> or <b>/spt</b> — Test network speed

<b>📊 Results Include:</b>
• Download speed (Mbps)
• Upload speed (Mbps)
• Server location
• Latency (ms)
• ISP information

<b>🔒 Restrictions:</b>
• <b>Sudo users</b> only

<b>⚠️ Note:</b>
Test may take 30-60 seconds to complete.`
}

func sptHandle(m *telegram.NewMessage) error {
	chatID := m.ChannelID()

	user, err := speedtest.FetchUserInfo()
	if err != nil {
		m.Reply(F(chatID, "spt_fetch_fail", locales.Arg{
			"error": err,
		}))
		return nil
	}

	servers, err := speedtest.FetchServers()
	if err != nil {
		m.Reply(F(chatID, "spt_servers_fetch_fail", locales.Arg{
			"error": err,
		}))
		return nil
	}

	best, err := servers.FindServer([]int{})
	if err != nil || len(best) == 0 {
		m.Reply(F(chatID, "spt_best_server_fail", locales.Arg{
			"error": err,
		}))
		return nil
	}
	server := best[0]

	statusMsg, err := m.Reply(F(chatID, "spt_running_download"))
	if err != nil {
		return err
	}

	server.DownloadTest()

	utils.EOR(statusMsg, F(chatID, "spt_running_upload"))
	server.UploadTest()

	output := F(chatID, "spt_result", locales.Arg{
		"ip":          user.IP,
		"isp":         user.Isp,
		"lat":         user.Lat,
		"lon":         user.Lon,
		"server_name": server.Name,
		"country":     server.Country,
		"sponsor":     server.Sponsor,
		"distance_km": fmt.Sprintf("%.2f", server.Distance),
		"latency_ms": fmt.Sprintf(
			"%.2f",
			float64(server.Latency.Microseconds())/1000,
		),
		"dl_mbps": fmt.Sprintf("%.2f", server.DLSpeed/1024/1024),
		"ul_mbps": fmt.Sprintf("%.2f", server.ULSpeed/1024/1024),
	})

	utils.EOR(statusMsg, output)
	return nil
}
