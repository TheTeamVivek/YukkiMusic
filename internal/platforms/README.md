# 🎼 YukkiMusic Platform System

> **A modular, extensible, and priority-based framework for music source integration.**

---

## 📋 Table of Contents

1. [Overview](#-overview)
2. [How It Works](#%EF%B8%8F-how-it-works)
3. [Available Platforms](#-available-platforms)
4. [Priority System](#-priority-system)
5. [Adding New Platforms](#-adding-new-platforms)
6. [Platform Interface](#-platform-interface)
7. [Troubleshooting](#-troubleshooting)

---

## 🌟 Overview

The **Platform System** is the heart of YukkiMusic's music fetching and downloading capabilities. Each platform is a self-contained module that:

✅ **Validates** if it can handle a given query  
✅ **Fetches** track metadata (title, duration, artwork)  
✅ **Downloads** the actual media file  
✅ **Handles errors** gracefully with fallbacks  

When a user requests a song, YukkiMusic:
1. Iterates through all registered platforms (by priority)
2. Checks if the first platform can handle the request
3. Uses the first valid platform
4. Falls back to next platform if fetch/download fails

---

## ⚙️ How It Works

### Registration Flow

```
Platform Registration (init())
         ↓
Priority-Based Registry
         ↓
User Requests Song
         ↓
Check Platforms (High → Low Priority)
         ↓
First Valid Platform Handles
         ↓
Fetch Metadata → Download → Play
         ↓
If Error → Try Next Platform
```

### Internal Mechanism

Each platform is stored in a **registry** with:
- `PlatformName` - Unique identifier
- `Priority` - Integer (higher = checked first)
- `Platform` - Implementation of the interface

When you call `GetOrderedPlatforms()`:
1. All platforms are sorted by priority (descending)
2. Returned in order of importance
3. Bot checks first valid one

---

## 📱 Available Platforms

### 1. **Telegram** (Priority: 100)
**Status**: ✅ Fully Supported

Handles direct Telegram audio/video files.

```
Input: Telegram link (t.me/channel/12345)
↓
Output: Streams audio/video directly from Telegram
```

**Features**:
- Download Telegram media files
- Support for voice messages, audio, video
- Auto-detect duration and metadata
- Fast streaming without extra processing

**When Used**:
- Direct Telegram links
- Reply to Telegram media
- Telegram document files

---

### 2. **Youtubify API** (Priority: 100)
**Status**: ✅ Requires API Key

Premium API for YouTube downloads with video support.

**Features**:
- Fast and stable downloads
- Video + Audio support
- API-based access
- High reliability

**Configuration**:
```bash
YOUTUBIFY_API_URL=https://youtubify.me
YOUTUBIFY_API_KEY=your_key_here
```

**Notes**: Paid service, requires valid API key

---

### 3. **Spotify** (Priority: 95)
**Status**: ✅ Fully Supported

Fetches Spotify metadata and downloads via YouTube fallback.

```
Input: Spotify track/playlist/album/artist URL
↓
Fetch Spotify metadata → Search YouTube → Download
```

**Features**:
- Track, playlist, album, artist support
- Automatic YouTube search for downloads
- High-quality metadata extraction
- Smart title matching

**Configuration**:
```bash
SPOTIFY_CLIENT_ID=your_client_id
SPOTIFY_CLIENT_SECRET=your_client_secret
```

**When Used**:
- Spotify track links
- Spotify playlist links
- Spotify album links
- Spotify artist links

---

### 4. **YouTube** (Priority: 90)
**Status**: ✅ Fully Supported

Fetches YouTube video metadata **only** (not download).

```
Input: YouTube URL or Search Query
↓
Output: Track metadata (title, duration, thumbnail)
         (Actual download via fallback platform)
```

**Features**:
- YouTube URL validation
- Playlist support
- Video search
- Web scraping for accurate data
- YTSearch fallback for reliability

**When Used**:
- YouTube links (youtube.com, youtu.be)
- Text search queries
- Playlist URLs

**Note**: YouTube platform **doesn't download**. Downloads handled by other platforms.

---
### 5. **JioSaavn** (Priority: 88)
**Status**: ✅ Fully Supported

Fetches and downloads tracks from JioSaavn (Indian music streaming service).

```
Input: JioSaavn track/album/playlist URL
↓
Fetch metadata → Decrypt download URL → Download 320kbps audio
```

**Features**:
- Track, album, and playlist support
- 320kbps audio quality downloads
- Automatic image quality enhancement (500x500)
- URL decryption using DES ECB cipher
- 2-hour metadata caching

**When Used**:
- JioSaavn song links (`jiosaavn.com/song/*/TOKEN`)
- JioSaavn album links (`jiosaavn.com/album/*/TOKEN`)
- JioSaavn playlist links (`jiosaavn.com/featured/*/TOKEN`)

**Example URLs**:
```
https://www.jiosaavn.com/song/believer/QwB2AzhoA3I
https://www.jiosaavn.com/album/evolve/Rq47dL29uZo_
https://www.jiosaavn.com/featured/its-indie-english/AMoxtXyKHoU_
```

**Notes**: 
- Audio only, no video support
- API logic based on [jiosaavn-api](https://github.com/sumitkolhe/jiosaavn-api/) by Sumit Kolhe (MIT License)
- Works with both `jiosaavn.com` and `saavn.com` domains

---

### 6. **SoundCloud** (Priority: 85)
**Status**: ✅ Fully Supported

Fetches and downloads SoundCloud tracks using yt-dlp.

**Features**:
- Track and playlist support
- Metadata extraction via yt-dlp
- Direct audio downloads
- Cookie-based authentication

**When Used**:
- SoundCloud track links
- SoundCloud playlist links

---

### 7. **Fallen API** (Priority: 80)
**Status**: ✅ Requires API Key

Premium API for YouTube downloads (audio only).

**Features**:
- Stable audio downloads
- API-based access
- Telegram CDN support

**Configuration**:
```bash
FALLEN_API_URL=https://beta.fallenapi.fun
FALLEN_API_KEY=your_key_here
```

**Notes**: Paid service, audio only, no video support

---

### 8. **DirectStream** (Priority: 65)
**Status**: ✅ Fully Supported

Handles direct audio/video URLs and streaming links.

```
Input: Direct URL (.mp3, .mp4, .m3u8, etc.)
↓
Validate → Return URL for streaming
```

**Features**:
- Direct streaming without download
- M3U8/HLS stream support
- MPEG-DASH support
- Automatic format detection
- Live stream detection

**When Used**:
- Direct audio/video URLs
- CDN links
- HLS/DASH streams
- Any direct media URL

**Priority Note**: Runs **before** YtDlp to handle direct streams that yt-dlp might fail on.

---

### 9. **YT-DLP** (Priority: 60)
**Status**: ✅ Free Method

Universal downloader for YouTube and other platforms.

```
Input: Any URL
↓
yt-dlp (local binary)
↓
Output: Audio/Video file
```

**Features**:
- Universal platform support
- Metadata extraction
- Complete local control
- Cookie-based authentication
- Smart URL detection
- Live stream detection
- Automatic fallback

**Configuration**:
```bash
COOKIES_LINK=https://batbin.me/paste_id1 https://batbin.me/paste_id2
```

**Installation**:
```bash
# macOS
brew install yt-dlp

# Linux
sudo apt install yt-dlp

# Windows
pip install yt-dlp
```

**New Features**:
- ✅ Can extract metadata from any URL
- ✅ Validates URLs using yt-dlp JSON extraction
- ✅ Skips direct streams (handled by DirectStream)
- ✅ Detects and rejects live streams
- ✅ Smart cookie usage (only for YouTube)
- ✅ Playlist support

**Pros**:
- ✅ Free forever
- ✅ Full control
- ✅ Works with most platforms
- ✅ Universal fallback

**Cons**:
- ⚠️ Requires yt-dlp installed
- ⚠️ Needs updated cookies for YouTube
- ⚠️ More resource-intensive
- ⚠️ Cannot handle live streams

---

## 📊 Priority System

| Priority | Platform | Purpose |
|----------|----------|---------|
| **100** | Telegram | Direct media files |
| **100** | Youtubify API | YouTube video downloads |
| **95** | Spotify | Spotify metadata + YouTube fallback |
| **90** | YouTube | Video metadata & search |
| **88** | JioSaavn | JioSaavn audio downloads |
| **85** | SoundCloud | SoundCloud downloads |
| **80** | Fallen API | YouTube audio downloads |
| **65** | DirectStream | Direct URLs & streams |
| **60** | YT-DLP | Universal fallback |

---
### Why Priority Matters

**Higher priority = checked first**

Example flow for direct stream URL:
```
Direct stream URL received
↓
Check Telegram (100) → ❌ Not valid for URL
↓
Check Youtubify (100) → ❌ Download-only
↓
Check Spotify (95) → ❌ Not Spotify
↓
Check YouTube (90) → ❌ Not YouTube
↓
Check SoundCloud (85) → ❌ Not SoundCloud
↓
Check Fallen API (80) → ❌ Download-only
↓
Check DirectStream (65) → ✅ Valid! Extract metadata
↓
Download needed → DirectStream returns URL
```

Example flow for YouTube video:
```
YouTube URL received
↓
Check Telegram (100) → ❌ Not valid for YouTube
↓
Check Youtubify (100) → ❌ Download-only
↓
Check Spotify (95) → ❌ Not Spotify
↓
Check YouTube (90) → ✅ Fetch metadata
↓
Download needed → Check Youtubify (100) or Fallen (80) or YtDlp (60)
```

---

## 🧠 Adding New Platforms

### Step 1: Create New File

```bash
# Create file for your platform
touch internal/platforms/myplatform.go
```

### Step 2: Define Struct

```go
package platforms

import (
    "context"
    state "main/internal/core/models"
    "github.com/amarnathcjd/gogram/telegram"
)

const PlatformMyPlatform state.PlatformName = "MyPlatform"

type MyPlatform struct {
    name state.PlatformName
    // Add API key, client, etc. if needed
    APIKey string
}
```

### Step 3: Implement Interface

```go
func (p *MyPlatform) Name() state.PlatformName {
    return p.name
}

func (p *MyPlatform) IsValid(query string) bool {
    // Return true if this platform can handle the query
    return strings.HasPrefix(query, "https://myservice.com/")
}

func (p *MyPlatform) GetTracks(query string, video bool) ([]*state.Track, error) {
    // Fetch and return track metadata
    // video flag indicates if user wants video playback
}

func (p *MyPlatform) IsDownloadSupported(source state.PlatformName) bool {
    // Return true if we can download from this source
    return source == PlatformMyPlatform
}

func (p *MyPlatform) Download(
    ctx context.Context,
    track *state.Track,
    mystic *telegram.NewMessage,
) (string, error) {
    // Download and return file path
    // Use mystic for progress updates
}
```

### Step 4: Register Platform

```go
func init() {
    // Pick a priority (higher = checked first)
    priority := 85
    Register(priority, &MyPlatform{
        name: PlatformMyPlatform,
        APIKey: os.Getenv("MY_API_KEY"),
    })
}
```

### Complete Example

```go
package platforms

import (
    "context"
    "errors"
    "fmt"
    "os"
    "strings"
    
    "github.com/amarnathcjd/gogram/telegram"
    state "main/internal/core/models"
)

const PlatformAppleMusic state.PlatformName = "AppleMusic"

type AppleMusicPlatform struct {
    name state.PlatformName
    token string
}

func init() {
    Register(87, &AppleMusicPlatform{
        name: PlatformAppleMusic,
        token: os.Getenv("APPLE_MUSIC_TOKEN"),
    })
}

func (a *AppleMusicPlatform) Name() state.PlatformName {
    return a.name
}

func (a *AppleMusicPlatform) IsValid(query string) bool {
    return strings.Contains(query, "music.apple.com")
}

func (a *AppleMusicPlatform) GetTracks(query string, video bool) ([]*state.Track, error) {
    if a.token == "" {
        return nil, errors.New("Apple Music token not configured")
    }
    
    // Implement Apple Music API integration
    // Return track metadata
    
    return nil, nil
}

func (a *AppleMusicPlatform) IsDownloadSupported(source state.PlatformName) bool {
    return false // Apple Music doesn't allow downloads
}

func (a *AppleMusicPlatform) Download(ctx context.Context, _ *state.Track, _ *telegram.NewMessage) (string, error) {
    return "", errors.New("Apple Music downloads not supported")
}
```

---

## 🔌 Platform Interface

### Full Interface Definition

```go
type Platform interface {
    // Unique platform identifier
    Name() state.PlatformName

    // Check if platform can handle this query
    IsValid(query string) bool

    // Fetch track metadata
    // video: true if video playback requested
    // Return tracks even if video not supported (set track.Video = false)
    GetTracks(query string, video bool) ([]*state.Track, error)

    // Check if we can download from specific source
    IsDownloadSupported(source state.PlatformName) bool

    // Download track and return local file path
    // Use mystic for progress updates (if provided)
    Download(ctx context.Context, track *state.Track, mystic *telegram.NewMessage) (string, error)
}
```

### Track Model

```go
type Track struct {
    ID        string          // Unique track ID
    Title     string          // Track name
    Duration  int             // Length in seconds
    Artwork   string          // Thumbnail URL
    URL       string          // Source URL
    Requester string          // User mention (HTML)
    Video     bool            // Video playback flag
    Source    PlatformName    // Which platform found this
}
```

---

## 🔧 Implementation Tips

### Error Handling

```go
// Always provide meaningful error messages
func (p *MyPlatform) GetTracks(query string, _ bool) ([]*state.Track, error) {
    if query == "" {
        return nil, errors.New("query cannot be empty")
    }
    
    if !p.IsValid(query) {
        return nil, fmt.Errorf("unsupported URL format: %s", query)
    }
    
    // Handle network errors gracefully
    // Don't crash, just return error
}
```

### Progress Updates

```go
func (p *MyPlatform) Download(ctx context.Context, track *state.Track, mystic *telegram.NewMessage) (string, error) {
    // Get progress manager from message
    pm := utils.GetProgress(mystic)
    
    // Download with progress updates
    // Progress will be sent to Telegram automatically
    
    // Handle cancellation
    select {
    case <-ctx.Done():
        return "", ctx.Err() // User cancelled
    default:
        // Continue download
    }
}
```

### Helper Functions

```go
// Use shared helper functions from base_platform.go
func (p *MyPlatform) Download(ctx context.Context, track *state.Track, _ *telegram.NewMessage) (string, error) {
    // Check if already downloaded
    if path, err := CheckDownloadedFile(track.ID); err == nil {
        return path, nil
    }

    // Ensure downloads directory exists
    if err := EnsureDownloadsDir(); err != nil {
        return "", err
    }

    // Your download logic here...
}
```

---

## 🐛 Troubleshooting

### Platform Not Being Used

**Problem**: Your platform is registered but not being called.

**Solution**:
1. Check `IsValid()` - Does it return `true` for your input?
2. Check priority - Is it higher than blocking platforms?
3. Add logging: `gologging.Debug("Platform check for: " + query)`

### Download Fails

**Problem**: `GetTracks()` works but `Download()` fails.

**Solution**:
1. Check `IsDownloadSupported()` - Returns `true` for source?
2. Check network access - Can you reach external APIs?
3. Check file permissions - Can you write to `downloads/` directory?

### Context Cancellation

**Problem**: Download doesn't stop when user cancels.

**Solution**:
```go
// Always check context in Download()
select {
case <-ctx.Done():
    return "", ctx.Err()
default:
    // Download here
}
```

### YtDlp Not Extracting Info

**Problem**: YtDlp fails to extract metadata from URL.

**Solution**:
1. Check if URL is a live stream (not supported)
2. Check if it's a direct stream (use DirectStream instead)
3. Verify yt-dlp is installed: `yt-dlp --version`
4. Update yt-dlp: `pip install --upgrade yt-dlp`

---



## 📊 Platform Comparison

| Feature | Telegram | Youtubify | Spotify | YouTube | JioSaavn | SoundCloud | Fallen | DirectStream | YT-DLP |
|---------|----------|-----------|---------|---------|----------|------------|--------|--------------|--------|
| **Setup** | Built-in | API Key | API Key | Built-in | Built-in | Binary | API Key | Built-in | Binary |
| **Audio** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Video** | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ |
| **Cost** | Free | Paid | Free | Free | Free | Free | Paid | Free | Free |
| **Metadata** | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ | ❌ | Basic | ✅ |
| **Quality** | Original | High | High | High | 320kbps | Best | Best | Varies | Best |

---

## 🎯 Best Practices

### 1. Validate Input Early

```go
func (p *MyPlatform) GetTracks(query string, _ bool) ([]*state.Track, error) {
    if query == "" {
        return nil, errors.New("empty query")
    }
    
    if !p.IsValid(query) {
        return nil, fmt.Errorf("invalid format: %s", query)
    }
    
    // Now process safely
}
```

### 2. Handle Context Cancellation

```go
func (p *MyPlatform) Download(ctx context.Context, track *state.Track, _ *telegram.NewMessage) (string, error) {
    // Create request with context
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    
    // Automatically cancelled if context done
    resp, err := http.DefaultClient.Do(req)
    
    // Check context
    if ctx.Err() != nil {
        return "", ctx.Err()
    }
}
```

### 3. Use Logging

```go
import "github.com/Laky-64/gologging"

func (p *MyPlatform) Download(...) (string, error) {
    gologging.InfoF("Starting download for: %s", track.ID)
    gologging.ErrorF("Download failed: %v", err)
}
```

---

## 🎯 Credits

### Third-Party Libraries & APIs

- **JioSaavn API Logic**: Adapted from [jiosaavn-api](https://github.com/sumitkolhe/jiosaavn-api/) by [Sumit Kolhe](https://github.com/sumitkolhe)
  - License: MIT License
  - Copyright (c) 2024 Sumit Kolhe
  - Used for: API endpoints, URL decryption, and metadata extraction logic

- **YouTube Search**: Web scraping logic adapted from [TgMusicBot](https://github.com/AshokShau/TgMusicBot)
  - License: GNU GPL v3
  - Copyright (c) 2025 Ashok Shau
  - Used for: YouTube search result parsing and metadata extraction

## 📞 Support

- Found a bug in a platform? Use `/bug` command
- Want to add a platform? Check examples above
- Join [Support Chat](https://t.me/TheTeamVk) for help

---

**Happy Platform Development! 🎼**