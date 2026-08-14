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
	"yukkimusic/internal/logger"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config"
	"yukkimusic/internal/database"
)

// BotCommands holds all bot commands separated by user type and chat type.
type BotCommands struct {
	PrivateUserCommands  []td.BotCommand
	PrivateSudoCommands  []td.BotCommand
	PrivateOwnerCommands []td.BotCommand
	GroupUserCommands    []td.BotCommand
	GroupAdminCommands   []td.BotCommand
}

func cmd(command, description string) td.BotCommand {
	return td.BotCommand{Command: command, Description: description}
}

var AllCommands = BotCommands{
	PrivateUserCommands: []td.BotCommand{
		cmd("start", "🚀 Start the bot"),
		cmd("help", "📖 Show the help menu"),
		cmd("ping", "🏓 Check if the bot is alive"),
	},

	PrivateSudoCommands: []td.BotCommand{
		cmd("ac", "📡 Show active voice chats"),
		cmd("stats", "📊 Show bot stats"),
		cmd("logger", "📝 Enable/disable logger channel"),
		cmd("autoleave", "⏳ Enable/disable auto leave"),
	},

	PrivateOwnerCommands: []td.BotCommand{
		cmd("addsudo", "➕ Add a sudo user"),
		cmd("delsudo", "➖ Remove a sudo user"),
		cmd("blockuser", "🚫 Block a user"),
		cmd("unblockuser", "✅ Unblock a user"),
		cmd("blockchat", "⛔ Block a chat"),
		cmd("unblockchat", "✅ Unblock a chat"),
		cmd("blacklisted", "📛 List blacklisted chats and users"),
		cmd("maintenance", "🛠️ Enable/disable maintenance mode"),
	},

	GroupUserCommands: []td.BotCommand{
		cmd("play", "🎵 Play a song"),
		cmd("queue", "📜 Show the queue"),
		cmd("position", "⏱️ Show the current position of the song"),
		cmd("reload", "🔄 Reload the admin cache"),
		cmd("authlist", "👥 List authorized users"),
	},

	GroupAdminCommands: []td.BotCommand{
		// Playback
		cmd("play", "🎵 Play a song"),
		cmd("cplay", "🎵 Play in the linked channel"),
		cmd("fplay", "⏩ Force play a song"),
		cmd("cfplay", "⏩ Force play in the linked channel"),

		// Pause / Resume
		cmd("pause", "⏸️ Pause the current song"),
		cmd("cpause", "⏸️ Pause in the linked channel"),
		cmd("resume", "▶️ Resume the current song"),
		cmd("cresume", "▶️ Resume in the linked channel"),

		// Skip
		cmd("skip", "⏭️ Skip the current song"),
		cmd("cskip", "⏭️ Skip in the linked channel"),

		// Replay
		cmd("replay", "🔁 Replay the current song"),
		cmd("creplay", "🔁 Replay in the linked channel"),

		// End / Stop
		cmd("end", "⏹️ Stop the song and leave voice chat"),
		cmd("cstop", "⏹️ Stop and leave the linked channel's voice chat"),

		// Mute / Unmute
		cmd("mute", "🔇 Mute the bot in the voice chat"),
		cmd("unmute", "🔊 Unmute the bot in the voice chat"),
		cmd("cmute", "🔇 Mute in the linked channel's voice chat"),
		cmd("cunmute", "🔊 Unmute in the linked channel's voice chat"),

		// Seek
		cmd("seek", "⏩ Seek to a specific position"),
		cmd("seekback", "⏪ Seek back in the song"),
		cmd("cseek", "⏩ Seek in the linked channel's song"),
		cmd("cseekback", "⏪ Seek back in the linked channel's song"),

		// Speed
		cmd("speed", "⚡ Set the playback speed"),
		cmd("cspeed", "⚡ Set the playback speed in the linked channel"),

		// Queue management
		cmd("queue", "📜 Show the queue"),
		cmd("cqueue", "📜 Show the linked channel's queue"),
		cmd("position", "⏱️ Show the current position of the song"),
		cmd("cposition", "⏱️ Show the current position in the linked channel"),
		cmd("jump", "🎯 Jump to a specific song in the queue"),
		cmd("cjump", "🎯 Jump to a song in the linked channel's queue"),
		cmd("remove", "🗑️ Remove a song from the queue"),
		cmd("cremove", "🗑️ Remove a song from the linked channel's queue"),
		cmd("move", "🔄 Move a song in the queue"),
		cmd("cmove", "🔄 Move a song in the linked channel's queue"),
		cmd("clear", "🧹 Clear the queue"),
		cmd("cclear", "🧹 Clear the linked channel's queue"),
		cmd("shuffle", "🔀 Shuffle the queue"),
		cmd("cshuffle", "🔀 Shuffle the linked channel's queue"),
		cmd("loop", "🔁 Loop the current song"),
		cmd("cloop", "🔁 Loop the current song in the linked channel"),

		cmd("setcplay", "⚙️ Configure channelplay for your chat"),

		// Settings & access
		cmd("playmode", "🎛️ Control who can use /play"),
		cmd("adminmode", "🛡️ Control who can use admin music commands"),
		cmd("cmddelete", "🧽 Toggle automatic deletion of bot commands"),
		cmd("settings", "⚙️ Configure chat settings"),
		cmd("addauth", "➕ Add a user to the authorized list"),
		cmd("delauth", "➖ Remove a user from the authorized list"),
		cmd("reload", "🔄 Reload the admin cache"),
		cmd("creload", "🔄 Reload the admin cache in the linked channel"),
	},
}

func setBotCommands(bot *td.Client) {
	type scopedCmds struct {
		scope td.BotCommandScope
		cmds  []td.BotCommand
	}

	entries := []scopedCmds{
		{&td.BotCommandScopeAllPrivateChats{}, AllCommands.PrivateUserCommands},
		{&td.BotCommandScopeAllGroupChats{}, AllCommands.GroupUserCommands},
		{
			&td.BotCommandScopeAllChatAdministrators{},
			append(AllCommands.GroupUserCommands, AllCommands.GroupAdminCommands...),
		},
		{
			&td.BotCommandScopeChat{ChatId: config.OwnerID},
			append(
				append(AllCommands.PrivateUserCommands, AllCommands.PrivateSudoCommands...),
				AllCommands.PrivateOwnerCommands...,
			),
		},
	}

	for _, e := range entries {
		if err := bot.SetCommands(e.cmds, "", &td.SetCommandsOpts{Scope: e.scope}); err != nil {
			logger.Errorf("Failed to set bot commands(Scope: %T): %v", e.scope, err)
		}
	}

	// Sudo users get their own command scope in private.
	sudoers, err := database.Sudoers()
	if err != nil {
		logger.Error("Failed to fetch sudoers: " + err.Error())
		return
	}

	sudoCmds := append(AllCommands.PrivateUserCommands, AllCommands.PrivateSudoCommands...)
	for _, id := range sudoers {
		scope := &td.BotCommandScopeChat{ChatId: id}
		if err := bot.SetCommands(sudoCmds, "", &td.SetCommandsOpts{Scope: scope}); err != nil {
			logger.Error("Failed to set sudo commands: " + err.Error())
		}
	}
}
