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

	"yukkimusic/config"
	"yukkimusic/internal/core"
)

func Init(tdbot *td.Client, assistants *core.AssistantManager) {
	assistants.ForEach(func(a *core.Assistant) {
		a.Client.UpdatesGetState()
	})

	// td client handlers
	tdbot.OnCommand("start", WithBlacklistMessage(startHandler))
	tdbot.OnCommand("help", WithBlacklistMessage(helpHandler))
	tdbot.OnCommand("ping", WithBlacklistMessage(pingHandler))
	tdbot.OnCommand("mute", WithBlacklistMessage(muteHandler))
	tdbot.OnCommand("cmute", WithBlacklistMessage(cmuteHandler))
	tdbot.OnCommand("unmute", WithBlacklistMessage(unmuteHandler))
	tdbot.OnCommand("cunmute", WithBlacklistMessage(cunmuteHandler))
	tdbot.OnCommand("pause", WithBlacklistMessage(pauseHandler))
	tdbot.OnCommand("cpause", WithBlacklistMessage(cpauseHandler))
	tdbot.OnCommand("resume", WithBlacklistMessage(resumeHandler))
	tdbot.OnCommand("cresume", WithBlacklistMessage(cresumeHandler))
	tdbot.OnCommand("position", WithBlacklistMessage(positionHandler))
	tdbot.OnCommand("pos", WithBlacklistMessage(positionHandler))
	tdbot.OnCommand("cposition", WithBlacklistMessage(cpositionHandler))
	tdbot.OnCommand("cpos", WithBlacklistMessage(cpositionHandler))
	tdbot.OnCommand("replay", WithBlacklistMessage(replayHandler))
	tdbot.OnCommand("creplay", WithBlacklistMessage(creplayHandler))
	tdbot.OnCommand("loop", WithBlacklistMessage(loopHandler))
	tdbot.OnCommand("setloop", WithBlacklistMessage(loopHandler))
	tdbot.OnCommand("cloop", WithBlacklistMessage(cloopHandler))
	tdbot.OnCommand("csetloop", WithBlacklistMessage(cloopHandler))
	tdbot.OnCommand("speed", WithBlacklistMessage(speedHandler))
	tdbot.OnCommand("setspeed", WithBlacklistMessage(speedHandler))
	tdbot.OnCommand("speedup", WithBlacklistMessage(speedHandler))
	tdbot.OnCommand("cspeed", WithBlacklistMessage(cspeedHandler))
	tdbot.OnCommand("csetspeed", WithBlacklistMessage(cspeedHandler))
	tdbot.OnCommand("cspeedup", WithBlacklistMessage(cspeedHandler))
	tdbot.OnCommand("shuffle", WithBlacklistMessage(shuffleHandler))
	tdbot.OnCommand("cshuffle", WithBlacklistMessage(cshuffleHandler))
	tdbot.OnCommand("seek", WithBlacklistMessage(seekHandler))
	tdbot.OnCommand("cseek", WithBlacklistMessage(cseekHandler))
	tdbot.OnCommand("seekback", WithBlacklistMessage(seekbackHandler))
	tdbot.OnCommand("cseekback", WithBlacklistMessage(cseekbackHandler))
	tdbot.OnCommand("jump", WithBlacklistMessage(jumpHandler))
	tdbot.OnCommand("cjump", WithBlacklistMessage(cjumpHandler))
	tdbot.OnCommand("queue", WithBlacklistMessage(queueHandler))
	tdbot.OnCommand("cqueue", WithBlacklistMessage(cqueueHandler))
	tdbot.OnCommand("remove", WithBlacklistMessage(removeHandler))
	tdbot.OnCommand("cremove", WithBlacklistMessage(cremoveHandler))
	tdbot.OnCommand("clear", WithBlacklistMessage(clearHandler))
	tdbot.OnCommand("cclear", WithBlacklistMessage(cclearHandler))
	tdbot.OnCommand("move", WithBlacklistMessage(moveHandler))
	tdbot.OnCommand("cmove", WithBlacklistMessage(cmoveHandler))
	tdbot.OnCommand("ac", WithBlacklistMessage(activeHandler))
	tdbot.OnCommand("active", WithBlacklistMessage(activeHandler))
	tdbot.OnCommand("activevc", WithBlacklistMessage(activeHandler))
	tdbot.OnCommand("activevoice", WithBlacklistMessage(activeHandler))
	tdbot.OnCommand("addsudo", WithBlacklistMessage(handleAddSudo))
	tdbot.OnCommand("ev", WithBlacklistMessage(evalCommandHandler))
	tdbot.OnCommand("json", WithBlacklistMessage(jsonHandle))
	tdbot.OnCommand("(log|logs)", WithBlacklistMessage(logsHandler))
	tdbot.OnCommand("skip", WithBlacklistMessage(skipHandler))
	tdbot.OnCommand("cskip", WithBlacklistMessage(cskipHandler))
	tdbot.OnCommand("(end|stop)", WithBlacklistMessage(stopHandler))
	tdbot.OnCommand("(cstop|cend)", WithBlacklistMessage(cstopHandler))
	tdbot.OnCommand("(nothumb|nothumbs)", WithBlacklistMessage(nothumbHandler))
	tdbot.OnCommand("(channelplay|setcplay)", WithBlacklistMessage(setCPlayHandler))
	tdbot.OnCommand("addsudoer", WithBlacklistMessage(handleAddSudo))
	tdbot.OnCommand("sudoadd", WithBlacklistMessage(handleAddSudo))
	tdbot.OnCommand("blocked", WithBlacklistMessage(handleBlacklisted))
	tdbot.OnCommand("blacklisted", WithBlacklistMessage(handleBlacklisted))
	tdbot.OnCommand("blockchat", WithBlacklistMessage(handleBlockChat))
	tdbot.OnCommand("blacklistchat", WithBlacklistMessage(handleBlockChat))
	tdbot.OnCommand("blackchat", WithBlacklistMessage(handleBlockChat))
	tdbot.OnCommand("blchat", WithBlacklistMessage(handleBlockChat))
	tdbot.OnCommand("blockuser", WithBlacklistMessage(handleBlockUser))
	tdbot.OnCommand("blacklistuser", WithBlacklistMessage(handleBlockUser))
	tdbot.OnCommand("blackuser", WithBlacklistMessage(handleBlockUser))
	tdbot.OnCommand("bluser", WithBlacklistMessage(handleBlockUser))
	tdbot.OnCommand("broadcast", WithBlacklistMessage(broadcastHandler))
	tdbot.OnCommand("gcast", WithBlacklistMessage(broadcastHandler))
	tdbot.OnCommand("bcast", WithBlacklistMessage(broadcastHandler))
	tdbot.OnCommand("delsudo", WithBlacklistMessage(handleDelSudo))
	tdbot.OnCommand("delsudoer", WithBlacklistMessage(handleDelSudo))
	tdbot.OnCommand("sudodel", WithBlacklistMessage(handleDelSudo))
	tdbot.OnCommand("remsudo", WithBlacklistMessage(handleDelSudo))
	tdbot.OnCommand("rmsudo", WithBlacklistMessage(handleDelSudo))
	tdbot.OnCommand("sudorem", WithBlacklistMessage(handleDelSudo))
	tdbot.OnCommand("dropsudo", WithBlacklistMessage(handleDelSudo))
	tdbot.OnCommand("unsudo", WithBlacklistMessage(handleDelSudo))
	tdbot.OnCommand("lang", WithBlacklistMessage(langHandler))
	tdbot.OnCommand("language", WithBlacklistMessage(langHandler))
	tdbot.OnCommand("maintenance", WithBlacklistMessage(handleMaintenance))
	tdbot.OnCommand("maint", WithBlacklistMessage(handleMaintenance))
	tdbot.OnCommand("sudoers", WithBlacklistMessage(handleGetSudoers))
	tdbot.OnCommand("listsudo", WithBlacklistMessage(handleGetSudoers))
	tdbot.OnCommand("sudolist", WithBlacklistMessage(handleGetSudoers))
	tdbot.OnCommand("unblockchat", WithBlacklistMessage(handleUnblockChat))
	tdbot.OnCommand("unblacklistchat", WithBlacklistMessage(handleUnblockChat))
	tdbot.OnCommand("unblackchat", WithBlacklistMessage(handleUnblockChat))
	tdbot.OnCommand("whitechat", WithBlacklistMessage(handleUnblockChat))
	tdbot.OnCommand("unblchat", WithBlacklistMessage(handleUnblockChat))
	tdbot.OnCommand("unblockuser", WithBlacklistMessage(handleUnblockUser))
	tdbot.OnCommand("unblacklistuser", WithBlacklistMessage(handleUnblockUser))
	tdbot.OnCommand("unbluser", WithBlacklistMessage(handleUnblockUser))
	tdbot.OnCommand("whitelistuser", WithBlacklistMessage(handleUnblockUser))
	tdbot.OnCommand("autoleave", WithBlacklistMessage(autoLeaveHandler))
	tdbot.OnCommand("adminmode", WithBlacklistMessage(adminModeHandler))
	tdbot.OnCommand("cleanmode", WithBlacklistMessage(cleanModeHandler))
	tdbot.OnCommand("logger", WithBlacklistMessage(handleLogger))
	tdbot.OnCommand("playmode", WithBlacklistMessage(playmodeHandler))
	tdbot.OnCommand("reload", WithBlacklistMessage(reloadHandler))
	tdbot.OnCommand("restart", WithBlacklistMessage(handleRestart))
	tdbot.OnCommand("settings", WithBlacklistMessage(settingsHandler))
	tdbot.OnCommand("stats", WithBlacklistMessage(statsHandler))
	tdbot.OnCommand("auth", WithBlacklistMessage(addAuthHandler))
	tdbot.OnCommand("addauth", WithBlacklistMessage(addAuthHandler))
	tdbot.OnCommand("cmddelete", WithBlacklistMessage(cmdDeleteHandler))
	tdbot.OnCommand("commanddelete", WithBlacklistMessage(cmdDeleteHandler))
	tdbot.OnCommand("creload", WithBlacklistMessage(creloadHandler))
	tdbot.OnUpdateNewCallbackQuery(WithBlacklistCallback(broadcastCancelCB), callbackquery.Regex("^bcast_cancel$"))
	tdbot.OnUpdateNewCallbackQuery(WithBlacklistCallback(langCallbackHandler), callbackquery.Regex("^lang:[a-z]$"))
	tdbot.OnUpdateNewCallbackQuery(WithBlacklistCallback(restartConfirmHandler), callbackquery.Regex("^restart:(bot|replay)$"))
	tdbot.OnUpdateNewCallbackQuery(WithBlacklistCallback(settingsCallbackHandler), callbackquery.Regex("^set|info:"))
	tdbot.OnUpdateNewCallbackQuery(WithBlacklistCallback(startCB), callbackquery.Equal("start"))
	tdbot.OnUpdateNewCallbackQuery(WithBlacklistCallback(helpCB), callbackquery.Equal("help_cb"))
	tdbot.OnUpdateNewCallbackQuery(WithBlacklistCallback(helpCallbackHandler), callbackquery.Regex("^help:(.+)$"))
	tdbot.OnUpdateChatMember(handleParticipantUpdate, nil)
	tdbot.OnMessage(WithBlacklistMessage(handleActions), actionFilter)
	tdbot.OnUpdateFile(downloadUpdateHandler, nil)

	tdbot.OnUpdateChatReadOutbox(cleanModeReadHandler, nil)
	tdbot.OnUpdateNewCallbackQuery(WithBlacklistCallback(cancelHandler), callbackquery.Equal("cancel"))
	tdbot.OnUpdateNewCallbackQuery(WithBlacklistCallback(closeHandler), callbackquery.Equal("close"))
	tdbot.OnUpdateNewCallbackQuery(WithBlacklistCallback(roomHandle), callbackquery.Regex("^room:-?\\d+:\\w+$"))
	tdbot.OnUpdateNewCallbackQuery(WithBlacklistCallback(emptyCBHandler), callbackquery.Regex("progress"))

	assistants.ForEach(func(a *core.Assistant) {
		a.Ntg.OnStreamEnd(streamEndHandler)
	})

	go MonitorRooms()

	autoLeaveSvc.Start()
	cleanScheduler.start()

	if config.SetCmds && config.OwnerID != 0 {
		go setBotCommands(tdbot)
	}
}
