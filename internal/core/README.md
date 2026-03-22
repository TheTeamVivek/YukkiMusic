# 🏛️ YukkiMusic Core Package

The `core` package is the central orchestration and state management layer of the YukkiMusic bot. It maintains the real-time status of active playback sessions (Rooms), tracks chat-specific properties (ChatState), and manages the lifecycle of Telegram assistant accounts.

## 🌟 Overview

This package provides thread-safe access to the bot's state across multiple goroutines. It acts as the glue between the Telegram API (`gogram`), the media playback engine (`ntgcalls`), and the feature-specific `modules`.

## 📦 Key types

- **`RoomState`**: Represents an active voice chat playback session. It holds the current track, playback position, queue, and speed. It also manages scheduled timers for auto-resume and auto-unmute.
- **`ChatState`**: Tracks transient chat metadata such as whether the assistant is present, banned, or if the voice chat is active. It handles the logic for the assistant joining a chat.
- **`Assistant`**: Wraps a Telegram client and its associated media context (`ntgcalls` wrapper). Each assistant represents a separate Telegram account used for streaming.
- **`AssistantManager`**: Manages the pool of available assistants and handles their distribution across different chats.
- **`Track`** (in `models` subpackage): Models the metadata for a single media item, including its source, duration, and requester.

## 🚀 Usage

### 🎵 Room Management
The `RoomState` is the primary entry point for playback control.

```go
// Get or create a room for a chat
ass, err := core.Assistants.ForChat(chatID)
if err != nil {
    return err
}
room, _ := core.GetRoom(chatID, ass, true)

// Add a track to the queue and start playing if idle
err := room.Play(track, localPath, false)
```

### 👤 Chat State and Joining
`ChatState` is used to ensure the assistant is ready to stream in a chat.

```go
state, err := core.GetChatState(chatID)
if err != nil {
    return err
}

// Check presence and attempt to join if missing
present, err := state.IsAssistantPresent(false)
if err != nil {
    return err
}

if !present {
    err = state.TryJoin()
}
```

## 🔒 Concurrency

All state types (`RoomState`, `ChatState`, `AssistantManager`) are thread-safe. They use internal `sync.RWMutex` locks to protect their fields. Methods that modify state (like `Play`, `Pause`, `SetLoop`) handle locking internally. 

> [!WARNING]
> **Deadlock Avoidance**: When calling methods that modify `RoomState.Data` while holding the `RoomState` lock, use caution. Internal methods like `cleanupFile` and `updatePosition` are designed to be safe when called from within the state machine.

## 🏗️ What core does not own

- **Media Resolution**: The `core` package does not know how to find tracks on YouTube or Spotify; that belongs in `platforms`.
- **Persistence**: Permanent settings like authorized users or sudoers are managed by the `database` package.
- **Command Parsing**: Logic for individual bot commands (e.g., `/play`, `/skip`) lives in `modules`.
- **Low-level Streaming**: The actual byte-shuffling and FFmpeg management is handled by `ntgcalls`.

## 💡 Notes

> [!IMPORTANT]
> **Initialization**: `core.Init()` must be called during bot startup to initialize the bot client and all configured assistants. It returns a shutdown function and an error.

> [!IMPORTANT]
> **Cleanup**: `DeleteRoom(chatID)` must be called when a session ends to stop playback, clean up temporary files, and release resources.

- **Assistant Indexing**: `GetAssistantIndexFunc` must be initialized (usually to `database.AssistantIndex`) to allow the `AssistantManager` to distribute chats across the assistant pool.
