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
	td "github.com/AshokShau/gotdbot"
	"github.com/AshokShau/gotdbot/filters/callbackquery"
	"github.com/amarnathcjd/gogram/telegram"

	"yukkimusic/config"
	"yukkimusic/internal/core"
)

func Init(tdbot *td.Client, bot *telegram.Client, assistants *core.AssistantManager) {
	bot.UpdatesGetState()
	bot.Use(blacklistMessageMiddleware)
	assistants.ForEach(func(a *core.Assistant) {
		a.Client.UpdatesGetState()
	})

	// Basic commands
	bot.AddCommandHandler("(ac|active|activevc|activevoice)", activeHandler, sudoOnlyFilter, ignoreChannelFilter).SetGroup(100)
	bot.AddCommandHandler("(addsudo|addsudoer|sudoadd)", handleAddSudo, ownerFilter, ignoreChannelFilter).SetGroup(100)
	bot.AddCommandHandler("(bash|sh)", shellHandle, ownerFilter).SetGroup(100)
	bot.AddCommandHandler("(blocked|blacklisted)", handleBlacklisted, ownerFilter, ignoreChannelFilter).SetGroup(100)
	bot.AddCommandHandler("(blockchat|blacklistchat|blackchat|blchat)", handleBlockChat, ownerFilter, ignoreChannelFilter).SetGroup(100)
	bot.AddCommandHandler("(blockuser|blacklistuser|blackuser|bluser)", handleBlockUser, ownerFilter, ignoreChannelFilter).SetGroup(100)
	bot.AddCommandHandler("(broadcast|gcast|bcast)", broadcastHandler, ownerFilter, ignoreChannelFilter).SetGroup(100)
	bot.AddCommandHandler("(delsudo|delsudoer|sudodel|remsudo|rmsudo|sudorem|dropsudo|unsudo)", handleDelSudo, ownerFilter, ignoreChannelFilter).SetGroup(100)
	bot.AddCommandHandler("(lang|language)", langHandler, superGroupFilter, adminFilter).SetGroup(100)
	bot.AddCommandHandler("(log|logs)", logsHandler, sudoOnlyFilter, ignoreChannelFilter).SetGroup(100)
	bot.AddCommandHandler("(maintenance|maint)", handleMaintenance, ownerFilter, ignoreChannelFilter).SetGroup(100)
	bot.AddCommandHandler("(sudoers|listsudo|sudolist)", handleGetSudoers, ignoreChannelFilter).SetGroup(100)
	bot.AddCommandHandler("(unblockchat|unblacklistchat|unblackchat|whitechat|unblchat)", handleUnblockChat, ownerFilter, ignoreChannelFilter).SetGroup(100)
	bot.AddCommandHandler("(unblockuser|unblacklistuser|unbluser|whitelistuser)", handleUnblockUser, ownerFilter, ignoreChannelFilter).SetGroup(100)
	bot.AddCommandHandler("autoleave", autoLeaveHandler, sudoOnlyFilter, ignoreChannelFilter).SetGroup(100)
	bot.AddCommandHandler("adminmode", adminModeHandler, superGroupFilter, adminFilter).SetGroup(100)
	bot.AddCommandHandler("cleanmode", cleanModeHandler, superGroupFilter, adminFilter).SetGroup(100)
	bot.AddCommandHandler("eval", evalHandle, ownerFilter).SetGroup(100)
	bot.AddCommandHandler("ev", evalCommandHandler, ownerFilter).SetGroup(100)
	bot.AddCommandHandler("json", jsonHandle).SetGroup(100)
	bot.AddCommandHandler("logger", handleLogger, sudoOnlyFilter, ignoreChannelFilter).SetGroup(100)
	bot.AddCommandHandler("play", playHandler, superGroupFilter).SetGroup(100)
	bot.AddCommandHandler("playmode", playmodeHandler, superGroupFilter, adminFilter).SetGroup(100)
	bot.AddCommandHandler("reload", reloadHandler, superGroupFilter).SetGroup(100)
	bot.AddCommandHandler("restart", handleRestart, ownerFilter, ignoreChannelFilter).SetGroup(100)
	bot.AddCommandHandler("settings", settingsHandler, superGroupFilter, adminFilter).SetGroup(100)
	bot.AddCommandHandler("skip", skipHandler, superGroupFilter, authFilter).SetGroup(100)
	bot.AddCommandHandler("stats", statsHandler, ignoreChannelFilter, sudoOnlyFilter).SetGroup(100)
	bot.AddCommandHandler("vplay", vplayHandler, superGroupFilter).SetGroup(100)

	// CPlay commands
	bot.AddCommandHandler("(auth|addauth)", addAuthHandler, superGroupFilter, adminFilter).SetGroup(100)
	bot.AddCommandHandler("(cfplay|fcplay|cplayforce)", cfplayHandler, superGroupFilter, authFilter).SetGroup(100)
	bot.AddCommandHandler("(channelplay|setcplay)", setCPlayHandler, superGroupFilter, authFilter).SetGroup(100)
	bot.AddCommandHandler("(cstop|cend)", cstopHandler, superGroupFilter, authFilter).SetGroup(100)
	bot.AddCommandHandler("(cmddelete|commanddelete)", cmdDeleteHandler, superGroupFilter, adminFilter).SetGroup(100)
	bot.AddCommandHandler("creload", creloadHandler, superGroupFilter, authFilter).SetGroup(100)
	bot.AddCommandHandler("crestore", crestoreHandler, superGroupFilter, authFilter).SetGroup(100)
	bot.AddCommandHandler("cskip", cskipHandler, superGroupFilter, authFilter).SetGroup(100)
	bot.AddCommandHandler("(end|stop)", stopHandler, superGroupFilter, authFilter).SetGroup(100)
	bot.AddCommandHandler("(fplay|playforce)", fplayHandler, superGroupFilter, authFilter).SetGroup(100)
	bot.AddCommandHandler("(fvplay|vfplay|vplayforce)", fvplayHandler, superGroupFilter, authFilter).SetGroup(100)
	bot.AddCommandHandler("(fvcplay|fvcpay|vcplayforce)", fvcplayHandler, superGroupFilter, authFilter).SetGroup(100)
	bot.AddCommandHandler("(nothumb|nothumbs)", nothumbHandler, superGroupFilter, authFilter).SetGroup(100)
	bot.AddCommandHandler("(vcplay|cvplay)", vcplayHandler, superGroupFilter).SetGroup(100)
	bot.AddCommandHandler("cplay", cplayHandler, superGroupFilter).SetGroup(100)
	bot.AddCommandHandler("crestore", crestoreHandler, superGroupFilter, authFilter).SetGroup(100)
	bot.AddCommandHandler("restore", restoreHandler, superGroupFilter, authFilter).SetGroup(100)

	bot.AddCallbackHandler("^bcast_cancel$", broadcastCancelCB).SetGroup(90)
	bot.AddCallbackHandler("^cancel$", cancelHandler).SetGroup(90)
	bot.AddCallbackHandler("^close$", closeHandler).SetGroup(90)
	bot.AddCallbackHandler("^lang:[a-z]$", langCallbackHandler).SetGroup(90)
	bot.AddCallbackHandler("^restart:(bot|replay)$", restartConfirmHandler).SetGroup(90)
	bot.AddCallbackHandler("^room:-?\\d+:\\w+$", roomHandle).SetGroup(90)
	bot.AddCallbackHandler("^set|info:", settingsCallbackHandler).SetGroup(90)
	bot.AddCallbackHandler("progress", emptyCBHandler).SetGroup(90)

	// td client handlers
	tdbot.OnCommand("start", startHandler)
	tdbot.OnCommand("help", helpHandler)
	tdbot.OnCommand("ping", pingHandler)
	tdbot.OnCommand("mute", muteHandler)
	tdbot.OnCommand("cmute", cmuteHandler)
	tdbot.OnCommand("unmute", unmuteHandler)
	tdbot.OnCommand("cunmute", cunmuteHandler)
	tdbot.OnCommand("pause", pauseHandler)
	tdbot.OnCommand("cpause", cpauseHandler)
	tdbot.OnCommand("resume", resumeHandler)
	tdbot.OnCommand("cresume", cresumeHandler)
	tdbot.OnCommand("position", positionHandler)
	tdbot.OnCommand("pos", positionHandler)
	tdbot.OnCommand("cposition", cpositionHandler)
	tdbot.OnCommand("cpos", cpositionHandler)
	tdbot.OnCommand("replay", replayHandler)
	tdbot.OnCommand("creplay", creplayHandler)
	tdbot.OnCommand("loop", loopHandler)
	tdbot.OnCommand("setloop", loopHandler)
	tdbot.OnCommand("cloop", cloopHandler)
	tdbot.OnCommand("csetloop", cloopHandler)
	tdbot.OnCommand("speed", speedHandler)
	tdbot.OnCommand("setspeed", speedHandler)
	tdbot.OnCommand("speedup", speedHandler)
	tdbot.OnCommand("cspeed", cspeedHandler)
	tdbot.OnCommand("csetspeed", cspeedHandler)
	tdbot.OnCommand("cspeedup", cspeedHandler)
	tdbot.OnCommand("shuffle", shuffleHandler)
	tdbot.OnCommand("cshuffle", cshuffleHandler)
	tdbot.OnCommand("seek", seekHandler)
	tdbot.OnCommand("cseek", cseekHandler)
	tdbot.OnCommand("seekback", seekbackHandler)
	tdbot.OnCommand("cseekback", cseekbackHandler)
	tdbot.OnCommand("jump", jumpHandler)
	tdbot.OnCommand("cjump", cjumpHandler)
	tdbot.OnCommand("queue", queueHandler)
	tdbot.OnCommand("cqueue", cqueueHandler)
	tdbot.OnCommand("remove", removeHandler)
	tdbot.OnCommand("cremove", cremoveHandler)
	tdbot.OnCommand("clear", clearHandler)
	tdbot.OnCommand("cclear", cclearHandler)
	tdbot.OnCommand("move", moveHandler)
	tdbot.OnCommand("cmove", cmoveHandler)
	tdbot.OnUpdateNewCallbackQuery(startCB, callbackquery.Equal("start"))
	tdbot.OnUpdateNewCallbackQuery(helpCB, callbackquery.Equal("help_cb"))
	tdbot.OnUpdateNewCallbackQuery(helpCallbackHandler, callbackquery.Regex("^help:(.+)$"))
	tdbot.OnUpdateChatMember(handleParticipantUpdate, nil)
	tdbot.OnMessage(handleActionsTd, actionFilter)

	bot.On("edit:/eval", evalHandle).SetGroup(80)
	bot.On("edit:/ev", evalCommandHandler).SetGroup(80)

	bot.AddRawHandler(&telegram.UpdateReadChannelOutbox{}, cleanModeReadHandler)

	assistants.ForEach(func(a *core.Assistant) {
		a.Ntg.OnStreamEnd(streamEndHandler)
	})

	go MonitorRooms()

	autoLeaveSvc.Start()
	cleanScheduler.start()

	if config.SetCmds && config.OwnerID != 0 {
		go setBotCommands(bot)
	}
}
