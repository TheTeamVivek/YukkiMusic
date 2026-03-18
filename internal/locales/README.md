# 🌍 Localization System

This document explains the localization system used in YukkiMusic and how to add new languages.

## 📂 File Structure

The localization system is powered by YAML files in the `internal/locales/` directory. Each language has its own file, e.g., `en.yml` for English or `hi.yml` for Hindi.

- `internal/locales/*.yml`: Language files.
- `internal/locales/loader.go`: Logic for loading and retrieving translated strings.

## 🛠️ YAML Structure

Each language file is a flat map of keys to strings:

```yaml
name: "🇺🇸 English"

# Simple string
CLOSE_BTN: "Close"

# String with interpolation
active_chats_info: "📊 <b>Active Chats:</b>\n\nThere are <i>{count}</i> active chats across the bot system."

# Multiline string
assistant_restricted_warning: |
  ⚠️ <b>Assistant Restricted</b>

  The assistant {assistant} (ID: <code>{id}</code>) has been <b>banned</b>...
```

## 🏗️ How it Works

1. **Loading:** On startup, the `Load()` function in `loader.go` reads all `.yml` files using Go's `embed.FS` and unmarshals them into a `loadedLocales` map.
2. **Retrieving:** The `Get(lang, key, values Arg)` function is used to fetch a string. It defaults to the `DEFAULT_LANG` (from `config.go`) if the requested language or key is missing.
3. **Interpolation:** The `Arg` map (which is a `map[string]any`) is used for string interpolation. The system replaces `{key}` in the YAML string with the corresponding value from the map.

## ➕ Adding a New Language

To add support for a new language, follow these steps:

1. **Create the YAML file:** Add a new file in `internal/locales/` named with the language code (e.g., `es.yml` for Spanish).
2. **Copy the structure:** Start by copying the keys from `en.yml`.
3. **Add the `name` field:** The `name` field in the YAML should be the name of the language with its flag emoji (e.g., `name: "🇪🇸 Español"`).
4. **Translate the strings:** Translate all values into your language.
5. **Set the Default (Optional):** If you want to make this the bot's default language, update the `DEFAULT_LANG` in your environment or `.env` file.

## 💡 Developer Tip: Helper Functions

Instead of calling `locales.Get` directly, use the `F` or `FWithLang` helpers in `internal/modules/helpers.go`:

- `F(chatID, key, args...)`: Automatically detects the chat's language and retrieves the string.
- `FWithLang(lang, key, args...)`: Explicitly uses a language code to retrieve a string.

---
**Note:** All YAML files in `internal/locales/` are automatically embedded into the binary during compilation. You don't need to manually register new files in the code.
