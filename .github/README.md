<div align="center">

# 🎧 YukkiMusic

**⚡ A blazing-fast, reliable Telegram bot for streaming music in group voice chats — built with Go.**

[![GitHub Stars](https://img.shields.io/github/stars/TheTeamVivek/YukkiMusic?style=for-the-badge&label=Stars&color=FFD700&logo=github&logoColor=white)](https://github.com/TheTeamVivek/YukkiMusic/stargazers)
[![GitHub Forks](https://img.shields.io/github/forks/TheTeamVivek/YukkiMusic?style=for-the-badge&label=Forks&color=00C853&logo=github&logoColor=white)](https://github.com/TheTeamVivek/YukkiMusic/fork)
[![Go Version](https://img.shields.io/badge/Go-1.27+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Version](https://img.shields.io/badge/Version-v3.3.0-FF9800?style=for-the-badge&logo=semver&logoColor=white)](https://github.com/TheTeamVivek/YukkiMusic/releases)
[![Support](https://img.shields.io/badge/Telegram-Support-26A69A?style=for-the-badge&logo=telegram&logoColor=white)](https://t.me/TheTeamVk)

<br><br>

**🚀 Deploy your own instance in one click:**

<a href="https://heroku.com/deploy?template=https://github.com/TheTeamVivek/YukkiMusic">
  <img src="https://www.herokucdn.com/deploy/button.svg" alt="Deploy to Heroku">
</a>

<br>

</div>

---

## ✨ Features

| | |
| :---: | :--- |
| 🚀 | **High Performance** — optimized Go engine for ultra-low latency playback |
| 🎼 | **Smart Providers** — YouTube, Telegram files & custom APIs with automatic fallback |
| 🌍 | **Localization** — English, हिन्दी, العربية, Türkçe out of the box |
| 🤖 | **Multi-Assistant** — multiple Telegram accounts, auto-distributed across chats |
| 🛠 | **Admin Suite** — speed, seek, loop, video mode & advanced queue management |
| 🍪 | **Cookies Support** — bypass YouTube bot-detection with ease |

---

## 🚀 Quick Start

### 💻 Manual Installation

```bash
# 1. Clone & enter
git clone https://github.com/TheTeamVivek/YukkiMusic.git && cd YukkiMusic

# 2. Setup environment
bash install.sh
cp sample.env .env   # then fill API_ID, API_HASH, BOT_TOKEN, MONGO_DB_URI, STRING_SESSIONS

# 3. Tidy & run
go mod tidy
go run ./cmd/app
```

### 🐳 Docker Deployment

```bash
# 1. Clone & enter
git clone https://github.com/TheTeamVivek/YukkiMusic.git && cd YukkiMusic

# 2. Prepare environment
cp sample.env .env   # fill required values

# 3. Build and start in background
docker compose up -d --build
```

<details>
<summary><b>📜 Docker commands cheat-sheet</b></summary>

```bash
docker compose logs -f yukkimusic          # live logs
docker compose logs --tail=200 yukkimusic  # last 200 lines
docker compose ps                          # running state
docker compose restart yukkimusic          # restart
docker compose stop yukkimusic             # stop (keeps containers)
docker compose start yukkimusic            # start again
docker compose down                        # full cleanup
git pull && docker compose up -d --build   # update
```

</details>

---

## ⚙️ Configuration

All settings are managed via environment variables in your `.env` file.

| Variable | Description |
| :--- | :--- |
| `API_ID` / `API_HASH` | From [my.telegram.org](https://my.telegram.org) |
| `BOT_TOKEN` | Bot token from [@BotFather](https://t.me/BotFather) (alias: `TOKEN`) |
| `MONGO_DB_URI` | MongoDB Atlas or local connection string |
| `STRING_SESSIONS` | Assistant account session strings |

> [!IMPORTANT]
> 📖 **Docs you'll need**
> - 👉 [Configuration Guide](../config/README.md) — every env variable with defaults
> - 🍪 [Cookies Guide](../config/cookies/README.md) — set up YouTube cookies
> - 🌍 [Languages Guide](../internal/locales/README.md) — add a new language

---

## 🎮 Commands

### 🎵 Playback & Queue

| Command | Description | | Command | Description |
| :--- | :--- | :---: | :--- | :--- |
| `/play <query>` | Play from YouTube / TG | | `/queue` | View upcoming tracks |
| `/fplay <query>` | Force play (skip current) | | `/cqueue` | Channel queue view |
| `/vplay` / `/fvplay` | Video mode play | | `/remove` | Remove a track |
| `/pause` / `/resume` | Toggle playback | | `/move` | Reorder a track |
| `/mute` / `/unmute` | Toggle audio | | `/clear` | Clear the queue |
| `/skip` | Next in queue | | `/seek` / `/seekback` | Seek forward / back |
| `/stop` / `/end` | Stop & leave VC | | `/loop` · `/speed` | Loop count · tempo |

### 🛡️ Admin & Owner

| Command | Description | | Command | Description |
| :--- | :--- | :---: | :--- | :--- |
| `/addsudo` / `/delsudo` | Manage sudo users | | `/playmode` | Who can use `/play` |
| `/addauth` / `/delauth` | Manage auth users | | `/adminmode` | Who can use admin cmds |
| `/stats` | Bot statistics | | `/cleanmode` | Auto-clean settings |
| `/broadcast` | Message all active chats | | `/maintenance` | Global maintenance toggle |
| `/reload` | Reload admin cache & VC | | `/lang` | Change chat language |

---

## 🏗️ Project Structure

```text
YukkiMusic/
├── cmd/app/            # 🚪 Entry point
├── config/             # ⚙️  Env loader + cookies setup
├── internal/
│   ├── core/           # 🎛  Bot init, rooms & playback engine
│   ├── database/       # 🗄  MongoDB schemas & operations
│   ├── locales/        # 🌍  YAML translations
│   ├── modules/        # 🧩  Command handlers
│   ├── platforms/      # 🎼  Source providers (YouTube, etc.)
│   └── utils/          # 🔧  Helpers
├── docker-compose.yml  # 🐳 Container deployment
└── go.mod              # 📦 Dependencies
```

---

## 🤝 Contributing & Support

- 🐛 **Bugs / ideas** → [GitHub Issues](https://github.com/TheTeamVivek/YukkiMusic/issues)
- 💬 **Support** → [@TheTeamVk](https://t.me/TheTeamVk)
- 📢 **Updates** → [@TheTeamVivek](https://t.me/TheTeamVivek)

---

<div align="center">

**Made with ❤️ by [Vivek Kumar](https://github.com/Vivekkumar-in)**

<i>Licensed under the GNU General Public License v3.0</i>

</div>
