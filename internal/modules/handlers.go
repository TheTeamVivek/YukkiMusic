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
	"log"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/internal/database"
	"main/ubot"
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

	{Pattern: "(ac|active|activevc|activevoice)", Handler: activeHandler, Filters: []telegram.Filter{sudoOnlyFilter, ignoreChannelFilter}},
	{Pattern: "(maintenance|maint)", Handler: handleMaintenance, Filters: []telegram.Filter{ownerFilter, ignoreChannelFilter}},
	{Pattern: "logger", Handler: handleLogger, Filters: []telegram.Filter{sudoOnlyFilter, ignoreChannelFilter}},
	{Pattern: "autoleave", Handler: autoLeaveHandler, Filters: []telegram.Filter{sudoOnlyFilter, ignoreChannelFilter}},

	{Pattern: "help", Handler: helpHandler, Filters: []telegram.Filter{ignoreChannelFilter}},
	{Pattern: "ping", Handler: pingHandler, Filters: []telegram.Filter{ignoreChannelFilter}},
	{Pattern: "start", Handler: startHandler, Filters: []telegram.Filter{ignoreChannelFilter}},
	{Pattern: "stats", Handler: statsHandler, Filters: []telegram.Filter{ignoreChannelFilter, sudoOnlyFilter}},
	{Pattern: "bug", Handler: bugHandler, Filters: []telegram.Filter{ignoreChannelFilter}},
	{Pattern: "(lang|language)", Handler: langHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},

	// SuperGroup & Admin Filters
	{Pattern: "(play|vplay)", Handler: playHandler, Filters: []telegram.Filter{superGroupFilter}},
	{Pattern: "(fplay|forceplay)", Handler: fplayHandler, Filters: []telegram.Filter{superGroupFilter, authFilter}},
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
	{Pattern: "close", Handler: closeHandler},
	{Pattern: "cancel", Handler: cancelHandler},
	{Pattern: `^room:(\w+)$`, Handler: roomHandle},
	{Pattern: "progress", Handler: emptyCBHandler},
}

func Init(c, u *telegram.Client, n *ubot.Context) {
	c.UpdatesGetState()
	u.UpdatesGetState()

	for _, h := range handlers {
		if len(h.Filters) > 0 {
			c.On("command:"+h.Pattern, SafeMessageHandler(h.Handler), h.Filters...)
			// c.AddCommandHandler(h.Pattern, SafeMessageHandler(h.Handler), h.Filters...) //.SetGroup("commands")
		}
	}

	for _, h := range cbHandlers {
		c.AddCallbackHandler(h.Pattern, SafeCallbackHandler(h.Handler), h.Filters...) //.SetGroup("callback")
	}

	c.On("edit:/eval", evalHandle)       //.SetGroup("edit")
	c.On("edit:/ev", evalCommandHandler) //.SetGroup("edit")

	c.On("participant", handleParticipantUpdate) //.SetGroup("pu")

	c.AddActionHandler(handleActions) //.SetGroup("service_msg")

	n.OnStreamEnd(onStreamEndHandler)

	go MonitorRooms()
	if is, _ := database.GetAutoLeave(); is {
		go startAutoLeave()
	}
	if config.SetCmds && config.OwnerID != 0 {
		go setBotCommands(c)
	}
}

func setBotCommands(c *telegram.Client) {
	// Set commands for normal users in private chats
	if _, err := c.BotsSetBotCommands(&telegram.BotCommandScopeUsers{}, "", AllCommands.PrivateUserCommands); err != nil {
		gologging.Error("Failed to set PrivateUserCommands " + err.Error())
	}

	// Set commands for normal users in group chats
	if _, err := c.BotsSetBotCommands(&telegram.BotCommandScopeChats{}, "", AllCommands.GroupUserCommands); err != nil {
		gologging.Error("Failed to set GroupUserCommands " + err.Error())
	}

	// Set commands for chat admins
	if _, err := c.BotsSetBotCommands(
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
			if _, err := c.BotsSetBotCommands(&telegram.BotCommandScopePeer{
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
	if _, err := c.BotsSetBotCommands(&telegram.BotCommandScopePeer{
		Peer: &telegram.InputPeerUser{UserID: config.OwnerID, AccessHash: 0},
	}, "", ownerCommands); err != nil {
		gologging.Error("Failed to set PrivateOwnerCommands " + err.Error())
	}
}
