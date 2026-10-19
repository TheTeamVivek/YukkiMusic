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

func Init(bot *td.Client, assistants *core.ClientManager) {
	assistants.ForEach(func(a *core.Assistant) {
		a.Client.UpdatesGetState()
	})

	// td client handlers
	bot.OnCommand("start", WithBlacklistMessage(startHandler))
	bot.OnCommand("help", WithBlacklistMessage(helpHandler))
	bot.OnCommand("ping", WithBlacklistMessage(pingHandler))
	bot.OnCommand("mute", WithBlacklistMessage(muteHandler))
	bot.OnCommand("cmute", WithBlacklistMessage(cmuteHandler))
	bot.OnCommand("unmute", WithBlacklistMessage(unmuteHandler))
	bot.OnCommand("cunmute", WithBlacklistMessage(cunmuteHandler))
	bot.OnCommand("pause", WithBlacklistMessage(pauseHandler))
	bot.OnCommand("cpause", WithBlacklistMessage(cpauseHandler))
	bot.OnCommand("resume", WithBlacklistMessage(resumeHandler))
	bot.OnCommand("cresume", WithBlacklistMessage(cresumeHandler))
	bot.OnCommand("position", WithBlacklistMessage(positionHandler)).AddAliases("pos")
	bot.OnCommand("cposition", WithBlacklistMessage(cpositionHandler)).AddAliases("cpos")
	bot.OnCommand("replay", WithBlacklistMessage(replayHandler))
	bot.OnCommand("creplay", WithBlacklistMessage(creplayHandler))
	bot.OnCommand("loop", WithBlacklistMessage(loopHandler)).AddAliases("setloop")
	bot.OnCommand("cloop", WithBlacklistMessage(cloopHandler)).AddAliases("csetloop")
	bot.OnCommand("speed", WithBlacklistMessage(speedHandler)).AddAliases("setspeed", "speedup")
	bot.OnCommand("cspeed", WithBlacklistMessage(cspeedHandler)).AddAliases("csetspeed", "cspeedup")
	bot.OnCommand("shuffle", WithBlacklistMessage(shuffleHandler))
	bot.OnCommand("cshuffle", WithBlacklistMessage(cshuffleHandler))
	bot.OnCommand("seek", WithBlacklistMessage(seekHandler))
	bot.OnCommand("cseek", WithBlacklistMessage(cseekHandler))
	bot.OnCommand("seekback", WithBlacklistMessage(seekbackHandler))
	bot.OnCommand("cseekback", WithBlacklistMessage(cseekbackHandler))
	bot.OnCommand("jump", WithBlacklistMessage(jumpHandler))
	bot.OnCommand("cjump", WithBlacklistMessage(cjumpHandler))
	bot.OnCommand("queue", WithBlacklistMessage(queueHandler))
	bot.OnCommand("cqueue", WithBlacklistMessage(cqueueHandler))
	bot.OnCommand("remove", WithBlacklistMessage(removeHandler))
	bot.OnCommand("cremove", WithBlacklistMessage(cremoveHandler))
	bot.OnCommand("clear", WithBlacklistMessage(clearHandler))
	bot.OnCommand("cclear", WithBlacklistMessage(cclearHandler))
	bot.OnCommand("move", WithBlacklistMessage(moveHandler))
	bot.OnCommand("cmove", WithBlacklistMessage(cmoveHandler))
	bot.OnCommand("ac", WithBlacklistMessage(activeHandler)).AddAliases("active", "activevc", "activevoice")
	bot.OnCommand("addsudo", WithBlacklistMessage(handleAddSudo)).AddAliases("addsudoer", "sudoadd")
	bot.OnCommand("sh", WithBlacklistMessage(shellHandler)).AddAliases("bash", "shell")
	bot.OnCommand("json", WithBlacklistMessage(jsonHandle))
	bot.OnCommand("log", WithBlacklistMessage(logsHandler)).AddAliases("logs")
	bot.OnCommand("skip", WithBlacklistMessage(skipHandler))
	bot.OnCommand("cskip", WithBlacklistMessage(cskipHandler))
	bot.OnCommand("play", WithBlacklistMessage(playHandler))
	bot.OnCommand("fplay", WithBlacklistMessage(fplayHandler)).AddAliases("playforce")
	bot.OnCommand("cfplay", WithBlacklistMessage(cfplayHandler)).AddAliases("fcplay", "cforceplay", "cplayforce")
	bot.OnCommand("vplay", WithBlacklistMessage(vplayHandler))
	bot.OnCommand("fvplay", WithBlacklistMessage(fvplayHandler)).AddAliases("vfplay", "vplayforce")
	bot.OnCommand("vcplay", WithBlacklistMessage(vcplayHandler)).AddAliases("cvplay")
	bot.OnCommand("fvcplay", WithBlacklistMessage(fvcplayHandler)).AddAliases("vcplayforce")
	bot.OnCommand("cplay", WithBlacklistMessage(cplayHandler))
	bot.OnCommand("end", WithBlacklistMessage(stopHandler)).AddAliases("stop")
	bot.OnCommand("cstop", WithBlacklistMessage(cstopHandler)).AddAliases("cend")
	bot.OnCommand("nothumb", WithBlacklistMessage(nothumbHandler)).AddAliases("nothumbs")
	bot.OnCommand("channelplay", WithBlacklistMessage(setCPlayHandler)).AddAliases("setcplay")
	bot.OnCommand("blocked", WithBlacklistMessage(handleBlacklisted)).AddAliases("blacklisted")
	bot.OnCommand("blockchat", WithBlacklistMessage(handleBlockChat)).AddAliases("blacklistchat", "blackchat", "blchat")
	bot.OnCommand("blockuser", WithBlacklistMessage(handleBlockUser)).AddAliases("blacklistuser", "blackuser", "bluser")
	bot.OnCommand("broadcast", WithBlacklistMessage(broadcastHandler)).AddAliases("gcast", "bcast")
	bot.OnCommand("delsudo", WithBlacklistMessage(handleDelSudo)).AddAliases("delsudoer", "sudodel", "remsudo", "rmsudo", "sudorem", "dropsudo", "unsudo")
	bot.OnCommand("lang", WithBlacklistMessage(langHandler)).AddAliases("language")
	bot.OnCommand("maintenance", WithBlacklistMessage(handleMaintenance)).AddAliases("maint")
	bot.OnCommand("sudoers", WithBlacklistMessage(handleGetSudoers)).AddAliases("listsudo", "sudolist")
	bot.OnCommand("unblockchat", WithBlacklistMessage(handleUnblockChat)).AddAliases("unblacklistchat", "unblackchat", "whitechat", "unblchat")
	bot.OnCommand("unblockuser", WithBlacklistMessage(handleUnblockUser)).AddAliases("unblacklistuser", "unbluser", "whitelistuser")
	bot.OnCommand("autoleave", WithBlacklistMessage(autoLeaveHandler))
	bot.OnCommand("adminmode", WithBlacklistMessage(adminModeHandler))
	bot.OnCommand("cleanmode", WithBlacklistMessage(cleanModeHandler))
	bot.OnCommand("logger", WithBlacklistMessage(handleLogger))
	bot.OnCommand("playmode", WithBlacklistMessage(playmodeHandler))
	bot.OnCommand("autoplay", WithBlacklistMessage(autoplayHandler))
	bot.OnCommand("reload", WithBlacklistMessage(reloadHandler))
	bot.OnCommand("restart", WithBlacklistMessage(handleRestart))
	bot.OnCommand("settings", WithBlacklistMessage(settingsHandler))
	bot.OnCommand("stats", WithBlacklistMessage(statsHandler))
	bot.OnCommand("auth", WithBlacklistMessage(addAuthHandler)).AddAliases("addauth", "authorize")
	bot.OnCommand("delauth", WithBlacklistMessage(delAuthHandler)).AddAliases("unauth", "delauthuser")
	bot.OnCommand("authlist", WithBlacklistMessage(authListHandler)).AddAliases("authusers")
	bot.OnCommand("privacy", WithBlacklistMessage(privacyHandler))
	bot.OnCommand("cmddelete", WithBlacklistMessage(cmdDeleteHandler)).AddAliases("commanddelete")
	bot.OnCommand("creload", WithBlacklistMessage(creloadHandler))
	bot.OnUpdateNewCallbackQuery(WithBlacklistCallback(broadcastCancelCB), callbackquery.Regex("^bcast_cancel$"))
	bot.OnUpdateNewCallbackQuery(WithBlacklistCallback(langCallbackHandler), callbackquery.Regex("^lang:(select|[a-z]{2})$"))
	bot.OnUpdateNewCallbackQuery(WithBlacklistCallback(restartConfirmHandler), callbackquery.Regex("^restart:(bot|replay)$"))
	bot.OnUpdateNewCallbackQuery(WithBlacklistCallback(settingsCallbackHandler), callbackquery.Regex("^set|info:"))
	bot.OnUpdateNewCallbackQuery(WithBlacklistCallback(startCB), callbackquery.Equal("start"))
	bot.OnUpdateNewCallbackQuery(WithBlacklistCallback(helpCB), callbackquery.Equal("help_cb"))
	bot.OnUpdateNewCallbackQuery(WithBlacklistCallback(helpCallbackHandler), callbackquery.Regex("^help:(.+)$"))
	bot.OnUpdateChatMember(handleParticipantUpdate, nil)
	bot.OnMessage(WithBlacklistMessage(handleActions), actionFilter)
	bot.OnUpdateFile(downloadUpdateHandler, nil)

	bot.OnUpdateChatReadOutbox(cleanModeReadHandler, nil)
	bot.OnUpdateNewCallbackQuery(WithBlacklistCallback(cancelHandler), callbackquery.Equal("cancel"))
	bot.OnUpdateNewCallbackQuery(WithBlacklistCallback(closeHandler), callbackquery.Equal("close"))
	bot.OnUpdateNewCallbackQuery(WithBlacklistCallback(roomHandle), callbackquery.Regex("^room:-?\\d+:\\w+$"))
	bot.OnUpdateNewCallbackQuery(WithBlacklistCallback(queueDeleteHandler), callbackquery.Regex("^qdel:-?\\d+:\\d+:\\d+$"))
	bot.OnUpdateNewCallbackQuery(WithBlacklistCallback(queueMoreHandler), callbackquery.Regex("^qmore:-?\\d+$"))
	bot.OnUpdateNewCallbackQuery(WithBlacklistCallback(authDeleteHandler), callbackquery.Regex("^authdel:-?\\d+:-?\\d+$"))
	bot.OnUpdateNewCallbackQuery(WithBlacklistCallback(dashboardCallback), callbackquery.Regex("^dashboard:[a-z]+$"))
	bot.OnUpdateNewCallbackQuery(WithBlacklistCallback(emptyCBHandler), callbackquery.Regex("progress"))

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
