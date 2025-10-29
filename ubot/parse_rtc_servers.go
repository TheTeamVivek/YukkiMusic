package ubot

import (
	tg "github.com/amarnathcjd/gogram/telegram"

	"github.com/TheTeamVivek/YukkiMusic/ntgcalls"
)

func parseRTCServers(connections []tg.PhoneConnection) []ntgcalls.RTCServer {
	rtcServers := make([]ntgcalls.RTCServer, len(connections))
	for i, c := range connections {
		switch connection := c.(type) {
		case *tg.PhoneConnectionWebrtc:
			rtcServers[i] = ntgcalls.RTCServer{
				ID:       connection.ID,
				Ipv4:     connection.Ip,
				Ipv6:     connection.Ipv6,
				Username: connection.Username,
				Password: connection.Password,
				Port:     connection.Port,
				Turn:     connection.Turn,
				Stun:     connection.Stun,
			}
		case *tg.PhoneConnectionObj:
			rtcServers[i] = ntgcalls.RTCServer{
				ID:      connection.ID,
				Ipv4:    connection.Ip,
				Ipv6:    connection.Ipv6,
				Port:    connection.Port,
				Turn:    true,
				Tcp:     connection.Tcp,
				PeerTag: connection.PeerTag,
			}
		}
	}
	return rtcServers
}
