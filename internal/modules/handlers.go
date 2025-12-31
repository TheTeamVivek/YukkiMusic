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
	"log"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
	"main/internal/database"
	"main/ntgcalls"
)

type MsgHandlerDef struct {
	Pattern string
	Handler telegram.MessageHandler
	Filters []telegram.Filter
}

type CbHandlerDef struct {
	Pattern string
	Handler telegram.CallbackHandler
	Filters []telegram.Filter
}

var handlers = []MsgHandlerDef{
	{Pattern: "json", Handler: jsonHandle},
	{Pattern: "eval", Handler: evalHandle, Filters: []telegram.Filter{ownerFilter}},
	{Pattern: "ev", Handler: evalCommandHandler, Filters: []telegram.Filter{ownerFilter}},
	{Pattern: "(bash|sh)", Handler: shellHandle, Filters: []telegram.Filter{ownerFilter}},
	{Pattern: "restart", Handler: handleRestart, Filters: []telegram.Filter{ownerFilter, ignoreChannelFilter}},

	{Pattern: "(addsudo|addsudoer|sudoadd)", Handler: handleAddSudo, Filters: []telegram.Filter{ownerFilter, ignoreChannelFilter}},
	{Pattern: "(delsudo|delsudoer|sudodel|remsudo|rmsudo|sudorem|dropsudo|unsudo)", Handler: handleDelSudo, Filters: []telegram.Filter{ownerFilter, ignoreChannelFilter}},
	{Pattern: "(sudoers|listsudo|sudolist)", Handler: handleGetSudoers, Filters: []telegram.Filter{ignoreChannelFilter}},

	{Pattern: "(speedtest|spt)", Handler: sptHandle, Filters: []telegram.Filter{sudoOnlyFilter, ignoreChannelFilter}},

	{Pattern: "(broadcast|gcast|bcast)", Handler: broadcastHandler, Filters: []telegram.Filter{ownerFilter, ignoreChannelFilter}},

	{Pattern: "(ac|active|activevc|activevoice)", Handler: activeHandler, Filters: []telegram.Filter{sudoOnlyFilter, ignoreChannelFilter}},
	{Pattern: "(maintenance|maint)", Handler: handleMaintenance, Filters: []telegram.Filter{ownerFilter, ignoreChannelFilter}},
	{Pattern: "logger", Handler: handleLogger, Filters: []telegram.Filter{sudoOnlyFilter, ignoreChannelFilter}},
	{Pattern: "autoleave", Handler: autoLeaveHandler, Filters: []telegram.Filter{sudoOnlyFilter, ignoreChannelFilter}},
	{Pattern: "(log|logs)", Handler: logsHandler, Filters: []telegram.Filter{sudoOnlyFilter, ignoreChannelFilter}},
	
	{Pattern: "help", Handler: helpHandler, Filters: []telegram.Filter{ignoreChannelFilter}},
	{Pattern: "ping", Handler: pingHandler, Filters: []telegram.Filter{ignoreChannelFilter}},
	{Pattern: "start", Handler: startHandler, Filters: []telegram.Filter{ignoreChannelFilter}},
	{Pattern: "stats", Handler: statsHandler, Filters: []telegram.Filter{ignoreChannelFilter, sudoOnlyFilter}},
	{Pattern: "bug", Handler: bugHandler, Filters: []telegram.Filter{ignoreChannelFilter}},
	{Pattern: "(lang|language)", Handler: langHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},

	// SuperGroup & Admin Filters

	{Pattern: "stream", Handler: streamHandler, Filters: []telegram.Filter{superGroupFilter}},
	{Pattern: "streamstop", Handler: streamStopHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "streamstatus", Handler: streamStatusHandler, Filters: []telegram.Filter{superGroupFilter}},
	{Pattern: "(rtmp|setrtmp)", Handler: setRTMPHandler},

	// play/cplay/vplay/fplay commands
	{Pattern: "play", Handler: playHandler, Filters: []telegram.Filter{superGroupFilter}},
	{Pattern: "(fplay|playforce)", Handler: fplayHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "cplay", Handler: cplayHandler, Filters: []telegram.Filter{superGroupFilter}},
	{Pattern: "(cfplay|fcplay|cplayforce)", Handler: cfplayHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "vplay", Handler: vplayHandler, Filters: []telegram.Filter{superGroupFilter}},
	{Pattern: "(fvplay|vfplay|vplayforce)", Handler: fvplayHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "(vcplay|cvplay)", Handler: vcplayHandler, Filters: []telegram.Filter{superGroupFilter}},
	{Pattern: "(fvcplay|fvcpay|vcplayforce)", Handler: fvcplayHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},

	{Pattern: "(speed|setspeed|speedup)", Handler: speedHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "skip", Handler: skipHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "pause", Handler: pauseHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "resume", Handler: resumeHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "replay", Handler: replayHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "mute", Handler: muteHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "unmute", Handler: unmuteHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "seek", Handler: seekHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "seekback", Handler: seekbackHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "jump", Handler: jumpHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "position", Handler: positionHandler, Filters: []telegram.Filter{superGroupFilter}},
	{Pattern: "queue", Handler: queueHandler, Filters: []telegram.Filter{superGroupFilter}},
	{Pattern: "clear", Handler: clearHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "remove", Handler: removeHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "move", Handler: moveHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "shuffle", Handler: shuffleHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "(loop|setloop)", Handler: loopHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "(end|stop)", Handler: stopHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "reload", Handler: reloadHandler, Filters: []telegram.Filter{superGroupFilter}},
	{Pattern: "addauth", Handler: addAuthHandler, Filters: []telegram.Filter{superGroupFilter, adminFilter}},
	{Pattern: "delauth", Handler: delAuthHandler, Filters: []telegram.Filter{superGroupFilter, adminFilter}},
	{Pattern: "authlist", Handler: authListHandler, Filters: []telegram.Filter{superGroupFilter}},

	// CPlay commands
	{Pattern: "(cplay|cvplay)", Handler: cplayHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "(cfplay|fcplay|cforceplay)", Handler: cfplayHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "cpause", Handler: cpauseHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "cresume", Handler: cresumeHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "cmute", Handler: cmuteHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "cunmute", Handler: cunmuteHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "(cstop|cend)", Handler: cstopHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "cqueue", Handler: cqueueHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "cskip", Handler: cskipHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "(cloop|csetloop)", Handler: cloopHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "cseek", Handler: cseekHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "cseekback", Handler: cseekbackHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "cjump", Handler: cjumpHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "cremove", Handler: cremoveHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "cclear", Handler: cclearHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "cmove", Handler: cmoveHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "channelplay", Handler: channelPlayHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "(cspeed|csetspeed|cspeedup)", Handler: cspeedHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "creplay", Handler: creplayHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "cposition", Handler: cpositionHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "cshuffle", Handler: cshuffleHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
	{Pattern: "creload", Handler: creloadHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
}

var cbHandlers = []CbHandlerDef{
	{Pattern: "start", Handler: startCB},
	{Pattern: "help_cb", Handler: helpCB},
	{Pattern: "^lang:[a-z]", Handler: langCallbackHandler},
	{Pattern: `^help:(.+)`, Handler: helpCallbackHandler},

	{Pattern: "^close$", Handler: closeHandler},
	{Pattern: "^cancel$", Handler: cancelHandler},
	{Pattern: "^bcast_cancel$", Handler: broadcastCancelCB},

	{Pattern: `^room:(\w+)$`, Handler: roomHandle},
	{Pattern: "progress", Handler: emptyCBHandler},
}

func Init(bot *telegram.Client, assistants *core.AssistantManager) {
	bot.UpdatesGetState()
	assistants.ForEach(func(a *core.Assistant) {
		a.Client.UpdatesGetState()
	})

	for _, h := range handlers {
		bot.AddCommandHandler(h.Pattern, SafeMessageHandler(h.Handler), h.Filters...).SetGroup(100)
	}

	for _, h := range cbHandlers {
		bot.AddCallbackHandler(h.Pattern, SafeCallbackHandler(h.Handler), h.Filters...).SetGroup(90)
	}

	bot.On("edit:/eval", evalHandle).SetGroup(80)
	bot.On("edit:/ev", evalCommandHandler).SetGroup(80)

	bot.On("participant", handleParticipantUpdate).SetGroup(70)

	bot.AddActionHandler(handleActions).SetGroup(60)

	assistants.ForEach(func(a *core.Assistant) {
		a.Ntg.OnStreamEnd(ntgOnStreamEnd)
	})

	go MonitorRooms()

	if is, _ := database.GetAutoLeave(); is {
		go startAutoLeave()
	}

	if config.SetCmds && config.OwnerID != 0 {
		go setBotCommands(bot)
	}

	cplayCommands := []string{
		"/cfplay", "/vcplay", "/fvcplay",
		"/cpause", "/cresume", "/cskip", "/cstop",
		"/cmute", "/cunmute", "/cseek", "/cseekback",
		"/cjump", "/cremove", "/cclear", "/cmove",
		"/cspeed", "/creplay", "/cposition", "/cshuffle",
		"/cloop", "/cqueue", "/creload",
	}

	for _, cmd := range cplayCommands {
		baseCmd := "/" + cmd[2:] // Remove 'c' prefix
		if baseHelp, exists := helpTexts[baseCmd]; exists {
			helpTexts[cmd] = fmt.Sprintf(`<i>Channel play variant of %s</i>

<b>⚙️ Requires:</b>
First configure channel using: <code>/channelplay --set [channel_id]</code>

%s

<b>💡 Note:</b>
This command affects the linked channel's voice chat, not the current group.`, baseCmd, baseHelp)
		}
	}
}

func ntgOnStreamEnd(chatID int64, _ ntgcalls.StreamType, _ ntgcalls.StreamDevice) {
	onStreamEndHandler(chatID)
}

func setBotCommands(bot *telegram.Client) {
	// Set commands for normal users in private chats
	if _, err := bot.BotsSetBotCommands(&telegram.BotCommandScopeUsers{}, "", AllCommands.PrivateUserCommands); err != nil {
		gologging.Error("Failed to set PrivateUserCommands " + err.Error())
	}

	// Set commands for normal users in group chats
	if _, err := bot.BotsSetBotCommands(&telegram.BotCommandScopeChats{}, "", AllCommands.GroupUserCommands); err != nil {
		gologging.Error("Failed to set GroupUserCommands " + err.Error())
	}

	// Set commands for chat admins
	if _, err := bot.BotsSetBotCommands(
		&telegram.BotCommandScopeChatAdmins{},
		"",
		append(AllCommands.GroupUserCommands, AllCommands.GroupAdminCommands...),
	); err != nil {
		gologging.Error("Failed to set GroupAdminCommands " + err.Error())
	}

	// Set commands for sudo users in their private chat
	sudoers, err := database.GetSudoers()
	if err != nil {
		log.Printf("Failed to get sudoers for setting commands: %v", err)
	} else {
		sudoCommands := append(AllCommands.PrivateUserCommands, AllCommands.PrivateSudoCommands...)
		for _, sudoer := range sudoers {
			if _, err := bot.BotsSetBotCommands(&telegram.BotCommandScopePeer{
				Peer: &telegram.InputPeerUser{UserID: sudoer, AccessHash: 0},
			},
				"",
				sudoCommands,
			); err != nil {
				gologging.Error("Failed to set PrivateSudoCommands " + err.Error())
			}
		}
	}

	ownerCommands := append(AllCommands.PrivateUserCommands, AllCommands.PrivateSudoCommands...)
	ownerCommands = append(ownerCommands, AllCommands.PrivateOwnerCommands...)
	if _, err := bot.BotsSetBotCommands(&telegram.BotCommandScopePeer{
		Peer: &telegram.InputPeerUser{UserID: config.OwnerID, AccessHash: 0},
	}, "", ownerCommands); err != nil {
		gologging.Error("Failed to set PrivateOwnerCommands " + err.Error())
	}
}
