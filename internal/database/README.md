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
- **Timeout**: 5-30 seconds per operation

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
  "ass_index": 2
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
getBotState()           // Fetch current state
updateBotState(state)   // Update state
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
  "ass_index": 1
}
```

**Cached**: Yes (per-chat, 60 minutes TTL)

**Operations**:
```go
getChatSettings(chatID)          // Fetch settings
updateChatSettings(settings)     // Update settings
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
err := database.DeleteSudo(userID)

// Get all sudoers
sudoers, err := database.GetSudoers()
```

### Served Statistics

```go
// Get all served users
users, err := database.GetServed(true)  // true = get users

// Get all served chats
chats, err := database.GetServed()      // false = get chats

// Check if served
isServed, err := database.IsServed(userID, true)

// Mark as served
err := database.AddServed(userID, true)

// Remove from served
err := database.DeleteServed(userID, true)
```

### Auth Users

```go
// Check if auth user
isAuth, err := database.IsAuthUser(chatID, userID)

// Add auth user
err := database.AddAuthUser(chatID, userID)

// Remove auth user
err := database.RemoveAuthUser(chatID, userID)

// Get all auth users for chat
users, err := database.GetAuthUsers(chatID)
```

### Chat Language

```go
// Get chat language
lang, err := database.GetChatLanguage(chatID)
// Returns: chat language ("en", "hi", etc.), if not found defaults to config.DEFAULT_LANG

// Set chat language
err := database.SetChatLanguage(chatID, "hi")
```

### Channel Play (CPlay)

```go
// Get linked channel ID
cplayID, err := database.GetCPlayID(chatID)

// Set linked channel
err := database.SetCPlayID(chatID, cplayID)

// Get chat from channel ID
chatID, err := database.GetChatIDFromCPlayID(cplayID)
```

### RTMP Configuration

```go
// Get RTMP settings
url, key, err := database.GetRTMP(chatID)

// Set RTMP settings
err := database.SetRTMP(chatID, url, key)
```

### Maintenance Mode

```go
// Check if maintenance enabled
isMaint, err := database.IsMaintenance()

// Get maintenance reason
reason, err := database.GetMaintReason()

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
enabled, err := database.GetAutoLeave()

// Set auto-leave status
err := database.SetAutoLeave(true)
```

### Assistant Assignment

```go
// Get assigned assistant for chat
index, err := database.GetAssistantIndex(chatID, totalAssistants)

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

**Cache Keys**:
```go
"bot_state"                           // Global state
"chat_settings_" + chatID            // Per-chat settings
"cplayid_" + cplayID                 // Channel play mapping
```

**Manual Cache Invalidation**:
```go
dbCache.Delete("bot_state")
dbCache.Delete("chat_settings_" + chatID)
```

---

### 2. Data Migration

**Automatic Migration from v1 to v2**:

```go
func migrateData(mongoURI string) {
    // Runs once on startup
    // Checks if old database exists
    // Migrates old collections
    // Sets migration flag
    // No manual action needed
}
```

**Migrates**:
- ✅ cplaymode → cplay_id
- ✅ tgusersdb → served.users
- ✅ chats → served.chats
- ✅ sudoers → sudoers array

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

### Cache Hierarchy

```
User Request
    ↓
Level 1: In-Memory Cache (< 1ms)
    ├─ If Hit: Return immediately
    └─ If Miss: Go to Level 2
    ↓
Level 2: MongoDB Query (50-500ms)
    ├─ If Found: Cache result
    └─ If Not Found: Use default
    ↓
Return Result
```

### Cache Configuration

```go
// Default cache expiration
const defaultTTL = 60 * time.Minute

// Create cache
dbCache := utils.NewCache[string, any](defaultTTL)

// Set value with default TTL
dbCache.Set("key", value)

// Set value with custom TTL
dbCache.Set("key", value, 30*time.Minute)

// Get value
val, exists := dbCache.Get("key")

// Delete value
dbCache.Delete("key")
```

### Cache Invalidation

Automatic on updates:

```go
func updateBotState(state *BotState) error {
    // Update MongoDB
    _, err := settingsColl.UpdateOne(ctx, filter, update)
    
    if err == nil {
        // Invalidate cache
        dbCache.Set("bot_state", state)  // Update
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

### Chat ↔ CPlay (Channel)

```
One Chat ↔ Zero or One Channel (one-to-zero-or-one)

Example:
chat_settings:
  _id: -1001234567890 (main chat)
  cplay_id: -1001111111111 (linked channel)

Then:
- Play in main chat
- Stream outputs to linked channel
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

```go
// ❌ Bad: Multiple queries
for chatID := range chatIDs {
    settings, _ := getChatSettings(chatID)  // N queries
}

// ✅ Good: Batch query
chats, _ := chatSettingsColl.Find(ctx, bson.M{
    "_id": bson.M{"$in": chatIDs},
})
```

### 2. Cache Tuning

```go
// Default: 60 minutes
// Increase for stable data:
cache := utils.NewCache(2 * time.Hour)

// Decrease for frequently-changing data:
cache := utils.NewCache(5 * time.Minute)
```
---

## 📊 Database Statistics

### Typical Document Sizes

| Collection | Avg Size | Count | Total |
|-----------|----------|-------|-------|
| bot_settings | 2 KB | 1 | 2 KB |
| chat_settings | 500 B | 1000 | 500 KB |
| **Total** | - | - | ~502 KB |

### Typical Query Times

| Operation | Time | Cached |
|-----------|------|--------|
| Get bot state | 100ms | < 1ms |
| Get chat settings | 50ms | < 1ms |
| List sudoers | 60ms | < 1ms |
| Rebalance (1000 chats) | 5s | N/A |

---

## 🔐 Security Best Practices

### 1. Environment Variables

```bash
# ✅ Good
MONGO_DB_URI=mongodb+srv://user:pass@cluster.mongodb.net/YukkiMusic

# ❌ Bad (in code)
mongoURI := "mongodb+srv://user:pass@..."
```

### 2. Connection Security

```bash
# ✅ Use MongoDB Atlas with:
# - IP Whitelisting enabled
# - TLS/SSL required
# - Strong password

# ❌ Avoid local MongoDB without auth
```

### 3. Data Access

```go
// ✅ Validate user before data access
if userID != config.OwnerID && !isSudo(userID) {
    return fmt.Errorf("unauthorized")
}

// ❌ Don't expose sensitive data
_ = database.GetSudoers()  // Public query
```

---

## 📝 File Structure

```
internal/database/
├── README.md                  # This file
├── database.go                # Initialization & setup
├── helpers.go                 # Utility functions
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
database.DeleteSudo(userID)

// Manage auth users per chat
isAuth, _ := database.IsAuthUser(chatID, userID)
database.AddAuthUser(chatID, userID)
database.RemoveAuthUser(chatID, userID)

// Language management
lang, _ := database.GetChatLanguage(chatID)
database.SetChatLanguage(chatID, "hi")

// Served tracking
database.AddServed(userID, true)  // Mark user
database.AddServed(chatID)        // Mark chat

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
