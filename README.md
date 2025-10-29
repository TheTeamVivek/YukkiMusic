<h1 align="center">🎧 <b>YukkiMusic</b></h1>

<p align="center">
  <i>⚡ A blazing-fast, reliable, and feature-packed Telegram bot for streaming music in group voice chats — built with Go.</i>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.24-blue?style=for-the-badge&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-GNU%20GPLv3-blue.svg?style=for-the-badge" alt="License: GPLv3">
  <a href="https://goreportcard.com/report/github.com/TheTeamVivek/YukkiMusic">
    <img src="https://goreportcard.com/badge/github.com/TheTeamVivek/YukkiMusic?style=for-the-badge" alt="Go Report Card">
  </a>
</p>

---

## ✨ Features

- 🎶 **High-Quality Audio:** Enjoy crystal-clear and uninterrupted playback.  
- 🧠 **Smart Queue System:** Manage your playlist with ease — play, skip, or reorder.  
- ⏯️ **Full Playback Control:** Commands for pause, resume, skip, seek, and replay.  
- 🛡️ **Admin Tools:** Secure command access for group administrators.  
- ⚙️ **Fully Configurable:** Customize everything through environment variables.  
- 🪶 **Lightweight & Efficient:** Designed for performance even under heavy use.

> [!NOTE]  
> 🔸 **Video playback is not supported.**  
> 🔸 Only **YouTube** and **Telegram audio files** are supported.

---

## 🚀 Getting Started

### 🧩 Prerequisites

- 🐹 **Go:** Version `1.24` or higher  
- 🎧 **FFmpeg:** Required for audio processing

---

### 🖥️ VPS Deployment

1. **Clone the Repository:**

```
git clone https://github.com/TheTeamVivek/YukkiMusic.git
cd YukkiMusic
```

2. **Install FFmpeg:**
```
sudo apt update && sudo apt install ffmpeg -y
```

3. **Configure:**

```
cp sample.env .env
nano .env
```

   Fill in the configuration variables as shown below.

4. **Install Dependencies & Run:**

```shell

go mod tidy
bash setup_ntgcalls.sh
go build -o app ./cmd/app
./app
```

---

### ☁️ Heroku Deployment

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/TheTeamVivek/YukkiMusic)

1. Click the **“Deploy to Heroku”** button above.  
2. Fill in all required **environment variables**.  
3. Hit **“Deploy app”** — and you’re live! 🚀

---

## ⚙️ Configuration

All settings are managed using a `.env` file or environment variables.

### 🔐 Required Variables

| Variable | Description |
|:----------|:-------------|
| `API_ID` | Telegram API ID — get it from [my.telegram.org](https://my.telegram.org). |
| `API_HASH` | Telegram API Hash — get it from [my.telegram.org](https://my.telegram.org). |
| `TOKEN` | Bot token from [@BotFather](https://t.me/BotFather). |
| `MONGO_DB_URI` | MongoDB connection string. |
| `STRING_SESSION` | Pyrogram session string for the assistant client. |

---

### ⚙️ Optional Variables

| Variable | Description | Default |
|:----------|:-------------|:----------|
| `FALLEN_API_KEY` | API key for the [Fallen API](https://tgmusic.fallenapi.fun/) (YouTube downloader). |  |
| `FALLEN_API_URL` | API URL for the [Fallen API](https://tgmusic.fallenapi.fun/). |  |
| `OWNER_ID` | User ID of the bot owner. |  |
| `LOGGER_ID` | Chat ID for logging errors and events. |  |
| `DURATION_LIMIT` | Maximum track duration in seconds. | `4200` (70 minutes) |
| `QUEUE_LIMIT` | Maximum queue size per chat. | `7` |
| `START_IMG_URL` | Start image URL for `/start` message. | [Default Image](https://raw.githubusercontent.com/Vivekkumar-IN/assets/master/images.png) |
| `SUPPORT_CHAT` | Support group link. | [@TheTeamVk](https://t.me/TheTeamVk) |
| `SUPPORT_CHANNEL` | Update channel link. | [@TheTeamVivek](https://t.me/TheTeamVivek) |
| `COOKIES_LINK` | `batbin.me` link to `yt-dlp` cookies file. |  |
| `SET_CMDS` | Set bot commands automatically on startup. | `false` |
| `MAX_AUTH_USERS` | Max number of authorized users per chat. | `25` |

---

## 💬 Commands

Type `/help` in your bot’s chat to view the complete list of available commands.

---

## 🤝 Contributing

Contributions are **welcome and appreciated**!  
- 🍴 Fork the repo  
- ✨ Make your changes  
- 💌 Submit a pull request  

You can also open an [issue](https://github.com/TheTeamVivek/YukkiMusic/issues/new) if you find bugs or have feature requests.

---

## ❤️ Support

💬 **Telegram:** [@TheTeamVk](https://t.me/TheTeamVk)  
📂 **GitHub Issues:** [Report a Problem](https://github.com/TheTeamVivek/YukkiMusic/issues/new)

---

## 📜 License

🧾 This project is licensed under the **GNU GPLv3 License** — see the [LICENSE](LICENSE) file for details.
