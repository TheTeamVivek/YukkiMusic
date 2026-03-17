# 🎮 YukkiMusic Modules & Commands System

> **Comprehensive Command Handler Architecture**

---

## 📋 Table of Contents

1. [Overview](#-overview)
2. [Module Structure](#-module-structure)
3. [Command Categories](#-command-categories)
4. [Handler Pipeline](#-handler-pipeline)
5. [Filters & Permissions](#-filters--permissions)
6. [Command Implementation](#-command-implementation)
7. [Error Handling](#-error-handling)
8. [Best Practices](#-best-practices)

---

## 🌟 Overview

The **Modules System** implements all bot commands and features through a handler pipeline.

**Location**: `internal/modules/`

### Key Concepts

✅ **Modular Design** - Each file handles specific feature area  
✅ **Filter Pipeline** - Permissions checked before execution  
✅ **Error Recovery** - Panics caught and logged safely  
✅ **Help System** - Built-in help for every command  
✅ **Localization** - All messages translated  

### Module Organization

```
internal/modules/
├── handlers.go              # Command registration & setup
├── helpers.go               # Shared utilities
├── filters.go               # Permission filters
├── flag_help.go             # Help flag handling
│
├── PLAYBACK CONTROL
├── play.go                  # Play command
├── skip.go                  # Skip command
├── pause.go                 # Pause command
├── resume.go                # Resume command
├── mute.go                  # Mute command
├── unmute.go                # Unmute command
├── seek.go                  # Seek/seekback/jump
├── replay.go                # Replay command
├── speed.go                 # Speed control
│
├── QUEUE MANAGEMENT
├── queue.go                 # Queue listing
├── remove.go                # Remove from queue
├── clear.go                 # Clear queue
├── move.go                  # Move in queue
├── shuffle.go               # Shuffle queue
├── loop.go                  # Loop tracking
│
├── ADMIN FEATURES
├── auth.go                  # Auth user management
├── stop.go                  # Stop playback
├── reload.go                # Reload admin cache
├── position.go              # Show position
│
├── BOT CONTROL
├── sudoers.go               # Sudo user management
├── maint.go                 # Maintenance mode
├── logger.go                # Logger control
├── autoleave.go             # Auto-leave config
├── active.go                # Active chats
├── stats.go                 # Bot statistics
│
├── UTILITIES
├── help.go                  # Help system
├── ping.go                  # Ping/status
├── start.go                 # Start command
├── bug.go                   # Bug reporting
├── language.go              # Language selection
├── broadcast.go             # Broadcast messages
│
├── DEVELOPER
├── dev.go                   # Shell/JSON commands
├── eval.go                  # Code evaluation
├── watcher.go               # Event watchers
├── monitor.go               # Room monitoring
├── restart.go               # Restart command
│
└── CHANNEL PLAY
    └── Various c* commands
```

---

## 🏗️ Module Structure

### Standard Module Pattern

```go
package modules

import (
    "github.com/amarnathcjd/gogram/telegram"
    "main/internal/locales"
)

// Help text (optional)
func init() {
    helpTexts["mycommand"] = "Description and usage"
}

// Main handler
func mycommandHandler(m *telegram.NewMessage) error {
    chatID := m.ChannelID()
    
    // 1. Validation
    if m.Args() == "" {
        m.Reply(F(chatID, "usage_message"))
        return telegram.ErrEndGroup
    }
    
    // 2. Business logic
    result, err := performAction(m.Args())
    if err != nil {
        m.Reply(F(chatID, "error_key", locales.Arg{
            "error": err.Error(),
        }))
        return telegram.ErrEndGroup
    }
    
    // 3. Response
    m.Reply(F(chatID, "success_key", locales.Arg{
        "result": result,
    }))
    
    return telegram.ErrEndGroup
}
```

---

## 📂 Command Categories

### 1. Playback Control

**Files**: `play.go`, `skip.go`, `pause.go`, `resume.go`, `mute.go`, `unmute.go`, `seek.go`, `replay.go`, `speed.go`

#### Available Commands

| Command | Description | Admin Only |
|---------|-------------|-----------|
| `/play` | Play song from URL/search | ❌ |
| `/fplay` | Force play (skip queue) | ✅ |
| `/skip` | Skip to next track | ✅ |
| `/pause [seconds]` | Pause playback | ✅ |
| `/resume` | Resume playback | ✅ |
| `/mute [seconds]` | Mute audio | ✅ |
| `/unmute` | Unmute audio | ✅ |
| `/seek <seconds>` | Seek forward | ✅ |
| `/seekback <seconds>` | Seek backward | ✅ |
| `/jump <position>` | Jump to position | ✅ |
| `/replay` | Replay current track | ✅ |
| `/speed <speed>` | Set speed (0.5-4.0x) | ✅ |

#### Implementation Example: Play

```go
func handlePlay(m *telegram.NewMessage, opts *playOpts) error {
    // 1. Prepare room
    r, replyMsg, err := prepareRoomAndSearchMessage(m, opts.CPlay)
    if err != nil {
        return telegram.ErrEndGroup
    }
    
    // 2. Fetch tracks
    tracks, isActive, err := fetchTracksAndCheckStatus(m, replyMsg, r, opts.Video)
    if err != nil {
        return telegram.ErrEndGroup
    }
    
    // 3. Filter and validate
    tracks, availableSlots, err := filterAndTrimTracks(replyMsg, r, tracks)
    if err != nil {
        return telegram.ErrEndGroup
    }
    
    // 4. Play tracks
    err := playTracksAndRespond(m, replyMsg, r, tracks, mention, isActive, opts.Force, availableSlots)
    if err != nil {
        return err
    }
    
    return telegram.ErrEndGroup
}
```

---

## 📂 Queue Management

**Files**: `queue.go`, `remove.go`, `clear.go`, `move.go`, `shuffle.go`, `loop.go`

| Command | Description | Admin Only |
|---------|-------------|-----------|
| `/queue` | Show queue | ❌ |
| `/position` | Current position | ❌ |
| `/remove <index>` | Remove track | ✅ |
| `/clear` | Clear all tracks | ✅ |
| `/move <from> <to>` | Reorder tracks | ✅ |
| `/shuffle [on/off]` | Toggle shuffle | ✅ |
| `/loop <count>` | Set loop count | ✅ |

---

## 📂 User Management

**Files**: `auth.go`, `sudoers.go`

| Command | Description | Requires |
|---------|-------------|----------|
| `/addauth <user>` | Add auth user | Admin |
| `/delauth <user>` | Remove auth user | Admin |
| `/authlist` | List auth users | Any |
| `/addsudo <user>` | Add sudo | Owner |
| `/delsudo <user>` | Remove sudo | Owner |
| `/sudolist` | List sudoers | Any |

---

## 📂 Bot Management

**Files**: `maint.go`, `logger.go`, `autoleave.go`, `active.go`

| Command | Description | Requires |
|---------|-------------|----------|
| `/maintenance` | Maintenance mode | Owner |
| `/logger` | Logger control | Sudo |
| `/autoleave` | Auto-leave config | Sudo |
| `/ac` | Active chats | Sudo |

---

## 📂 Channel Play (CPlay)

**Files**: `play.go` (cplay variants)

| Command | Description | Admin Only |
|---------|-------------|-----------|
| `/cplay <query>` | Play in channel | ✅ |
| `/cfplay <query>` | Force play in channel | ✅ |
| `/cpause` | Pause in channel | ✅ |
| `/cresume` | Resume in channel | ✅ |
| `/cskip` | Skip in channel | ✅ |
| `/cqueue` | Queue in channel | ✅ |
| `/cspeed` | Speed in channel | ✅ |

---

## 🔄 Handler Pipeline

### Request Flow

```
User Command
    ↓
1. MESSAGE RECEIVED (telegram.NewMessage)
    ↓
2. FILTER CHECK (Permissions)
    ├─ Owner filter
    ├─ Sudo filter
    ├─ Admin filter
    ├─ Auth user filter
    ├─ Supergroup filter
    └─ Channel filter
    ↓
3. PANIC RECOVERY (SafeMessageHandler)
    └─ Catches panics, logs them
    ↓
4. HELP FLAG CHECK
    ├─ If "-h" or "--help" flag present
    └─ Show command help
    ↓
5. MAINTENANCE CHECK
    ├─ If maintenance mode on
    └─ Block non-owner/sudo users
    ↓
6. COMMAND HANDLER
    ├─ Validate input
    ├─ Execute business logic
    ├─ Generate response
    └─ Send reply
    ↓
7. ERROR HANDLING
    ├─ Log errors
    ├─ Send error message
    └─ Continue
```

### Code Flow

```go
// handlers.go
var handlers = []MsgHandlerDef{
    {
        Pattern: "play",
        Handler: playHandler,
        Filters: []telegram.Filter{superGroupFilter, authFilter}
    },
    // ... more handlers
}

// In Init()
for _, h := range handlers {
    bot.AddCommandHandler(h.Pattern, SafeMessageHandler(h.Handler), h.Filters...)
}
```

---

## 🔐 Filters & Permissions

### Available Filters

```go
// filters.go
var (
    superGroupFilter    = Custom(filterSuperGroup)    // Must be supergroup
    adminFilter         = Custom(filterChatAdmins)    // Must be chat admin
    authFilter          = Custom(filterAuthUsers)     // Admin or auth user
    ignoreChannelFilter = Custom(filterChannel)       // Not from channel
    sudoOnlyFilter      = Custom(filterSudo)          // Must be sudo/owner
    ownerFilter         = Custom(filterOwner)         // Must be owner
)
```

### Filter Implementation

```go
func filterSuperGroup(m *telegram.NewMessage) bool {
    if !filterChannel(m) {
        return false
    }
    
    switch m.ChatType() {
    case telegram.EntityChat:
        // EntityChat can be basic group or supergroup
        if m.Channel != nil && !m.Channel.Broadcast {
            database.AddServedChat(m.ChannelID())
            return true  // Supergroup ✓
        }
        warnAndLeave(m.Client, m.ChannelID())  // Basic group ✗
        database.RemoveServedChat(m.ChannelID())
        return false
        
    case telegram.EntityChannel:
        return false  // Pure channel ✗
        
    case telegram.EntityUser:
        m.Reply(F(m.ChannelID(), "only_supergroup"))
        database.AddServedUser(m.ChannelID())
        return false  // Private chat ✗
    }
    
    return false
}
```

---

## 🛡️ Error Handling

### Safe Handler Wrapper

```go
func SafeMessageHandler(handler func(*tg.NewMessage) error) func(*tg.NewMessage) error {
    return func(m *tg.NewMessage) (err error) {
        // Check maintenance mode
        if is, _ := database.IsMaintenanceEnabled(); is {
            if m.SenderID() != config.OwnerID {
                if ok, _ := database.IsSudo(m.SenderID()); !ok {
                    reason, _ := database.MaintenanceReason()
                    msg := F(m.ChannelID(), "maint", locales.Arg{"reason": reason})
                    m.Reply(msg)
                    return tg.ErrEndGroup
                }
            }
        }
        
        // Panic recovery
        defer func() {
            if r := recover(); r != nil {
                handlePanic(r, m, true)  // Log panic
                err = fmt.Errorf("internal panic occurred")
            }
        }()
        
        // ... execute handler ...
    }
}
```

---

## 🎯 Common Patterns

### Pattern 1: Database Operation

```go
func mycommandHandler(m *telegram.NewMessage) error {
    chatID := m.ChannelID()
    
    // No context needed in signatures! Managed internally by database package.
    
    // Example: Toggle setting
    current, _ := database.ThumbnailsDisabled(chatID)
    err := database.SetThumbnailsDisabled(chatID, !current)
    
    return telegram.ErrEndGroup
}
```

---

## 📞 Support

- **Issues?** Use `/bug` command in bot
- **Help?** Join [Support Chat](https://t.me/TheTeamVk)
- **Report?** [GitHub Issues](https://github.com/TheTeamVivek/YukkiMusic/issues)

---
**Build amazing commands! 🎮**
