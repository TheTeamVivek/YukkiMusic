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
		{"start", "Start the bot."},
		{"help", "Show help menu."},
		{"ping", "Check if the bot is alive."},
		{"sudolist", "List sudo users."},
	},
	PrivateSudoCommands: []*telegram.BotCommand{
		{"ac", "Show active voice chats."},
		{"stats", "Show bot stats."},

		{"logger", "Enable/disable logger channel."},
		{"autoleave", "Enable/disable auto leave."},
	},
	PrivateOwnerCommands: []*telegram.BotCommand{
		{"addsudo", "Add a sudo user."},
		{"delsudo", "Remove a sudo user."},
		{"maintenance", "Enable/disable maintenance mode."},
	},
	// Commands for group chats
	GroupUserCommands: []*telegram.BotCommand{
		{"play", "Play a song."},
		{"queue", "Show the queue."},
		{"position", "Show the current position of the song."},

		{"reload", "Reload the admin cache."},
		{"authlist", "List authorized users."},
		{"help", "Show help menu."},
		{"ping", "Check if the bot is alive."},
	},
	GroupAdminCommands: []*telegram.BotCommand{
		{"cplay", "Play a song in the linked channel."},
		{"cqueue", "Show the queue in the linked channel."},
		{"cposition", "Show the current position of the song in the linked channel."},
		{"fplay", "Force play a song."},
		{"speed", "Set the speed of the song."},
		{"skip", "Skip the current song."},
		{"pause", "Pause the current song."},
		{"resume", "Resume the current song."},
		{"replay", "Replay the current song."},
		{"mute", "Mute the bot in the voice chat."},
		{"unmute", "Unmute the bot in the voice chat."},
		{"seek", "Seek to a specific position in the song."},
		{"seekback", "Seek back to a specific position in the song."},
		{"jump", "Jump to a specific song in the queue."},
		{"clear", "Clear the queue."},
		{"remove", "Remove a song from the queue."},
		{"move", "Move a song in the queue."},
		{"shuffle", "Shuffle the queue."},
		{"loop", "Loop the current song."},
		{"end", "Stop the song."},
		{"addauth", "Add a user to the authorized list."},
		{"delauth", "Remove a user from the authorized list."},
		{"channelplay", "Set a channel as the play channel."},
		{"cfplay", "Force play a song in the linked channel."},
		{"cpause", "Pause the current song in the linked channel."},
		{"cresume", "Resume the current song in the linked channel."},
		{"cmute", "Mute the bot in the linked channel's voice chat."},
		{"cunmute", "Unmute the bot in the linked channel's voice chat."},
		{"cstop", "Stop the current song and leave the linked channel's voice chat."},
		{"cskip", "Skip the current song in the linked channel."},
		{"cloop", "Loop the current song in the linked channel."},
		{"cseek", "Seek to a specific position in the song in the linked channel."},
		{"cseekback", "Seek back to a specific position in the song in the linked channel."},
		{"cjump", "Jump to a specific song in the linked channel's queue."},
		{"cremove", "Remove a song from the linked channel's queue."},
		{"cclear", "Clear the linked channel's queue."},
		{"cmove", "Move a song in the linked channel's queue."},
		{"cspeed", "Set the speed of the song in the linked channel."},
		{"creplay", "Replay the current song in the linked channel."},
		{"cshuffle", "Shuffle the linked channel's queue."},
		{"creload", "Reload the admin cache in the linked channel."},
	},
}
