<h1 align="center">🎧 <b>YukkiMusic</b></h1>

<p align="center">
  <i>⚡ A blazing-fast, reliable, and feature-packed Telegram bot for streaming music in group voice chats — built with Go.</i>
</p>
<p align="center">

  <a href="https://go.dev/">
    <img src="https://img.shields.io/badge/Go-1.24-00ADD8?style=for-the-badge&logo=go&labelColor=000000&logoColor=white" alt="Go Version">
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

- 🎶 **High-Quality Audio:** Enjoy crystal-clear and uninterrupted playback.  
- 🧠 **Smart Queue System:** Manage your playlist with ease — play, skip, or reorder.  
- ⏯️ **Full Playback Control:** Commands for pause, resume, skip, seek, and replay.  
- 🛡️ **Admin Tools:** Secure command access for group administrators.  
- ⚙️ **Fully Configurable:** Customize everything through environment variables.  
- 🪶 **Lightweight & Efficient:** Designed for performance even under heavy use.

> [!NOTE]  
> 🔸 **Video playback is not supported.**  
> 🔸 Only **YouTube** and **Telegram audio files** are supported.

> [!NOTE]  
> 💡 To set up **music downloading** from YouTube, see the [Configuration](#-setting-up-youtube-downloads) section below — it explains how to use the **cookies** or **API** for downloads.

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
| `FALLEN_API_KEY` | API key for the [Fallen API](https://tgmusic.fallenapi.fun/) (YouTube downloader). You can get one from [@FallenApiBot](https://t.me/FallenApiBot). If you are using cookies, you can leave this empty. | *(empty)* |
| `FALLEN_API_URL` | Base URL for the [Fallen API](https://tgmusic.fallenapi.fun/). For most users, the default should work fine. | `https://tgmusic.fallenapi.fun` |
| `OWNER_ID` | User ID of the bot owner. |  |
| `LOGGER_ID` | Chat ID for logging errors and events. | *(empty)* |
| `DURATION_LIMIT` | Maximum track duration in seconds. | `4200` (70 minutes) |
| `QUEUE_LIMIT` | Maximum queue size per chat. | `7` |
| `START_IMG_URL` | Start image URL for `/start` message. | [Default Image](https://raw.githubusercontent.com/Vivekkumar-IN/assets/master/images.png) |
| `SUPPORT_CHAT` | Support group link. | [@TheTeamVk](https://t.me/TheTeamVk) |
| `SUPPORT_CHANNEL` | Update channel link. | [@TheTeamVivek](https://t.me/TheTeamVivek) |
| `COOKIES_LINK` | The [batbin.me](https://batbin.me) link where you pasted your `yt-dlp` cookies file. If you are using the Fallen API, you can leave this empty. You can also skip this if you manually place your cookies `.txt` file in `internal/cookies/`. | *(empty)* |
| `SET_CMDS` | Set [bot commands](https://raw.githubusercontent.com/Vivekkumar-IN/assets/refs/heads/master/bot_commands.png) automatically on startup. | `false` |
| `MAX_AUTH_USERS` | Max number of authorized users per chat. | `25` |

---
## 💬 Commands

Type `/help` in your bot’s chat to view the complete list of available commands.

---

## 🎧 Setting Up YouTube Downloads

YukkiMusic supports multiple methods to handle **YouTube downloads**.  
You can use any **one** of the following approaches depending on your setup.

---

### 🍪 1. Using Local Cookies Files

If you have your own YouTube cookies files:

- Place one or more `.txt` files inside:  
```
internal/cookies/
```

- Each file should follow the format:  
```
internal/cookies/<filename>.txt
```

- The bot will automatically detect and randomly use a cookie file from this directory at runtime.

> 💡 You can store multiple cookie files to reduce rate-limiting.

---

### 🌐 2. Using a Batbin Link (`COOKIES_LINK`)

If you prefer to host your cookies online:

1. Go to [batbin.me](https://batbin.me).  
2. Paste your full cookies content there and save.  
3. Copy the resulting link (for example, `https://batbin.me/abcd1234`).  
4. Add it in your variables or in `.env` file like this:  
```
COOKIES_LINK=https://batbin.me/abcd1234
```

> ⚙️ The bot will automatically fetch and save the cookies from your Batbin link into the `internal/cookies/` folder during startup.

---

### ⚡ 3. Using [Fallen API](https://tgmusic.fallenapi.fun/)

The simplest and most reliable method for most users.  
The **Fallen API** handles YouTube extraction and downloading on the server side — no cookies required.

- Get your API key from [@FallenApiBot](https://t.me/FallenApiBot).  
- In your `.env` file:  
```
FALLEN_API_KEY=your_api_key_here  
FALLEN_API_URL=https://tgmusic.fallenapi.fun
```

- If you don’t have a key, you can leave it empty.

> 💡 Recommended for users who don’t want to manage cookies manually.

---

### 🧩 4. Custom API or Advanced Integration

If you have your own API endpoint or downloader implementation,  
contact us at [@TheTeamVk](https://t.me/TheTeamVk) —  
we’ll provide ready-to-use **code templates** that you can integrate directly for your setup.

---
> ✅ **Summary:**  
> - Use **cookies** if you don’t want to pay for an API.  
> - Or contact us for a **custom solution** if you want to use your own API.
---

## 🧱 Platform System

YukkiMusic is powered by a **modular Platform System** — a flexible framework that allows it to fetch and download music from multiple sources like **YouTube**, **Telegram**, and more.  

Each platform works independently but connects seamlessly through a **priority-based registry**, ensuring the bot always picks the most efficient source for every query. ⚙️  

📖 **Learn More:**  
➡️ [📘 YukkiMusic Platform System](../internal/platforms/README.md)

> 💡 The Platform System is perfect for developers who want to add **custom APIs, new music sources, or modify fetching logic** without touching the bot’s core.
---

## 📌 To-Do

- [x] Add multi-platform download fallback  
      (if one platform fails, automatically try the next)

- [ ] Improve auth command checks  
      (simplify user extraction, bot checks, and errors)

- [ ] Add automatic rtmp check in CharState && iml multi lang
- [ ] Assitant restricted to send message but still able to join vc
- [ ] Broadcast command

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

🧾 This project is licensed under the **GNU GPLv3 License** — see the [LICENSE](../LICENSE) file for details.
