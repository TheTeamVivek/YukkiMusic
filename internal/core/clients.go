/*
 * ● YukkiMusic
 * ○ A high-performance engine for streaming music in Telegram voicechats.
 *
 * Copyright (C) 2026 TheTeamVivek
 *
 * This program is free software: you can redistribute it and/or modify it under the
 * terms of the GNU General Public License as published by the Free Software Foundation,
 * either version 3 of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
 * PARTICULAR PURPOSE. See the GNU General Public License for more details.
 *
 * Repository: https://github.com/TheTeamVivek/YukkiMusic
 */

package core

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config"
	"yukkimusic/internal/utils"
)

// Init initializes the bot and assistant clients.
// It returns a shutdown function and an error if initialization fails.
func Init() (func(), error) {
	utils.Log.Info("Starting bot client...")
	if err := initBot(); err != nil {
		return nil, err
	}

	utils.Log.Info("Starting assistant clients...")
	cm := &ClientManager{
		list: make([]*Assistant, 0, len(config.StringSessions)),
	}
	if err := initAssistants(cm); err != nil {
		return nil, fmt.Errorf("assistants initialization: %w", err)
	}

	Assistants = cm

	shutdown := func() {
		utils.Log.Info("Stopping bot...")
		Bot.Close()

		utils.Log.Info("Shutting down assistants...")
		Assistants.ForEach(func(a *Assistant) {
			a.Ntg.Close()
			a.Client.Stop()
		})

		utils.Log.Info("Shutdown complete.")
	}

	return shutdown, nil
}

func initBot() error {
	opts := td.DefaultClientConfig()
	opts.LibraryPath = "./libtdjson.so.1.8.69"
	opts.ParseMode = td.ParseModeHTML
	opts.PanicHandler = clientPanicHandler
	opts.ErrorHandler = clientErrorHandler
	opts.AutoRetry = &td.AutoRetry{
		ChatNotFound:    true,
		MessageNotFound: true,
		MaxFloodWait:    15 * time.Second,
	}
	opts.DatabaseDirectory = "database"
	opts.Logger = utils.Log

	client, err := td.NewClient(config.APIID, config.APIHash, config.Token, opts)
	if err != nil {
		return fmt.Errorf("failed to create bot client: %w", err)
	}

	err = client.Start()
	if err != nil {
		return fmt.Errorf("failed to start bot: %w", err)
	}
	Bot = client

	if config.LoggerID != 0 {
		_, _ = client.SendMessage(config.LoggerID, "Bot Started", nil)
	}

	utils.Log.Infof("Bot started as @%s", botUsername())

	return nil
}

// clientPanicHandler marshals the raw update that caused a handler panic and
// forwards it to the logger chat together with the panic value and stack trace.
func clientPanicHandler(c *td.Client, update td.TlObject, r any) {
	sendErrorReport(c, "panic", update, r, string(debug.Stack()))
}

// clientErrorHandler does the same for handler errors, then lets the library
// continue dispatching the remaining handlers.
func clientErrorHandler(c *td.Client, update td.TlObject, err error) error {
	sendErrorReport(c, "error", update, err, "")
	return nil
}

func sendErrorReport(c *td.Client, kind string, update td.TlObject, detail any, stack string) {
	if c == nil || config.LoggerID == 0 {
		return
	}

	var raw string
	if update != nil {
		if b, err := json.MarshalIndent(update, "", " "); err == nil {
			raw = string(b)
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "<b>Bot %s:</b> %v\n\n", strings.ToUpper(kind), detail)

	if raw != "" {
		b.WriteString("<pre>")
		b.WriteString("Raw update:\n")
		b.WriteString(raw)
		b.WriteString("</pre>\n\n")
	}

	if stack != "" {
		b.WriteString("<pre>")
		b.WriteString("Stack trace:\n")
		b.WriteString(stack)
		b.WriteString("</pre>")
	}

	text := b.String()

	go func() {
		defer func() { _ = recover() }()
		if len(text) > 4000 {
			sendErrorReportFile(c, text)
			return
		}
		_, _ = c.SendMessage(config.LoggerID, text, &td.SendMessageOpts{
			ParseMode: "HTML",
		})
	}()
}

func sendErrorReportFile(c *td.Client, text string) {
	f, err := os.CreateTemp("", "bot_error_*.txt")
	if err != nil {
		return
	}
	defer os.Remove(f.Name())
	defer f.Close()

	if _, err := f.WriteString(text); err != nil {
		return
	}

	_, _ = c.SendMessageContent(config.LoggerID, &td.InputMessageDocument{
		Document: &td.InputDocument{Document: &td.InputFileLocal{Path: f.Name()}},
	}, nil)
}

func initAssistants(cm *ClientManager) error {
	for i, sessionStr := range config.StringSessions {
		utils.Log.Infof("Initializing assistant[%d]...", i)

		assistant, err := cm.AddClient(sessionStr)
		if err != nil {
			return fmt.Errorf("failed to initialize assistant[%d]: %w", i, err)
		}

		if config.LoggerID != 0 {
			_, _ = assistant.Client.SendMessage(
				config.LoggerID,
				fmt.Sprintf("Assistant %d Started", i+1),
			)
		}

		m, _ := assistant.Client.SendMessage(botUsername(), "/start")
		if m != nil {
			_, _ = m.Delete()
		}

		if assistant.Self.Username != "" {
			utils.Log.Infof(
				"Assistant[%d] started as @%s",
				i,
				assistant.Self.Username,
			)
		} else {
			utils.Log.Infof("Assistant[%d] started as %s", i, assistant.Self.FirstName)
		}
	}

	return nil
}
