# 🎮 YukkiMusic Modules & Commands System

The **Modules System** implements all bot commands and features through a modular handler pipeline. It is responsible for parsing user input, enforcing permissions via filters, and orchestrating high-level bot behavior by interacting with the `core`, `database`, and `platforms` packages.

## 🌟 Overview

Each file in this package typically handles a specific feature area (e.g., `play.go`, `skip.go`, `sudoers.go`). The system follows a decentralized architecture where individual modules register their own handlers during initialization.

### ✅ Key Concepts

- **Modular Design**: Each feature is isolated in its own file.
- **Filter Pipeline**: Permissions and context (e.g., "must be admin") are checked before handler execution.
- **Panic Recovery**: Every handler is wrapped in a safe execution block that logs crashes and sends debug info to the owner.
- **Localization**: All user-facing messages are translated using the `F(chatID, key)` helper.

## 📝 Module registration

The central registration point is `handlers.go`. It defines lists of message handlers (`handlers`) and callback query handlers (`cbHandlers`). These are initialized and registered with the bot client in the `Init()` function.

### 🏗️ Adding a new command

To add a new command, define its logic in a handler function and add it to the `handlers` slice in `handlers.go`.

```go
// 1. Create your handler in a new or existing module file
func myCommandHandler(m *telegram.NewMessage) error {
    chatID := m.ChannelID()
    
    // Command logic here
    m.Reply(F(chatID, "success_message"))
    
    return telegram.ErrEndGroup
}

// 2. Register it in handlers.go
var handlers = []MsgHandlerDef{
    {
        Pattern: "mycommand",
        Handler: myCommandHandler,
        Filters: []telegram.Filter{superGroupFilter, authFilter},
    },
    // ...
}
```

## 📐 Standard Module Structure

Modules typically follow this lifecycle:
1. **Validation**: Check arguments and permissions.
2. **State Retrieval**: Get the relevant `core.RoomState` or `core.ChatState`.
3. **Action**: Invoke methods on `core` types or `database` helpers.
4. **Response**: Send a translated message back to the user.

### 🎵 Example: Skip Command (`skip.go`)

```go
func skipHandler(m *tg.NewMessage) error {
    r, err := getEffectiveRoom(m, false) // helper from helpers.go
    if err != nil {
        return err
    }

    next := r.NextTrack()
    if next == nil {
        m.Reply(F(m.ChannelID(), "queue_empty"))
        return r.Stop()
    }

    // Play handles the internal orchestration
    return r.Play(next, "", false)
}
```

## 🔐 Filters & Permissions

Filters are located in `filters.go` and are used to restrict command access. Common filters include:

- `ownerFilter`: Only the bot owner.
- `sudoOnlyFilter`: Owner or sudo users.
- `adminFilter`: Only chat administrators.
- `authFilter`: Administrators or users authorized via `/addauth`.
- `superGroupFilter`: Restricts command to supergroups and automatically handles bot leaving basic groups.

## 🌉 Dependencies and Boundaries

- **Imports `core`**: Modules rely on the `core` package for all real-time playback and state management.
- **Imports `database`**: Uses the database for persistent settings, auth lists, and statistics.
- **Imports `platforms`**: Uses platforms for resolving search queries and URLs into track metadata.

> [!CAUTION]
> **No Direct State Storage**: Modules should never store state themselves; all transient state must live in `core` and persistent state in `database`.

## 🛡️ Error Handling

> [!NOTE]
> All handlers must be wrapped with `SafeMessageHandler` or `SafeCallbackHandler` in `handlers.go`. 

This provides:
- Automatic maintenance mode blocking.
- Panic recovery and stack trace logging.
- Consistent error logging to the bot's logger chat.
