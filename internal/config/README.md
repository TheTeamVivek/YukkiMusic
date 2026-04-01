# ⚙️ Configuration Guide

This document describes all configuration variables used in YukkiMusic. These can be set as environment variables or in a `.env` file in the project root.

## 🚀 Flexible Loading

The configuration loader is designed to be forgiving with naming. For any variable listed below, you can use:
- **UPPER_CASE** (e.g., `API_ID`)
- **lower_case** (e.g., `api_id`)
- **No Underscores** (e.g., `APIID`)

## 🔴 Required Variables

These must be set for the bot to start.

| Variable | Type | Description |
| :--- | :--- | :--- |
| `API_ID` | `int32` | Your Telegram API ID from [my.telegram.org](https://my.telegram.org). |
| `API_HASH` | `string` | Your Telegram API Hash from [my.telegram.org](https://my.telegram.org). |
| `TOKEN` | `string` | Your Telegram Bot Token from [@BotFather](https://t.me/BotFather). Also accepts `BOT_TOKEN`. |
| `LOGGER_ID` | `int64` | Chat ID where the bot will send logs and error reports. Also accepts `LOG_GROUP_ID`.|
| `MONGO_DB_URI` | `string` | MongoDB connection string (e.g., `mongodb+srv://...`). |
| `STRING_SESSIONS`| `[]string`| Space, comma, or semicolon separated session strings for assistant accounts. Also accepts `STRING_SESSION`. |

## 🟢 Assistant Configuration

| Variable | Default | Options | Description |
| :--- | :--- | :--- | :--- |
| `SESSION_TYPE` | `pyrogram` | `pyrogram`, `telethon`, `gogram` | The library used to generate your `STRING_SESSIONS`. This **must** match the session format. |

## 👑 Ownership

| Variable | Default | Description |
| :--- | :--- | :--- |
| `OWNER_ID` | `0` | The Telegram User ID of the bot owner. Grants full administrative access. |

## 🎵 External APIs (Optional)

| Variable | Default | Description |
| :--- | :--- | :--- |
| `SPOTIFY_CLIENT_ID` | `None` | Spotify API Client ID for metadata resolution. |
| `SPOTIFY_CLIENT_SECRET` | `None` | Spotify API Client Secret for metadata resolution. |
| `FALLEN_API_KEY` | `""` | API Key for [Fallen API](https://beta.fallenapi.fun) YouTube downloader. |
| `FALLEN_API_URL` | `https://beta.fallenapi.fun` | Base URL for Fallen API. |

## 🛠️ Bot Behavior & Limits

| Variable | Default | Description |
| :--- | :--- | :--- |
| `DEFAULT_LANG` | `en` | Default language for bot responses (see `internal/locales/`). |
| `DURATION_LIMIT` | `4200` | Maximum track duration in seconds (70 minutes). |
| `QUEUE_LIMIT` | `24` | Maximum number of tracks in queue per chat. |
| `MAX_AUTH_USERS` | `25` | Max number of non-admin users allowed to control playback. |
| `LEAVE_ON_DEMOTED`| `false` | If `true`, the bot leaves the group if its admin rights are removed. |
| `SET_CMDS` | `false` | Automatically set bot commands in Telegram UI on startup. |
| `COOKIES_LINK` | `""` | URL to a `yt-dlp` cookies file (e.g., via batbin). |

## 🎨 Customization

| Variable | Default | Description |
| :--- | :--- | :--- |
| `START_IMG_URL` | `https://...` | Image URL displayed in the `/start` command response. |
| `PING_IMG_URL` | `https://...` | Image URL displayed in the `/ping` command response. |
| `SUPPORT_CHAT` | `https://t.me/TheTeamVk` | **Full URL** to the support group. |
| `SUPPORT_CHANNEL`| `https://t.me/TheTeamVivek` | **Full URL** to the announcement channel. |

## 🖥️ System Settings

| Variable | Default | Description |
| :--- | :--- | :--- |
| `PORT` | `8000` | Port for the internal debug pprof server. |
| `LOG_FILE` | `logs.txt` | Filename for system logs. |

---
**Note:** Changes to the `.env` file or environment variables require a bot restart to take effect.
