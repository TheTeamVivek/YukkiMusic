# Overview

[`YukkiMusic`](https://github.com/TheTeamVivek/YukkiMusic) is a feature-rich, fast, and powerful Telegram music bot that streams music in group voice chats.

## Key Features

- Supports **Spotify**, **YouTube**, **SoundCloud**, **Apple Music**, and more.
- Group call streaming powered by **PyTgCalls**.
- Admin control panel with commands like pause, resume, skip, seek, stop and much more.
- Queue system, lyrics, ping, and speedtest.
- Auto cleaner and auto-leave features.
- Multi-language support via `.yml` translation files.
- Heroku and vps compatible for easy deployment.

## Architecture

- `YukkiMusic/`: Main bot module
- `core/`: Core backend (call, filters, mongo, request, etc.)
- `plugins/`: Bot features, commands grouped by category
- `utils/`: Helper utilities, decorators, formatting, inline responses
- `config/`: Bot configuration and secrets
- `strings/`: Language translations
- `assets/`: Images used in inline and other features
- `app.py`: Entrypoint when not running with `__main__`

## Repository Info

- **GitHub**: [`TheTeamVivek/YukkiMusic`](https://github.com/TheTeamVivek/YukkiMusic)
- **License**: [`MIT LICENSE`](https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE)