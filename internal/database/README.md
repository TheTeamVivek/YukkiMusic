# 💾 YukkiMusic Database System

> **MongoDB-based Data Management for YukkiMusic**

---

## 📋 Table of Contents

1. [Overview](#-overview)
2. [Database Schema](#-database-schema)
3. [Collections](#-collections)
4. [Core Operations](#-core-operations)
5. [Advanced Features](#-advanced-features)
6. [Caching Strategy](#-caching-strategy)
7. [Troubleshooting](#-troubleshooting)
8. [Performance Tuning](#-performance-tuning)

---

## 🌟 Overview

The **Database System** manages all persistent data for YukkiMusic using MongoDB.

**Location**: `internal/database/`

### Responsibilities

✅ **User Management** - Auth users, sudoers, served users  
✅ **Chat Settings** - Language, RTMP config, assistant assignment  
✅ **Bot State** - Maintenance mode, auto-leave, logger status  
✅ **Cache Management** - In-memory caching with TTL  
✅ **Data Migration** - Seamless upgrades from old versions  

### Technology Stack

- **Database**: MongoDB (Cloud or Local)
- **Driver**: `go.mongodb.org/mongo-driver/v2`
- **Caching**: In-memory with TTL expiration
- **Timeout**: Internal context management (5s default)

---

## 📊 Database Schema

### Collections Overview

```
YukkiMusic (Database)
├── bot_settings
│   └── Global bot state (1 document)
├── chat_settings
│   └── Per-chat configuration (many documents)
└── [Migration tracking]
```

### Document Structure

```javascript
// bot_settings collection
{
  "_id": "global",
  "served": {
    "users": [123456789, 987654321, ...],
    "chats": [-1001234567890, ...]
  },
  "sudoers": [123456789, ...],
  "autoleave": true,
  "logger": true,
  "maint": {
    "enabled": false,
    "reason": "Server maintenance"
  }
}

// chat_settings collection
{
  "_id": -1001234567890,
  "cplay_id": 0,
  "auth_users": [111111111, 222222222],
  "language": "en",
  "rtmp_config": {
    "rtmp_url": "rtmps://...",
    "rtmp_key": "..."
  },
  "ass_index": 2,
  "no_thumb": false
}
```

---

## 📚 Collections

### 1. bot_settings

**Purpose**: Global bot configuration and statistics

**Fields**:

| Field | Type | Purpose |
|-------|------|---------|
| `_id` | String | Always "global" (singleton) |
| `served.users` | Array | Served user IDs |
| `served.chats` | Array | Served chat IDs |
| `sudoers` | Array | Sudo user IDs |
| `autoleave` | Boolean | Auto-leave inactive chats |
| `logger` | Boolean | Logger enabled |
| `maint.enabled` | Boolean | Maintenance mode on/off |
| `maint.reason` | String | Maintenance reason message |

**Example**:
```javascript
{
  "_id": "global",
  "served": {
    "users": [123456789, 987654321],
    "chats": [-1001234567890]
  },
  "sudoers": [123456789],
  "autoleave": true,
  "logger": true,
  "maint": {
    "enabled": false,
    "reason": ""
  }
}
```

**Cached**: Yes (60 minutes TTL)

**Operations**:
```go
getBotState()           // Fetch current state (internal)
updateBotState(state)   // Update state (internal)
modifyBotState(fn)      // atomic-like modification (internal)
```

---

### 2. chat_settings

**Purpose**: Per-chat configuration and user management

**Fields**:

| Field | Type | Purpose |
|-------|------|---------|
| `_id` | Int64 | Chat ID (unique) |
| `cplay_id` | Int64 | Linked channel play ID |
| `auth_users` | Array | Authorized user IDs |
| `language` | String | Chat language code |
| `rtmp_config.rtmp_url` | String | RTMP streaming URL |
| `rtmp_config.rtmp_key` | String | RTMP stream key |
| `ass_index` | Int | Assigned assistant index |
| `no_thumb` | Boolean | Disable thumbnails |

**Example**:
```javascript
{
  "_id": -1001234567890,
  "cplay_id": 0,
  "auth_users": [111111111],
  "language": "en",
  "rtmp_config": {
    "rtmp_url": "rtmps://dc5-1.rtmp.t.me/s/",
    "rtmp_key": "2146211959:yJaXZGb7KXp..."
  },
  "ass_index": 1,
  "no_thumb": false
}
```

**Cached**: Yes (per-chat, 60 minutes TTL)

**Operations**:
```go
getChatSettings(chatID)          // Fetch settings (internal)
updateChatSettings(settings)     // Update settings (internal)
modifyChatSettings(chatID, fn)   // atomic-like modification (internal)
```

---

## 🔄 Core Operations

### User Management

```go
// Check if user is sudo
isSudo, err := database.IsSudo(userID)

// Add sudo user
err := database.AddSudo(userID)

// Remove sudo user
err := database.RemoveSudo(userID)

// Get all sudoers
sudoers, err := database.Sudoers()
```

### Served Statistics

```go
// Get all served users
users, err := database.ServedUsers()

// Get all served chats
chats, err := database.ServedChats()

// Check if served
isServed, err := database.IsServedUser(userID)
isServed, err := database.IsServedChat(chatID)

// Mark as served
err := database.AddServedUser(userID)
err := database.AddServedChat(chatID)

// Remove from served
err := database.RemoveServedUser(userID)
err := database.RemoveServedChat(chatID)
```

### Auth Users

```go
// Check if auth user
isAuth, err := database.IsAuthorized(chatID, userID)

// Add auth user
err := database.Authorize(chatID, userID)

// Remove auth user
err := database.Unauthorize(chatID, userID)

// Get all auth users for chat
users, err := database.AuthorizedUsers(chatID)
```

### Chat Language

```go
// Get chat language
lang, err := database.Language(chatID)
// Returns: chat language ("en", "hi", etc.), if not found defaults to config.DefaultLang

// Set chat language
err := database.SetLanguage(chatID, "hi")
```

### Channel Play (CPlay)

```go
// Get linked channel ID
cplayID, err := database.LinkedChannel(chatID)

// Set linked channel
err := database.LinkChannel(chatID, channelID)
```

### RTMP Configuration

```go
// Get RTMP settings
url, key, err := database.RTMP(chatID)

// Set RTMP settings
err := database.SetRTMP(chatID, url, key)
```

### Maintenance Mode

```go
// Check if maintenance enabled
isMaint, err := database.IsMaintenanceEnabled()

// Get maintenance reason
reason, err := database.MaintenanceReason()

// Set maintenance mode
err := database.SetMaintenance(true, "Server maintenance")
err := database.SetMaintenance(false)  // Disable
```

### Logger

```go
// Check if logger enabled
enabled, err := database.IsLoggerEnabled()

// Set logger status
err := database.SetLoggerEnabled(true)
```

### Auto-Leave

```go
// Check if auto-leave enabled
enabled, err := database.AutoLeave()

// Set auto-leave status
err := database.SetAutoLeave(true)
```

### Assistant Assignment

```go
// Get assigned assistant for chat
index, err := database.AssistantIndex(chatID, totalAssistants)

// Rebalance assistants across all chats
err := database.RebalanceAssistantIndexes(totalAssistants)
```

---

## 🚀 Advanced Features

### 1. Automatic Caching

```
Read Operation
    ↓
Check Cache (60 min TTL)
    ↓
If Hit: Return cached value ⚡
If Miss: Query MongoDB → Update cache
```

---

### 2. Data Migration

**Automatic Migration from v1 to v2**:

```go
func migrateData() {
    // Runs once on startup
    // Checks if old database exists
    // Migrates old collections
    // Sets migration flag
    // No manual action needed
}
```

---

### 3. Rebalancing Algorithm

Distributes chats evenly across assistant accounts:

```
Total Chats: 100
Total Assistants: 3

Distribution:
- Assistant 1: 34 chats
- Assistant 2: 33 chats
- Assistant 3: 33 chats

Algorithm:
1. Calculate base = 100 / 3 = 33
2. Calculate remainder = 100 % 3 = 1
3. First assistants get +1 (base + 1)
4. Others get base
5. Reassign excess chats
```

**Trigger**:
```go
// Called when assistants added/removed
database.RebalanceAssistantIndexes(newAssistantCount)
```

---

## 💾 Caching Strategy

### Cache Invalidation

Automatic on updates:

```go
func updateBotState(state *BotState) error {
    // Update MongoDB (internal)
    // ...
    
    if err == nil {
        // Update cache
        dbCache.Set(botStateCacheKey, state)
    }
    
    return err
}
```

---

## 🔗 Database Relationships

### Chat ↔ Assistant

```
One Chat → One Assistant (many-to-one)

Example:
chat_settings:
  _id: -1001234567890
  ass_index: 2  ← Points to Assistant #2

core.Assistants.Get(2)  ← The actual assistant
```

### User Hierarchies

```
Owner (1)
  ├─ Sudo Users (multiple)
  └─ Auth Users (per-chat)
      ├─ Admin Users (Telegram admins)
      └─ Regular Users
```

---

## ⚡ Performance Tuning

### 1. Query Optimization

The system uses `BulkWrite` for rebalancing and efficient internal fetching to minimize roundtrips.

---

## 📊 Database Statistics

Typical document size is ~2KB for bot state and ~500B for chat settings. Standard operations are cached, providing sub-millisecond read times.

---

## 📝 File Structure

```
internal/database/
├── README.md                  # This file
├── database.go                # Initialization & setup
├── helpers.go                 # Internal context & slice utilities
├── bot_state.go              # Global state management
├── chat_settings.go          # Per-chat settings
├── auth_users.go             # Authorization management
├── served_stats.go           # User/chat tracking
├── sudo_users.go             # Sudo management
├── autoleave.go              # Auto-leave configuration
├── logger.go                 # Logger status
├── language.go               # Language preferences
├── cplay.go                  # Channel play management
├── rtmp_cfg.go               # RTMP configuration
├── assistant.go              # Assistant assignment
├── maintenance.go            # Maintenance mode
└── migrate_data.go           # Migration logic
```

---

## 🔧 Function Reference

### Common Operations

```go
// Check/Add/Remove sudo user
isSudo, _ := database.IsSudo(userID)
database.AddSudo(userID)
database.RemoveSudo(userID)

// Manage auth users per chat
isAuth, _ := database.IsAuthorized(chatID, userID)
database.Authorize(chatID, userID)
database.Unauthorize(chatID, userID)

// Language management
lang, _ := database.Language(chatID)
database.SetLanguage(chatID, "hi")

// Served tracking
database.AddServedUser(userID)
database.AddServedChat(chatID)

// Maintenance
database.SetMaintenance(true, "reason")
```

---

## 🆘 Support

- **Issues?** Use `/bug` command
- **Help needed?** Join [Support Chat](https://t.me/TheTeamVk)
- **Report bug?** [GitHub Issues](https://github.com/TheTeamVivek/YukkiMusic/issues)

---

**Keep data clean! 💾**
