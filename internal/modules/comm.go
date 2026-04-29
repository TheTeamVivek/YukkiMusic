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
	"log"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/database"
)

// BotCommands holds all bot commands separated by user type and chat type.
type BotCommands struct {
	PrivateUserCommands  []*telegram.BotCommand
	PrivateSudoCommands  []*telegram.BotCommand
	PrivateOwnerCommands []*telegram.BotCommand
	GroupUserCommands    []*telegram.BotCommand
	GroupAdminCommands   []*telegram.BotCommand
}

func cmd(command, description string) *telegram.BotCommand {
	return &telegram.BotCommand{Command: command, Description: description}
}

var AllCommands = BotCommands{
	PrivateUserCommands: []*telegram.BotCommand{
		cmd("start", "Start the bot."),
		cmd("help", "Show help menu."),
		cmd("ping", "Check if the bot is alive."),
	},

	PrivateSudoCommands: []*telegram.BotCommand{
		cmd("ac", "Show active voice chats."),
		cmd("stats", "Show bot stats."),
		cmd("logger", "Enable/disable logger channel."),
		cmd("autoleave", "Enable/disable auto leave."),
	},

	PrivateOwnerCommands: []*telegram.BotCommand{
		cmd("addsudo", "Add a sudo user."),
		cmd("delsudo", "Remove a sudo user."),
		cmd("maintenance", "Enable/disable maintenance mode."),
	},

	GroupUserCommands: []*telegram.BotCommand{
		cmd("play", "Play a song."),
		cmd("queue", "Show the queue."),
		cmd("position", "Show the current position of the song."),
		cmd("reload", "Reload the admin cache."),
		cmd("authlist", "List authorized users."),
	},

	GroupAdminCommands: []*telegram.BotCommand{
		// Playback
		cmd("play", "Play a song."),
		cmd("cplay", "Play a song in the linked channel."),
		cmd("fplay", "Force play a song."),
		cmd("cfplay", "Force play a song in the linked channel."),

		// Pause / Resume
		cmd("pause", "Pause the current song."),
		cmd("cpause", "Pause the current song in the linked channel."),
		cmd("resume", "Resume the current song."),
		cmd("cresume", "Resume the current song in the linked channel."),

		// Skip
		cmd("skip", "Skip the current song."),
		cmd("cskip", "Skip the current song in the linked channel."),

		// Replay
		cmd("replay", "Replay the current song."),
		cmd("creplay", "Replay the current song in the linked channel."),

		// End / Stop
		cmd("end", "Stop the song and leave voice chat."),
		cmd("cstop", "Stop the song and leave the linked channel's voice chat."),

		// Mute / Unmute
		cmd("mute", "Mute the bot in the voice chat."),
		cmd("unmute", "Unmute the bot in the voice chat."),
		cmd("cmute", "Mute the bot in the linked channel's voice chat."),
		cmd("cunmute", "Unmute the bot in the linked channel's voice chat."),

		// Seek
		cmd("seek", "Seek to a specific position in the song."),
		cmd("seekback", "Seek back in the song."),
		cmd("cseek", "Seek to a position in the linked channel's song."),
		cmd("cseekback", "Seek back in the linked channel's song."),

		// Speed
		cmd("speed", "Set the playback speed."),
		cmd("cspeed", "Set the playback speed in the linked channel."),

		// Queue management
		cmd("queue", "Show the queue."),
		cmd("cqueue", "Show the linked channel's queue."),
		cmd("position", "Show the current position of the song."),
		cmd("cposition", "Show the current position in the linked channel."),
		cmd("jump", "Jump to a specific song in the queue."),
		cmd("cjump", "Jump to a song in the linked channel's queue."),
		cmd("remove", "Remove a song from the queue."),
		cmd("cremove", "Remove a song from the linked channel's queue."),
		cmd("move", "Move a song in the queue."),
		cmd("cmove", "Move a song in the linked channel's queue."),
		cmd("clear", "Clear the queue."),
		cmd("cclear", "Clear the linked channel's queue."),
		cmd("shuffle", "Shuffle the queue."),
		cmd("cshuffle", "Shuffle the linked channel's queue."),
		cmd("loop", "Loop the current song."),
		cmd("cloop", "Loop the current song in the linked channel."),

		cmd("setcplay", "Configure channelplay for your chat."),

		// Settings & access
		cmd("playmode", "Control who can use /play."),
		cmd("adminmode", "Control who can use admin music commands."),
		cmd("cmddelete", "Toggle automatic deletion of bot commands."),
		cmd("settings", "Configure chat settings."),
		cmd("addauth", "Add a user to the authorized list."),
		cmd("delauth", "Remove a user from the authorized list."),
		cmd("reload", "Reload the admin cache."),
		cmd("creload", "Reload the admin cache in the linked channel."),
	},
}

func setBotCommands(bot *telegram.Client) {
	type scopedCmds struct {
		scope telegram.BotCommandScope
		cmds  []*telegram.BotCommand
	}

	entries := []scopedCmds{
		{&telegram.BotCommandScopeUsers{}, AllCommands.PrivateUserCommands},
		{&telegram.BotCommandScopeChats{}, AllCommands.GroupUserCommands},
		{
			&telegram.BotCommandScopeChatAdmins{},
			append(AllCommands.GroupUserCommands, AllCommands.GroupAdminCommands...),
		},
		{
			&telegram.BotCommandScopePeer{
				Peer: &telegram.InputPeerUser{UserID: config.OwnerID},
			},
			append(
				append(AllCommands.PrivateUserCommands, AllCommands.PrivateSudoCommands...),
				AllCommands.PrivateOwnerCommands...,
			),
		},
	}

	for _, e := range entries {
		if _, err := bot.BotsSetBotCommands(e.scope, "", e.cmds); err != nil {
			gologging.Error("Failed to set bot commands: " + err.Error())
		}
	}

	// Sudo users get their own command scope in private.
	sudoers, err := database.Sudoers()
	if err != nil {
		log.Printf("Failed to fetch sudoers: %v", err)
		return
	}

	sudoCmds := append(AllCommands.PrivateUserCommands, AllCommands.PrivateSudoCommands...)
	for _, id := range sudoers {
		peer := &telegram.BotCommandScopePeer{
			Peer: &telegram.InputPeerUser{UserID: id},
		}
		if _, err := bot.BotsSetBotCommands(peer, "", sudoCmds); err != nil {
			gologging.Error("Failed to set sudo commands: " + err.Error())
		}
	}
}
