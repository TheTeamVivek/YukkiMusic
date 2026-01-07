# ⚙️ Configuration Guide

This document provides a comprehensive guide to all configuration variables used in YukkiMusic. All variables are loaded from environment variables or a `.env` file.

---

## 📋 Table of Contents

- [Required Variables](#-required-variables)
- [Optional Variables](#-optional-variables)
- [Advanced Configuration](#-advanced-configuration)
- [How Variables Are Loaded](#-how-variables-are-loaded)

---

## 🔴 Required Variables

These variables **must** be set for the bot to function properly.

### `API_ID`
- **Type:** Integer (int32)
- **Description:** Your Telegram API ID obtained from [my.telegram.org](https://my.telegram.org).
- **Example:** `12345678`
- **How to get:**
  1. Visit [my.telegram.org](https://my.telegram.org)
  2. Log in with your phone number
  3. Go to "API Development Tools"
  4. Create an application and copy the **API ID**

### `API_HASH`
- **Type:** String
- **Description:** Your Telegram API Hash obtained from [my.telegram.org](https://my.telegram.org).
- **Example:** `abcdef1234567890abcdef1234567890`
- **How to get:** Follow the same steps as `API_ID` and copy the **API Hash**

### `TOKEN` (or `BOT_TOKEN`)
- **Type:** String
- **Description:** Your Telegram Bot Token obtained from [@BotFather](https://t.me/BotFather).
- **Example:** `1234567890:ABCdefGHIjklMNOpqrsTUVwxyz`
- **How to get:**
  1. Message [@BotFather](https://t.me/BotFather) on Telegram
  2. Send `/newbot` and follow the instructions
  3. Copy the token provided

### `MONGO_DB_URI`
- **Type:** String
- **Description:** MongoDB connection string for database storage.
- **Example:** `mongodb://localhost:27017/YukkiMusic`
- **Format:** `mongodb://[username:password@]host[:port]/database[?options]`
- **Free Options:** [MongoDB Atlas](https://www.mongodb.com/cloud/atlas) (Free tier available)

### `STRING_SESSIONS` (or `STRING_SESSION`)
- **Type:** String (space/comma/semicolon separated)
- **Description:** Pyrogram/Telethon/Gogram session strings for assistant accounts.
- **Example:** `session1 session2 session3`
- **Multiple Sessions:** You can provide multiple session strings separated by spaces, commas, or semicolons. The bot will load all of them as assistants.
- **How to generate:**
  - **Pyrogram:** Use [String Session Generator](https://telegram.tools/session-string-generator#pyrogram,user) or [@StringFatherBot](t.me/StringFatherBot) in Telegram.
  - **Telethon:** Use [Telethon String Session](https://telegram.tools/session-string-generator#telethon,user) or [@StringFatherBot](t.me/StringFatherBot) in Telegram.
  - **Gogram:** Use [gogram Session Generator.](https://session.gogram.fun/)

### `SESSION_TYPE`
- **Type:** String
- **Description:** The type of session string you're using.
- **Options:** `pyrogram`, `telethon`, `gogram`
- **Default:** `pyrogram`
- **Example:** `SESSION_TYPE=pyrogram`

---

## 🟢 Optional Variables

These variables are optional but recommended for enhanced functionality.

### Bot Ownership & Administration

#### `OWNER_ID`
- **Type:** Integer (int64)
- **Description:** Telegram User ID of the bot owner (you).
- **Example:** `123456789`
- **How to get:**
  1. Message [@userinfobot](https://t.me/userinfobot) on Telegram
  2. Copy your User ID
- **Permissions:** Full access to all bot commands and features.

#### `LOGGER_ID`
- **Type:** Integer (int64)
- **Description:** Chat ID where the bot sends error logs and important events.
- **Example:** `-1001234567890` (for groups/channels)
- **How to get:**
  1. Add [@MissRose_bot](https://t.me/MissRose_bot) to your group/channel
  2. Send `/id` to get the chat ID
- **Note:** Logs include errors, bug reports, and system events.

---

### Bot Limits & Restrictions

#### `DURATION_LIMIT`
- **Type:** Integer (seconds)
- **Description:** Maximum duration (in seconds) for tracks that can be played.
- **Default:** `4200` (70 minutes)
- **Example:** `3600` (60 minutes)
- **Range:** Any positive integer
- **Purpose:** Prevents users from queuing extremely long audio files.

#### `QUEUE_LIMIT`
- **Type:** Integer
- **Description:** Maximum number of tracks allowed in the queue per chat.
- **Default:** `7`
- **Example:** `10`
- **Range:** Any positive integer
- **Purpose:** Prevents queue spam and manages server resources.

#### `MAX_AUTH_USERS`
- **Type:** Integer
- **Description:** Maximum number of authorized users (non-admin users with playback control) per chat.
- **Default:** `25`
- **Example:** `50`
- **Range:** Any positive integer
- **Purpose:** Limits who can control playback in groups.

---

### Bot Behavior

#### `LEAVE_ON_DEMOTED`
- **Type:** Boolean
- **Description:** Whether the bot should automatically leave a group when demoted from admin.
- **Default:** `false`
- **Options:** `true`, `false`, `yes`, `no`, `1`, `0`, `enable`, `disable`
- **Example:** `LEAVE_ON_DEMOTED=true`
- **Use Case:** Useful if you want the bot to automatically leave groups where it loses admin rights.

#### `SET_CMDS`
- **Type:** Boolean
- **Description:** Automatically set bot commands in Telegram's UI on startup.
- **Default:** `false`
- **Options:** `true`, `false`, `yes`, `no`, `1`, `0`, `enable`, `disable`
- **Example:** `SET_CMDS=true`
- **Note:** Commands will be visible in the bot's menu button.

---

### Localization

#### `DEFAULT_LANG`
- **Type:** String
- **Description:** Default language for the bot's responses.
- **Default:** `en`
- **Available Languages:**
  - `en` - English (🇺🇸)
  - `hi` - हिन्दी (🇮🇳)
  - *(More languages can be added by creating YAML files in `internal/locales/`)*
- **Example:** `DEFAULT_LANG=hi`

---

### Customization

#### `START_IMG_URL`
- **Type:** String (URL)
- **Description:** Image URL displayed in the `/start` command.
- **Default:** `https://raw.githubusercontent.com/Vivekkumar-IN/assets/master/images.png`
- **Format:** Must be a direct image URL (HTTPS recommended)
- 
#### `PING_IMG_URL`
- **Type:** String (URL)
- **Description:** Image URL displayed in the `/ping` command.
- **Default:** `https://telegra.ph/file/91533956c91d0fd7c9f20.jpg`
- **Format:** Must be a direct image URL (HTTPS recommended)

#### `SUPPORT_CHAT`
- **Type:** String (URL or username)
- **Description:** Link to your bot's support group.
- **Default:** `https://t.me/TheTeamVk`
- **Format:** `https://t.me/YourGroup`
- **Example:** `SUPPORT_CHAT=https://t.me/MyBotSupport`

#### `SUPPORT_CHANNEL`
- **Type:** String (URL or username)
- **Description:** Link to your bot's updates/announcement channel.
- **Default:** `https://t.me/TheTeamVivek`
- **Format:** `https://t.me/YourChannel`
- **Example:** `SUPPORT_CHANNEL=https://t.me/MyBotUpdates`

---

### YouTube Download Methods

YukkiMusic supports multiple methods for downloading YouTube tracks. You can use **one or more** of these methods:

#### `COOKIES_LINK`
- **Type:** String (URL or space-separated URLs)
- **Description:** [Batbin.me](https://batbin.me) link(s) containing your `yt-dlp` cookies file(s).
- **Default:** *(empty)*
- **Format:** `https://batbin.me/paste_id` or multiple URLs separated by spaces
- **Multiple Cookies Example:** `https://batbin.me/id1 https://batbin.me/id2`
- **How to use:**
  1. Export your YouTube cookies using a browser extension
  2. Paste the cookies content on [batbin.me](https://batbin.me)
  3. Copy the resulting URL
  4. Set it as `COOKIES_LINK`
- **Alternative:** You can also manually place cookie `.txt` files in the `internal/cookies/` directory
- **Note:** Using cookies is free but may have rate limits. Consider using the Fallen API for better reliability.

---

#### `FALLEN_API_KEY`
- **Type:** String
- **Description:** API key for the [Fallen API](https://beta.fallenapi.fun/) YouTube downloader service.
- **Default:** *(empty)*
- **How to get:** Message [@FallenApiBot](https://t.me/FallenApiBot) on Telegram
- **Note:** If using cookies method, you can leave this empty.

#### `FALLEN_API_URL`
- **Type:** String
- **Description:** Base URL for the Fallen API service.
- **Default:** `https://beta.fallenapi.fun`

---
#### `YOUTUBIFY_API_KEY`
- **Type:** String
- **Description:** API key for the [Youtubify API](https://youtubify.me) YouTube downloader service.
- **Default:** *(empty)*
- **How to get:** Obtain it from [Youtubify site](https://youtubify.me).
- **Note:** If using cookies method, you can leave this empty.
  
### `YOUTUBIFY_API_URL`
- **Type:** String
- **Description:** Base URL for the **Youtubify API** YouTube downloader service.
- **Default:** `https://youtubify.me`

---

## 🔧 Advanced Configuration

### Environment Variable Loading

The config loader supports **multiple naming variants** for each variable. For example, `API_ID` can also be specified as:
- `api_id` (lowercase)
- `APIID` (no underscore)
- `ApiId` (title case)

This flexibility helps when migrating from other bots or when you prefer different naming conventions.

### Multiple Values

Some variables support **multiple values** separated by:
- **Spaces:** `value1 value2 value3`
- **Commas:** `value1,value2,value3`
- **Semicolons:** `value1;value2;value3`

**Variables supporting multiple values:**
- `STRING_SESSIONS`
- `COOKIES_LINK`

---

## 📝 Example `.env` File

```bash
# ==========================================
# REQUIRED VARIABLES
# ==========================================
API_ID=12345678
API_HASH=abcdef1234567890abcdef1234567890
TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
MONGO_DB_URI=mongodb://localhost:27017/YukkiMusic
STRING_SESSIONS=session_string_1 session_string_2
SESSION_TYPE=pyrogram

# ==========================================
# OPTIONAL - OWNERSHIP
# ==========================================
OWNER_ID=123456789
LOGGER_ID=-1001234567890

# ==========================================
# OPTIONAL - LIMITS & RESTRICTIONS
# ==========================================
DURATION_LIMIT=4200
QUEUE_LIMIT=7
MAX_AUTH_USERS=25

# ==========================================
# OPTIONAL - BOT BEHAVIOR
# ==========================================
LEAVE_ON_DEMOTED=false
SET_CMDS=true
DEFAULT_LANG=en

# ==========================================
# OPTIONAL - CUSTOMIZATION
# ==========================================
START_IMG_URL=https://example.com/start-image.png
PING_IMG_URL=https://example.com/ping-image.png
SUPPORT_CHAT=https://t.me/YourSupportGroup
SUPPORT_CHANNEL=https://t.me/YourChannel

# ==========================================
# OPTIONAL - YOUTUBE DOWNLOADS
# ==========================================
COOKIES_LINK=https://batbin.me/paste_id1 https://batbin.me/paste_id2

# Fallen API Downloader
FALLEN_API_KEY=your_api_key_here
FALLEN_API_URL=https://beta.fallenapi.fun

# Youtubify API Downloader
YOUTUBIFY_API_URL=https://youtubify.me
YOUTUBIFY_API_KEY=your_api_key_here
```

---

## 🔍 How Variables Are Loaded

The configuration system follows this priority order:

1. **Environment Variables** (highest priority)
2. **`.env` File** in project root
3. **Default Values** (lowest priority)

### Variable Naming Flexibility

The loader automatically tries multiple variants of each variable name:
```
API_ID → api_id → apiid → Api Id
```

This means you can use any of these naming styles, and the bot will find it.

---

## ⚠️ Common Issues

### Bot Not Starting

**Problem:** Bot fails to start with "TOKEN is required"

**Solution:**
- Check if `TOKEN` or `BOT_TOKEN` is set in `.env`
- Ensure there are no extra spaces or quotes
- Verify the token is valid by checking with [@BotFather](https://t.me/BotFather)

---

### MongoDB Connection Failed

**Problem:** "Failed to connect to MongoDB"

**Solution:**
- Verify `MONGO_DB_URI` format is correct
- Check if MongoDB service is running (for local installations)
- For MongoDB Atlas, ensure your IP is whitelisted
- Test connection string with MongoDB Compass

---

### Session String Invalid

**Problem:** "Invalid session string" or assistant fails to start

**Solution:**
- Ensure `SESSION_TYPE` matches your session string format
- Generate a fresh session string
- Check that session string doesn't contain extra spaces or newlines
- Verify API_ID and API_HASH match the ones used to generate the session

---

### YouTube Download Fails

**Problem:** Songs won't download from YouTube

**Solution:**
1. **Using Cookies:**
   - Ensure cookies are up to date (regenerate every few days)
   - Verify `COOKIES_LINK` URLs are accessible
   - Check `internal/cookies/` directory has `.txt` files

2. **Using Fallen API:**
   - Verify `FALLEN_API_KEY` is set correctly
   - Check API quota hasn't been exceeded
   - Ensure `FALLEN_API_URL` is reachable

3. **Using Youtubify API:**
   - Verify `YOUTUBIFY_API_URL` is reachable (default: `https://youtubify.me`)
   - Check the Youtubify service status at: https://youtubify.me
   - Try switching between Cookies, Fallen API, and Youtubify as fallback methods
   
---

## 🎯 Best Practices

1. **Security:**
   - Never commit `.env` to version control
   - Keep API keys and tokens secret
   - Regularly rotate sensitive credentials

2. **Performance:**
   - Use multiple assistant sessions for better load balancing
   - Set appropriate `QUEUE_LIMIT` and `DURATION_LIMIT` for your server
   - Enable `LOGGER_ID` to monitor errors

3. **User Experience:**
   - Set `SET_CMDS=true` for better command discoverability
   - Customize `START_IMG_URL` and `PING_IMG_URL` with your branding
   - Configure `SUPPORT_CHAT` and `SUPPORT_CHANNEL` for user assistance

4. **Reliability:**
   - Use both cookies and API for YouTube downloads as fallbacks
   - Set `LEAVE_ON_DEMOTED=true` to automatically clean up inactive groups
   - Configure `MAX_AUTH_USERS` to prevent abuse

---

## 🆘 Need Help?

If you're still having issues:

1. Check the main [README](../../.github/README.md) for general setup instructions
2. Join the [Support Chat](https://t.me/TheTeamVk) for community help
3. Report bugs using the `/bug` command in the bot
4. Open an issue on [GitHub](https://github.com/TheTeamVivek/YukkiMusic/issues)

---

## 🔗 Related Documentation

- [Main README](../../.github/README.md) - General setup and deployment guide
- [Platform System](../platforms/README.md)- Understanding music sources

---

**📌 Note:** This configuration guide is maintained for YukkiMusic v2.0+. Some variables may differ in older versions.
