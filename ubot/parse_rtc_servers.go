package ubot

import (
	tg "github.com/amarnathcjd/gogram/telegram"

	"yukkimusic/ntgcalls"
)

func parseRTCServers(connections []tg.PhoneConnection) []ntgcalls.RTCServer {
	rtcServers := make([]ntgcalls.RTCServer, len(connections))
	for i, c := range connections {
		switch connection := c.(type) {
		case *tg.PhoneConnectionWebrtc:
			rtcServers[i] = ntgcalls.RTCServer{
				ID:       connection.ID,
				IPv4:     connection.Ip,
				IPv6:     connection.Ipv6,
				Username: connection.Username,
				Password: connection.Password,
				Port:     connection.Port,
				TURN:     connection.Turn,
				STUN:     connection.Stun,
			}
		case *tg.PhoneConnectionObj:
			rtcServers[i] = ntgcalls.RTCServer{
				ID:      connection.ID,
				IPv4:    connection.Ip,
				IPv6:    connection.Ipv6,
				Port:    connection.Port,
				TURN:    true,
				TCP:     connection.Tcp,
				PeerTag: connection.PeerTag,
			}
		}
	}
	return rtcServers
}
