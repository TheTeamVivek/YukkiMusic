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

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/showwin/speedtest-go/speedtest"

	"main/internal/utils"
)

func sptHandle(m *telegram.NewMessage) error {
	user, err := speedtest.FetchUserInfo()
	if err != nil {
		m.Reply(fmt.Sprintf("⚠️ Failed to fetch network info: %v", err))
		return nil
	}

	servers, err := speedtest.FetchServers()
	if err != nil {
		m.Reply(fmt.Sprintf("⚠️ Failed to fetch servers: %v", err))
		return nil
	}

	best, err := servers.FindServer([]int{})
	if err != nil || len(best) == 0 {
		m.Reply(fmt.Sprintf("⚠️ Failed to find best server: %v", err))
		return nil
	}

	server := best[0]

	statusMsg, err := m.Reply("🚀 Running download test...")
	if err != nil {
		return err
	}

	server.DownloadTest()

	utils.EOR(statusMsg, "📤 Running upload test...")
	server.UploadTest()

	res := `<b>📡 SpeedTest Result</b>

<b><u>👤 Client</u></b>
IP: <code>%s</code>
ISP: %s
Lat: <code>%s</code>
Lon: <code>%s</code>

<b><u>🛰️ Server</u></b>
Name: %s
Country: %s
Sponsor: %s
Distance: %.2f km
Latency: %.2f ms

<b><u>⚡ Speed</u></b>
Download: <code>%.2f Mbps</code>
Upload: <code>%.2f Mbps</code>
`

	output := fmt.Sprintf(
		res,
		user.IP,
		user.Isp,
		user.Lat,
		user.Lon,
		server.Name,
		server.Country,
		server.Sponsor,
		server.Distance,
		float64(server.Latency.Microseconds())/1000,
		server.DLSpeed/1024/1024,
		server.ULSpeed/1024/1024,
	)

	utils.EOR(statusMsg, output)
	return nil
}
