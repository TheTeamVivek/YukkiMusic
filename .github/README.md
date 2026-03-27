<h1 align="center">🎧 <b>YukkiMusic</b></h1>

<p align="center">
  <i>⚡ A blazing-fast, reliable, and feature-packed Telegram bot for streaming music in group voice chats — built with Go.</i>
</p>

<p align="center">
  <a href="https://github.com/TheTeamVivek/YukkiMusic/stargazers">
    <img src="https://img.shields.io/github/stars/TheTeamVivek/YukkiMusic?style=for-the-badge&label=Stars&color=FFD700&logo=github&logoColor=white" alt="GitHub Stars">
  </a>
  <a href="https://github.com/TheTeamVivek/YukkiMusic/fork">
    <img src="https://img.shields.io/github/forks/TheTeamVivek/YukkiMusic?style=for-the-badge&label=Forks&color=00C853&logo=github&logoColor=white" alt="GitHub Forks">
  </a>
  <a href="https://go.dev/">
    <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version">
  </a>
  <a href="https://github.com/TheTeamVivek/YukkiMusic/releases">
    <img src="https://img.shields.io/badge/Version-v3.2.0-FF9800?style=for-the-badge&logo=semver&logoColor=white" alt="Release Version">
  </a>
  <a href="https://t.me/TheTeamVk">
    <img src="https://img.shields.io/badge/Telegram-Support-26A69A?style=for-the-badge&logo=telegram&logoColor=white" alt="Support Group">
  </a>
</p>

<hr>

<h2>✨ Key Features</h2>
<ul>
  <li>🚀 <b>High Performance:</b> Optimized Go-based engine for ultra-low latency.</li>
  <li>🎼 <b>Smart Provider System:</b> Supports YouTube, Telegram Files, and custom APIs.</li>
  <li>🔄 <b>Intelligent Fallback:</b> Automatically switches sources if one fails.</li>
  <li>🌍 <b>Localization:</b> Robust multi-language support (YAML-based).</li>
  <li>🛠 <b>Admin Suite:</b> Speed control, seeking, looping, and advanced queueing.</li>
</ul>

<hr>

<h2>🚀 Quick Start</h2>

<h3>☁️ One-Click Deploy</h3>
<p>
  <a href="https://heroku.com/deploy?template=https://github.com/TheTeamVivek/YukkiMusic">
    <img src="https://www.herokucdn.com/deploy/button.svg" alt="Deploy to Heroku">
  </a>
</p>

<h3>💻 Manual Installation</h3>
<pre><code>
# 1. Clone & Enter
git clone https://github.com/TheTeamVivek/YukkiMusic.git && cd YukkiMusic
<br>
# 2. Setup Environment
bash install.sh
cp sample.env .env  # Edit your API_ID, TOKEN, etc.
<br>
# 3. Tidy & Run
go mod tidy
go run ./cmd/app
</code></pre>

<hr>

<h2>⚙️ Configuration</h2>
<p>All settings are managed via environment variables. See the <b><a href="../internal/config/README.md">Configuration Guide</a></b> for full details.</p>

<table>
  <thead>
    <tr>
      <th>Variable</th>
      <th>Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>API_ID</code> / <code>API_HASH</code></td>
      <td>Obtained from <a href="https://my.telegram.org">my.telegram.org</a></td>
    </tr>
    <tr>
      <td><code>TOKEN</code></td>
      <td>Bot token from <a href="https://t.me/BotFather">@BotFather</a></td>
    </tr>
    <tr>
      <td><code>LOGGER_ID</code></td>
      <td>Channel ID for bot logs (Must start with -100)</td>
    </tr>
    <tr>
      <td><code>MONGO_DB_URI</code></td>
      <td>MongoDB Atlas or local connection string</td>
    </tr>
    <tr>
      <td><code>STRING_SESSIONS</code></td>
      <td>Assistant account sessions (Pyrogram/Telethon/Gogram)</td>
    </tr>
  </tbody>
</table>

<hr>

<h2>🎮 Commands Reference</h2>

<h3>🎵 Playback & Queue</h3>
<ul>
  <li><code>/play &lt;query&gt;</code> — Play from YouTube/Telegram</li>
  <li><code>/fplay &lt;query&gt;</code> — Force play (skips current track)</li>
  <li><code>/pause</code> | <code>/resume</code> — Toggle playback state</li>
  <li><code>/queue</code> — View the upcoming list</li>
  <li><code>/seek &lt;seconds&gt;</code> — Jump to a specific time</li>
</ul>

<h3>🛡️ Admin & Owner</h3>
<ul>
  <li><code>/addsudo</code> | <code>/delsudo</code> — Manage authorized controllers</li>
  <li><code>/stats</code> — View system and database metrics</li>
  <li><code>/broadcast</code> — Send a message to all active chats</li>
  <li><code>/maintenance</code> — Toggle global maintenance mode</li>
</ul>

<hr>

<h2>🏗️ Project Structure</h2>
<pre><code>YukkiMusic/
├── cmd/app/            # Main entry point
├── config/             # Configuration & Environment loader
├── internal/
│   ├── core/           # Bot initialization & playback logic
│   ├── database/       # MongoDB schemas & operations
│   ├── locales/        # YAML-based translation files
│   ├── modules/        # Command handlers (play, skip, etc.)
│   ├── platforms/      # Source providers (YouTube, TG, etc.)
│   └── utils/          # Global helper functions
└── go.mod              # Dependency management</code></pre>

<hr>

<h2>🤝 Contributing & Support</h2>
<ul>
  <li><b>Issues:</b> Found a bug or have a suggestion? Open a <b><a href="https://github.com/TheTeamVivek/YukkiMusic/issues">GitHub Issue</a></b>.</li>
  <li><b>Support:</b> Join our Telegram community for help: <b><a href="https://t.me/TheTeamVk">@TheTeamVk</a></b>.</li>
  <li><b>Updates:</b> Follow <b><a href="https://t.me/TheTeamVivek">@TheTeamVivek</a></b> for news.</li>
</ul>

<hr>

<p align="center">
  <b>Made with ❤️ by <a href="https://github.com/Vivekkumar-in">Vivek Kumar</a></b><br>
  <i>Licensed under the GNU General Public License v3.0</i>
</p>
