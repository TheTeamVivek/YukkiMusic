<div align="center">

# ⚙️ Environment Variables

**Configure YukkiMusic using a `.env` file in the project root**

[![dotenv](https://img.shields.io/badge/.env-optional-yellow?style=flat-square)](https://github.com/motdotla/dotenv)
[![restart](https://img.shields.io/badge/restart-required-red?style=flat-square)](#-note)

Variable names are **case-insensitive** and **underscores are optional** —
`API_ID`, `api_id`, and `APIID` all work the same.

</div>

---

## 🔴 Required

> These **must** be set or the bot won't start.

| Variable | Description |
| :--- | :--- |
| `API_ID` | Telegram API ID from [my.telegram.org](https://my.telegram.org) |
| `API_HASH` | Telegram API Hash from [my.telegram.org](https://my.telegram.org) |
| `BOT_TOKEN` | Bot token from [@BotFather](https://t.me/BotFather) · alias `TOKEN` |
| `MONGO_DB_URI` | MongoDB connection string (e.g. `mongodb+srv://...`) |
| `STRING_SESSIONS` | Assistant session strings (space/comma/semicolon separated) · alias `STRING_SESSION` |

---

## 👑 Ownership

| Variable | Default | Description |
| :--- | :--- | :--- |
| `OWNER_ID` | `0` | Telegram User ID of the bot owner — full admin access |

---

## 🤖 Bot Behavior

| Variable | Default | Description |
| :--- | :--- | :--- |
| `SESSION_TYPE` | `pyrogram` | Session type for assistant accounts |
| `DB_NAME` | `Yukki` | MongoDB database name |
| `DEFAULT_LANG` | `en` | Default language (`en`, `hi`, `ar`, `tr`) |
| `DURATION_LIMIT` | `4200` | Max track duration in seconds (70 min) |
| `QUEUE_LIMIT` | `24` | Max tracks in queue per chat |
| `MAX_AUTH_USERS` | `25` | Max non-admin users allowed to control playback |
| `LEAVE_ON_DEMOTED` | `false` | Leave the group if admin rights are removed |
| `SET_CMDS` | `false` | Set bot commands in Telegram UI on startup |
| `DISABLE_COLOUR` | `false` | Disable coloured inline buttons |
| `LOGGER_ID` | `0` | Chat ID for logs & error reports · alias `LOG_GROUP_ID` |

---

## 🎵 External APIs <sub><i>optional</i></sub>

| Variable | Default | Description |
| :--- | :--- | :--- |
| `FALLEN_API_KEY` | `""` | API key for [Fallen API](https://beta.fallenapi.fun) YouTube downloader |
| `FALLEN_API_URL` | `https://beta.fallenapi.fun` | Base URL for Fallen API |

---

## 🎨 Customization

| Variable | Default | Description |
| :--- | :--- | :--- |
| `START_IMAGES` | `https://files.catbox.moe/qri2mb.jpg` | Random `/start` image URLs · alias `START_IMG_URL` |
| `PING_IMG_URL` | `https://telegra.ph/file/91533956c91d0fd7c9f20.jpg` | Image shown in `/ping` response |
| `SUPPORT_CHAT` | `https://t.me/TheTeamVk` | Full URL to the support group |
| `SUPPORT_CHANNEL` | `https://t.me/TheTeamVivek` | Full URL to the announcement channel |
| `COOKIES_LINK` | `""` | URL(s) to a `yt-dlp` cookies file → [setup guide](cookies/README.md) |

---

## 🖥️ System

| Variable | Default | Description |
| :--- | :--- | :--- |
| `PORT` | `8000` | Port for the debug pprof server |

---

> [!NOTE]
> Any change to `.env` or environment variables requires a **bot restart** to take effect.

<div align="center">

→ Next: [🍪 Cookies Setup](cookies/README.md) · [🌍 Languages](../internal/locales/README.md)

</div>
