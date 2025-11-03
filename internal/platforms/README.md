# 🎧 **YukkiMusic Platform System**

Welcome to the **YukkiMusic Platform System Guide** — your complete walkthrough for understanding how platforms work and how you can easily add new ones to extend the bot’s music-handling power! 🚀  

---

## 🌍 **Overview**

The **Platform System** in YukkiMusic is a **modular, extensible, and priority-based framework** that allows the bot to fetch and download tracks from multiple sources — like **YouTube**, **Telegram**, or even **custom APIs**.  

Each platform is a self-contained unit ⚙️ that knows how to handle its own logic for fetching and downloading tracks.

> 💡 When a user requests a song, the bot checks all registered platforms (in order of priority) and uses the first one capable of handling the request.

---

## ⚙️ **How It Works**

The heart of the system lies in `registry.go`, where a **map-based registry** holds all registered platforms.  
Each entry has:
- 🧩 a unique **PlatformName**
- 🔢 a **priority value**
- 🧠 a **Platform implementation**

When `getOrderedPlatforms()` is called:
- All registered platforms are sorted in **descending order of priority**.
- The bot iterates through them and uses the first platform that supports the given query.

---
## 🔝 **Priority System**

| Priority | Platform | Description |
|-----------|-----------|--------------|
| 100 | Telegram | Handles direct Telegram audio/video links |
| 90 | YouTube | Extracts track metadata only (no downloading) |
| 70 | YtDlp | Downloads songs using cookies |

> [!NOTE]
> YukkiMusic always picks the **highest priority** platform that can handle the query.  
> For example:  
> - YouTube → used first for track info 🎵  
> - API or YtDlp → used for downloading the track ⬇️  
> - Telegram → handles direct Telegram media 💬  

> [!NOTE]
> ⚙️ **Rule:** Higher numbers = higher priority.

This ensures predictable and optimized handling.  
For example:  
- Telegram links → handled first ✅  
- YouTube URLs → checked next 🎬  
- Other APIs → checked later 🌐  
---

## 🧩 **The Platform Interface**

To create a new platform, implement the `state.Platform` interface (found in `internal/state/models.go`):

```go
type Platform interface {
    // A unique name for the platform.
    Name() PlatformName

    // Checks if this platform can return a track for given query.
    IsValid(query string) bool

    // Fetches track metadata.
    GetTracks(query string) ([]*Track, error)

    // Checks if downloading is supported from a given source.
    IsDownloadSupported(source PlatformName) bool

    // Downloads a track and returns the local file path.
    Download(ctx context.Context, track *Track, mystic *telegram.NewMessage) (string, error)
}
```

---

## 🧠 **How to Add a New Platform**

Adding a new platform is easy! Just follow these simple steps. 🌱  

---

### 🪪 **1. Define a Platform Constant**

Add your new platform constant inside `internal/state/models.go`:

```go
// internal/state/models.go

const (
    PlatformYouTube     PlatformName = "YouTube"
    PlatformTelegram    PlatformName = "Telegram"
    // ... other platforms
    PlatformMyPlatform  PlatformName = "MyPlatform" // 🆕 Add yours here!
)
```

---

### 📁 **2. Create a New File**

Create a new Go file for your platform:  
`internal/platforms/my_platform.go`

---

### 🧱 **3. Define Your Platform Struct**

Define the structure for your platform — include API keys or clients if needed.

```go
// internal/platforms/my_platform.go

package platforms

import (
    "context"
    
    "github.com/amarnathcjd/gogram/telegram"
    
    "github.com/TheTeamVivek/YukkiMusic/internal/state"
)

type MyPlatform struct {
    // Example: APIKey string
}
```

---

### 🧩 **4. Implement the Platform Interface**

Implement all interface methods step by step:

```go
func (p *MyPlatform) Name() state.PlatformName {
    return state.PlatformMyPlatform
}

func (p *MyPlatform) IsValid(query string) bool {
    // ✅ Check if your platform can handle the query
    return strings.HasPrefix(query, "https://my-service.com/")
}

func (p *MyPlatform) GetTracks(query string) ([]*state.Track, error) {
    // 🎵 Fetch and return track metadata
    // ...
}

func (p *MyPlatform) IsDownloadSupported(source state.PlatformName) bool {
    // 💾 Can this platform download the song for source
    return source == state.PlatformMyPlatform // yep this can Download song of its
}

func (p *MyPlatform) Download(ctx context.Context, track *state.Track, mystic *telegram.NewMessage) (string, error) {
    // ⬇️ Actual download logic here
    // ...
    // you can use mystic for displaying progress of download, make sure just progress if you encountered any error so just return it bot will handle it
}
```

---

### 🧩 **5. Register Your Platform**

Finally, register your platform so the bot recognizes it!  
Pick a **priority value** wisely.  
(Higher = checked earlier)

```go
func init() {
    priority := 85 // ⚡ Example: between YouTube (90) and Telegram (100)
    AddPlatform(priority, state.PlatformMyPlatform, &MyPlatform{})
}
```

---

## 🎉 **All Set!**

That’s it! 🎊  
You’ve just added a brand-new platform to **YukkiMusic**.  

Your new source will now:
- 🌐 Appear in the global registry  
- 🧠 Be auto-detected when fetching songs  
- 🔊 Support downloading (if implemented)

> [!TIP]
> 🧩 You can experiment with custom platforms like **SoundCloud**, **Spotify APIs**, or **Your Own Music API**!

---

## 💡 **Pro Tip**

You can view all registered platforms by checking your logs or calling the `getOrderedPlatforms()` function — it will list all platforms sorted by priority.

---

## 💫 **Example Platform Flow**

1. 🎵 **User Command:** `/play https://youtube.com/watch?v=xyz123`  
2. ⚙️ **Registry Loads Platforms** (sorted by priority):  
   `Telegram (100) → YouTube (90) → InflexAPI (80)`  
3. 🔍 **Validation:**  
   - Telegram ❌  
   - YouTube ✅ → Handles request  
4. 📦 **Track Fetched:** Metadata like title, duration, source  
5. ⬇️ **Download/Stream:** Audio fetched locally or streamed directly  
6. 🎧 **Playback:**  
   > **Now Playing:** Perfect – Ed Sheeran 🎶 (Source: YouTube)

✅ **Flow Summary:**  
YukkiMusic checks each platform by priority → first valid one handles → fetch → download → play.  
> “First valid platform wins, and music begins!” 🎵
