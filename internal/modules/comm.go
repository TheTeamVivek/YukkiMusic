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

import "github.com/amarnathcjd/gogram/telegram"

// BotCommands is a struct that holds all the bot commands
// separated by user type and chat type.
type BotCommands struct {
	// For private chats
	PrivateUserCommands  []*telegram.BotCommand
	PrivateSudoCommands  []*telegram.BotCommand
	PrivateOwnerCommands []*telegram.BotCommand
	// For group chats
	GroupUserCommands  []*telegram.BotCommand
	GroupAdminCommands []*telegram.BotCommand
}

// AllCommands holds all the bot commands.
var AllCommands = BotCommands{
	// Commands for private chats
	PrivateUserCommands: []*telegram.BotCommand{
		{Command: "start", Description: "Start the bot."},
		{Command: "help", Description: "Show help menu."},
		{Command: "ping", Description: "Check if the bot is alive."},
		{Command: "sudolist", Description: "List sudo users."},
	},
	PrivateSudoCommands: []*telegram.BotCommand{
		{Command: "ac", Description: "Show active voice chats."},
		{Command: "stats", Description: "Show bot stats."},

		{Command: "logger", Description: "Enable/disable logger channel."},
		{Command: "autoleave", Description: "Enable/disable auto leave."},
	},
	PrivateOwnerCommands: []*telegram.BotCommand{
		{Command: "addsudo", Description: "Add a sudo user."},
		{Command: "delsudo", Description: "Remove a sudo user."},
		{
			Command:     "maintenance",
			Description: "Enable/disable maintenance mode.",
		},
	},
	// Commands for group chats
	GroupUserCommands: []*telegram.BotCommand{
		{Command: "play", Description: "Play a song."},
		{Command: "queue", Description: "Show the queue."},
		{
			Command:     "position",
			Description: "Show the current position of the song.",
		},

		{Command: "reload", Description: "Reload the admin cache."},
		{Command: "authlist", Description: "List authorized users."},
		{Command: "help", Description: "Show help menu."},
		{Command: "ping", Description: "Check if the bot is alive."},
	},
	GroupAdminCommands: []*telegram.BotCommand{
		{Command: "cplay", Description: "Play a song in the linked channel."},
		{
			Command:     "cqueue",
			Description: "Show the queue in the linked channel.",
		},
		{
			Command:     "cposition",
			Description: "Show the current position of the song in the linked channel.",
		},
		{Command: "fplay", Description: "Force play a song."},
		{Command: "speed", Description: "Set the speed of the song."},
		{Command: "skip", Description: "Skip the current song."},
		{Command: "pause", Description: "Pause the current song."},
		{Command: "resume", Description: "Resume the current song."},
		{Command: "replay", Description: "Replay the current song."},
		{Command: "mute", Description: "Mute the bot in the voice chat."},
		{Command: "unmute", Description: "Unmute the bot in the voice chat."},
		{
			Command:     "seek",
			Description: "Seek to a specific position in the song.",
		},
		{
			Command:     "seekback",
			Description: "Seek back to a specific position in the song.",
		},
		{Command: "jump", Description: "Jump to a specific song in the queue."},
		{Command: "clear", Description: "Clear the queue."},
		{Command: "remove", Description: "Remove a song from the queue."},
		{Command: "move", Description: "Move a song in the queue."},
		{Command: "shuffle", Description: "Shuffle the queue."},
		{Command: "loop", Description: "Loop the current song."},
		{Command: "end", Description: "Stop the song."},
		{Command: "addauth", Description: "Add a user to the authorized list."},
		{
			Command:     "delauth",
			Description: "Remove a user from the authorized list.",
		},
		{
			Command:     "channelplay",
			Description: "Set a channel as the play channel.",
		},
		{
			Command:     "cfplay",
			Description: "Force play a song in the linked channel.",
		},
		{
			Command:     "cpause",
			Description: "Pause the current song in the linked channel.",
		},
		{
			Command:     "cresume",
			Description: "Resume the current song in the linked channel.",
		},
		{
			Command:     "cmute",
			Description: "Mute the bot in the linked channel's voice chat.",
		},
		{
			Command:     "cunmute",
			Description: "Unmute the bot in the linked channel's voice chat.",
		},
		{
			Command:     "cstop",
			Description: "Stop the current song and leave the linked channel's voice chat.",
		},
		{
			Command:     "cskip",
			Description: "Skip the current song in the linked channel.",
		},
		{
			Command:     "cloop",
			Description: "Loop the current song in the linked channel.",
		},
		{
			Command:     "cseek",
			Description: "Seek to a specific position in the song in the linked channel.",
		},
		{
			Command:     "cseekback",
			Description: "Seek back to a specific position in the song in the linked channel.",
		},
		{
			Command:     "cjump",
			Description: "Jump to a specific song in the linked channel's queue.",
		},
		{
			Command:     "cremove",
			Description: "Remove a song from the linked channel's queue.",
		},
		{Command: "cclear", Description: "Clear the linked channel's queue."},
		{
			Command:     "cmove",
			Description: "Move a song in the linked channel's queue.",
		},
		{
			Command:     "cspeed",
			Description: "Set the speed of the song in the linked channel.",
		},
		{
			Command:     "creplay",
			Description: "Replay the current song in the linked channel.",
		},
		{
			Command:     "cshuffle",
			Description: "Shuffle the linked channel's queue.",
		},
		{
			Command:     "creload",
			Description: "Reload the admin cache in the linked channel.",
		},
	},
}
