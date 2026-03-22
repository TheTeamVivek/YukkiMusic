# 🎼 YukkiMusic Platform System

> **A modular, extensible, and priority-based framework for music source integration.**

---

## 📋 Table of Contents

- [Overview](#-overview)
- [How It Works](#%EF%B8%8F-how-it-works)
- [Available Platforms](#-available-platforms)
- [Priority System](#-priority-system)
- [Adding New Platforms](#-adding-new-platforms)
- [Models](#-models)
- [Credits](#-credits)

---

## 🌟 Overview

The **Platform System** is the engine driving YukkiMusic's music fetching and downloading. Each platform is a self-contained module designed for a specific source.

### Key Capabilities:
- ✅ **Validation**: Smartly determines if a query/URL belongs to the platform.
- ✅ **Metadata**: Fetches titles, durations, and high-quality artwork.
- ✅ **Download**: Handles local caching and efficient media retrieval.
- ✅ **Resilience**: Integrated fallback system ensures playback even if a primary source fails.

---

## ⚙️ How It Works

### 🔄 The Lifecycle
1. **User Request**: A query or URL is received.
2. **Platform Selection**: The registry iterates through platforms by **Priority** (Highest first).
3. **Capability Check**: The first platform that returns `true` for `CanGetTracks` or `CanSearch` is selected.
4. **Resolution**: Metadata is fetched. If it fails, the system tries the next available platform.
5. **Execution**: The media is downloaded and passed to the playback engine.

### 🛠 Internal Registry
Platforms are registered during `init()` with a unique `PlatformName` and a `Priority` integer. Higher integers take precedence.

---

## 📱 Available Platforms

### 📡 Telegram (Priority: 100)
Handles native Telegram media files (Audio, Video, Voice).
- **Features**: Fast direct streaming, no external API dependencies.
- **When to Use**: Direct `t.me` links or replies to files.

### 🎧 Spotify (Priority: 95)
Resolves Spotify metadata and falls back to YouTube for downloads.
- **Features**: Seamless support for Tracks, Playlists, Albums, and Artists.
- **Configuration**: Requires `SPOTIFY_CLIENT_ID` and `SPOTIFY_CLIENT_SECRET`.

### 🎥 YouTube (Priority: 90)
Powerful metadata resolver and search engine.
- **Features**: Playlist support, advanced search, and high-accuracy results.
- **Note**: Does not handle direct downloads; serves as a metadata source for other downloaders.

### 🎵 SoundCloud (Priority: 85)
Native support for SoundCloud tracks and playlists via `yt-dlp`.
- **Features**: High-quality audio extraction and cookie support.

### ⚡ Fallen API (Priority: 80)
Premium, high-speed API for YouTube audio downloads.
- **Features**: Stable performance and Telegram CDN integration.
- **Configuration**: Requires `FALLEN_API_URL` and `FALLEN_API_KEY`.

### 🔗 DirectStream (Priority: 65)
Fallback for direct media URLs (`.mp3`, `.mp4`, `.m3u8`, etc.).
- **Features**: Support for HLS/M3U8 streams and MPEG-DASH.
- **When to Use**: CDN links and live broadcast streams.

### 🧰 YT-DLP (Priority: 60)
The ultimate universal fallback downloader.
- **Features**: Supports 1000+ sites, local control, and smart cookie rotation.
- **Pros**: Free, extremely versatile, and reliable.

---

## 📊 Priority System

| Priority | Platform | Primary Purpose |
| :--- | :--- | :--- |
| **100** | **Telegram** | Direct app-native media |
| **95** | **Spotify** | Premium metadata matching |
| **90** | **YouTube** | Global search & metadata |
| **85** | **SoundCloud** | Native indie music support |
| **80** | **Fallen API** | Optimized audio downloads |
| **65** | **DirectStream** | Direct URLs & HLS streams |
| **60** | **YT-DLP** | Universal compatibility layer |

---

### 💡 Why Priority Matters

**High priority ensures the best tool for the job is used first.**

1. **Direct Stream URL**:
   - Checked by Telegram (100) → ❌
   - Checked by Spotify (95) → ❌
   - ...
   - Checked by **DirectStream (65)** → ✅ **Handled!**

2. **YouTube Link**:
   - Checked by Telegram (100) → ❌
   - ...
   - Checked by **YouTube (90)** → ✅ **Metadata Resolved!**
   - Download phase → Uses **Fallen API (80)** or **YT-DLP (60)**.

---

## 🧠 Adding New Platforms

Creating a new platform is straightforward. Follow this boilerplate to get started:

### 1. Create the File
`internal/platforms/myplatform.go`

### 2. Implementation Boilerplate
```go
package platforms

import (
    "context"
    "errors"
    "main/internal/core/models"
    "github.com/amarnathcjd/gogram/telegram"
)

type MyPlatform struct {
    name models.PlatformName
}

func init() {
    // Priority: Higher = Checked earlier
    Register(50, &MyPlatform{name: "MyPlatform"})
}

func (p *MyPlatform) Name() models.PlatformName { return p.name }

func (p *MyPlatform) CanGetTracks(query string) bool {
    return strings.Contains(query, "myservice.com")
}

func (p *MyPlatform) GetTracks(query string, video bool) ([]*models.Track, error) {
    // Logic to fetch metadata
    return nil, nil
}

func (p *MyPlatform) CanDownload(source models.PlatformName) bool {
    return source == p.name
}

func (p *MyPlatform) Download(ctx context.Context, t *models.Track, m *telegram.NewMessage) (string, error) {
    // Logic to download file
    return "", errors.New("not implemented")
}

func (p *MyPlatform) CanSearch() bool { return false }
func (p *MyPlatform) Search(q string, v bool) ([]*models.Track, error) { return nil, nil }
```

---

## 🔌 Core Models

```go
type Track struct {
    ID        string       // Unique track identifier
    Title     string       // Display title
    Duration  int          // Length in seconds
    Artwork   string       // Thumbnail URL or local path
    URL       string       // Original source URL
    Requester string       // User who added the track (HTML)
    Video     bool         // Toggle for video playback
    Source    PlatformName // Originating platform
}
```

---

## 🎯 Credits & Support

- **Core Logic**: Adapted from various open-source music bots.
- **Search**: YouTube scraping logic inspired by [TgMusicBot](https://github.com/AshokShau/TgMusicBot).
- **Support**: Join our [Support Group](https://t.me/TheTeamVk) for integration help.

---
**Happy Coding! 🎼**
