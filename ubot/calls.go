package ubot

import "github.com/TheTeamVivek/YukkiMusic/ntgcalls"

func (ctx *Context) Calls() map[int64]*ntgcalls.CallInfo {
	return ctx.binding.Calls()
}
