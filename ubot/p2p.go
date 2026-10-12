package ubot

import (
	tg "github.com/amarnathcjd/gogram/telegram"
)

func (ctx *Context) newP2PConfig(GAorB []byte) (*P2PConfig, error) {
	dhConfigRaw, err := ctx.app.MessagesGetDhConfig(0, 256)
	if err != nil {
		return nil, err
	}
	dhConfig := dhConfigRaw.(*tg.MessagesDhConfigObj)
	return &P2PConfig{
		DHConfig:   dhConfig,
		IsOutgoing: GAorB == nil,
		GAorB:      GAorB,
		WaitData:   make(chan error),
	}, nil
}
