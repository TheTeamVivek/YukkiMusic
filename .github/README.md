<h1 align="center">🎧 <b>YukkiMusic</b></h1>

<p align="center">
  <i>⚡ A blazing-fast, reliable, and feature-packed Telegram bot for streaming music in group voice chats — built with Go.</i>
</p>
<p align="center">

  <a href="https://go.dev/">
    <img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go&labelColor=000000&logoColor=white" alt="Go Version">
  </a>
  <a href="https://github.com/TheTeamVivek/YukkiMusic/releases/tag/v2.0">
    <img src="https://img.shields.io/badge/Version-v2.0-FF9800?style=for-the-badge&logo=semver&labelColor=000000&logoColor=white" alt="Version">
  </a>
  <a href="../LICENSE">
    <img src="https://img.shields.io/badge/License-GPLv3-FF3860?style=for-the-badge&logo=gnu&labelColor=000000&logoColor=white" alt="License: GPLv3">
  </a>
  <a href="https://github.com/TheTeamVivek/YukkiMusic/stargazers">
    <img src="https://img.shields.io/github/stars/TheTeamVivek/YukkiMusic?style=for-the-badge&label=Stars&labelColor=000000&color=FFD700&logo=github&logoColor=white" alt="GitHub Stars">
  </a>
  <a href="https://github.com/TheTeamVivek/YukkiMusic/fork">
    <img src="https://img.shields.io/github/forks/TheTeamVivek/YukkiMusic?style=for-the-badge&label=Forks&labelColor=000000&color=00C853&logo=github&logoColor=white" alt="GitHub Forks">
  </a>

</p>

---

## ✨ Features

### 🎵 Music Playback
- **YouTube Integration** - Search and stream YouTube videos directly
- **Telegram Media** - Play audio/video files shared in Telegram
- **Speed Control** - Adjust playback speed (0.5x - 4.0x)
- **Quality Streaming** - High-quality audio output with adaptive bitrates

### 🎮 Playback Control
- ▶️ Play / Resume / Pause
- ⏭️ Skip / Jump to position
- 🔁 Loop current track / Replay
- 🔀 Shuffle queue
- 🔇 Mute / Unmute with auto-resume
- ⏱️ Seek forward/backward

### 👥 User Management
- **Admin Controls** - Full playback management for group admins
- **Auth Users** - Grant non-admin users playback permissions
- **Multi-Account Support** - Multiple assistant accounts for load balancing
- **Permission System** - Owner, Sudoers, Admins, Auth Users hierarchy

### 📊 Queue Management
- Move tracks in queue
- Remove specific tracks
- Clear entire queue
- View queue with pagination

### 🌍 Localization
- **Multi-Language Support** - English, Hindi, and more
- **Per-Chat Language Settings** - Different languages for different groups
- User-friendly messages in preferred language

### 🔐 Advanced Features
- **Maintenance Mode** - Temporarily disable bot with custom message
- **Auto-Leave** - Automatically leave inactive chats
- **Broadcast** - Send messages to all users/chats at once
- **Logger Channel** - Log all bot activities and errors
- **Channel Play** - Stream to linked channels with separate control

---

## 🚀 Quick Start


### ☁️ Deploy to Heroku (One-Click)

Click the button below to deploy **YukkiMusic** instantly on Heroku:

<a href="https://heroku.com/deploy?template=https://github.com/TheTeamVivek/YukkiMusic">
  <img src="https://www.herokucdn.com/deploy/button.svg" alt="Deploy to Heroku">
</a>

---

### Prerequisites
- Go 1.25 or higher
- MongoDB (Cloud or Local)
- Telegram Bot Token (from [@BotFather](https://t.me/BotFather))
- API ID & Hash (from [my.telegram.org](https://my.telegram.org))
- Assistant Account Session String

### Installation

1. **Clone Repository**
```bash
git clone https://github.com/TheTeamVivek/YukkiMusic.git
cd YukkiMusic
```

2. **Install Dependencies**
```bash
bash install.sh && go mod tidy
```

3. **Configure Environment**
```bash
cp sample.env .env
# Edit .env with your credentials
```

4. **Get Required Credentials**
- **Bot Token**: Message [@BotFather](https://t.me/BotFather), use `/newbot`
- **API ID & Hash**: Visit [my.telegram.org](https://my.telegram.org)
- **Session String**: Use [@StringFatherBot](https://t.me/StringFatherBot) or online generator
- **MongoDB**: Free tier at [MongoDB Atlas](https://www.mongodb.com/cloud/atlas)

5. **Start the Bot**
```bash
go run main.go
```

---

## ⚙️ Configuration

All configuration is managed through environment variables. For detailed configuration instructions, see:

📖 **[Configuration Guide](../internal/config/README.md)**

### Key Variables
| Variable | Required | Purpose |
|----------|----------|---------|
| `API_ID` | ✅ | Telegram API ID |
| `API_HASH` | ✅ | Telegram API Hash |
| `TOKEN` | ✅ | Bot token from @BotFather |
| `MONGO_DB_URI` | ✅ | MongoDB connection string |
| `STRING_SESSIONS` | ✅ | Assistant account session strings |
| `OWNER_ID` | ❌ | Your Telegram User ID |
| `LOGGER_ID` | ❌ | Log channel ID |

See **[Configuration Reference](../internal/config/README.md)** for complete list of variables with examples.

---

## 📚 Commands

### User Commands
| Command | Description |
|---------|-------------|
| `/play <query>` | Play a song from YouTube or Telegram |
| `/queue` | Show current queue |
| `/position` | Show current track position |
| `/help` | Show command help |
| `/ping` | Check bot status |

### Admin Commands
| Command | Description |
|---------|-------------|
| `/pause [seconds]` | Pause for playback (optionally auto-resume) |
| `/resume` | Resume paused track |
| `/mute [seconds]` | Mute playback (optionally auto-unmute) |
| `/unmute` | Unmute playback |
| `/seek <seconds>` | Seek to specific position |
| `/loop <count>` | Loop track N times |
| `/shuffle` | Toggle shuffle mode |
| `/speed <speed>` | Set playback speed (0.5-4.0x) |
| `/skip` | Skip to next track |
| `/fplay <query>` | Force play (skip queue) |
| `/clear` | Clear entire queue |
| `/remove <index>` | Remove track from queue |
| `/move <from> <to>` | Move track in queue |
| `/jump <position>` | Jump to position in current track |
| `/replay` | Replay current track |
| `/addauth <user>` | Grant user playback permission |
| `/delauth <user>` | Revoke user playback permission |
| `/authlist` | List authorized users |
| `/reload` | Reload admin cache |
| `/cplay` | Channel play mode |

### Owner Commands
| Command | Description |
|---------|-------------|
| `/addsudo <user>` | Add sudo user |
| `/delsudo <user>` | Remove sudo user |
| `/sudolist` | List all sudo users |
| `/maintenance <on/off>` | Toggle maintenance mode |
| `/broadcast <message>` | Broadcast to all chats |
| `/stats` | Show bot statistics |
| `/restart` | Restart bot |

---

## 🎼 Platform System

YukkiMusic uses a **modular platform system** to support multiple music sources:

### Supported Platforms
1. **Telegram** (Priority: 100) - Direct Telegram audio/video files
2. **YouTube** (Priority: 90) - YouTube videos and playlists
3. **Youtubify API** (Priority: 80) - Premium YouTube downloads
4. **Fallen API** (Priority: 75) - YouTube downloads via Fallen API
5. **YT-DLP** (Priority: 70) - Direct yt-dlp integration

### How It Works
- When you request a song, the bot checks each platform by **priority**
- **First valid platform handles the request**
- Automatic fallback if one method fails
- Seamless track fetching and downloading

📖 **[Platform System Guide](../internal/platforms/README.md)** - Learn how to add custom platforms

---

## 📖 Documentation

- **[Configuration Guide](../internal/config/README.md)** - All environment variables explained
- **[Platform System](../internal/platforms/README.md)** - How platforms work and add custom sources
- **[Database Layer](../internal/database/README.md)** - MongoDB schemas, queries, and bot data management
- **[Modules System](../internal/modules/README.md)** - Command handlers, permissions, and feature modules
---

## 🏗️ Project Structure

```
YukkiMusic/
├── .github/
│   └── README.md              # Main documentation (you are here)
├── internal/
│   ├── config/                # Configuration management
│   │   ├── config.go
│   │   └── README.md          # Detailed config guide
│   ├── core/                  # Core bot logic
│   │   ├── clients.go         # Bot & assistant initialization
│   │   ├── room_state.go      # Playback state management
│   │   ├── chat_state.go      # Chat & assistant state
│   │   └── ...
│   ├── database/              # MongoDB operations
│   │   ├── bot_state.go
│   │   ├── chat_settings.go
│   │   └── ...
│   ├── modules/               # Command handlers
│   │   ├── play.go            # Play command
│   │   ├── queue.go           # Queue management
│   │   └── ...
│   ├── platforms/             # Music sources
│   │   ├── youtube.go         # YouTube integration
│   │   ├── telegram.go        # Telegram media
│   │   ├── ytdlp.go           # YT-DLP downloader
│   │   └── ...
│   ├── locales/               # Multi-language support
│   │   ├── en.yml            # English
│   │   ├── hi.yml            # Hindi
│   │   └── ...
│   ├── utils/                 # Utility functions
│   ├── cookies/               # YouTube cookie files
│   └── ...
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
└── main.go                    # Entry point
```

---

## 🐛 Bug Reports & Features

Found a bug? Have a feature request?

- **Use `/bug` command** in the bot to report directly
- **Join our [Support Chat](https://t.me/TheTeamVk)** for discussions
- **Open an [Issue on GitHub](https://github.com/TheTeamVivek/YukkiMusic/issues)**

---

## 🤝 Contributing

Contributions are welcome! Here's how to help:

1. **Fork the repository**
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Commit changes** (`git commit -m 'Add amazing feature'`)
4. **Push to branch** (`git push origin feature/amazing-feature`)
5. **Open a Pull Request**

### Adding a New Platform
See **[Platform System Guide](../internal/platforms/README.md#-how-to-add-a-new-platform)** for step-by-step instructions.

---

## 📜 License

This project is licensed under the **GNU General Public License v3.0** - see the [LICENSE](LICENSE) file for details.

```
YukkiMusic — A Telegram bot that streams music into group voice chats
Copyright (C) 2025 TheTeamVivek

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
```

---

## 🙌 Credits

- **Maintainer**: [Vivek Kumar](https://github.com/Vivekkumar-in)
- **Contributors**: All amazing developers who contributed to this project
- **Libraries**: Built with [gogram](https://github.com/AmarnathCJD/gogram), [ntgcalls](https://github.com/pytgcalls/ntgcalls), and more

---

## 📞 Support

- **Telegram Support Chat**: [@TheTeamVk](https://t.me/TheTeamVk)
- **Updates Channel**: [@TheTeamVivek](https://t.me/TheTeamVivek)
- **GitHub Issues**: [Report bugs](https://github.com/TheTeamVivek/YukkiMusic/issues)

---

## ⚡ Performance Tips

1. **Use multiple assistant accounts** - Distributes load across accounts
2. **Set appropriate limits** - Adjust `QUEUE_LIMIT` and `DURATION_LIMIT`
3. **Enable auto-leave** - Removes bot from inactive chats automatically
4. **Use MongoDB Atlas** - Better performance than local MongoDB
5. **Set up logger** - Monitor errors and optimize accordingly

---

**Made with ❤️ by TheTeamVivek**

**Happy Streaming! 🎶**
