package ubot

import "main/ntgcalls"

func (ctx *Context) Calls() map[int64]*ntgcalls.CallInfo {
	return ctx.binding.Calls()
}
