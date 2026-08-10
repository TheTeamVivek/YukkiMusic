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

package modules

import (
	"bytes"
	"context"
	"fmt"
	"go/build"
	"io"
	"os"
	"reflect"
	"strings"

	td "github.com/AshokShau/gotdbot"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"yukkimusic/internal/logger"

	"yukkimusic/config"
	"yukkimusic/internal/core"
)

func evalCommandHandler(c *td.Client, m *td.Message) error {
	if m.SenderID() != config.OwnerID {
		return nil
	}

	html := &td.SendTextMessageOpts{ParseMode: "HTML"}
	parts := strings.SplitN(m.Text(), " ", 2)
	var code string
	if len(parts) > 1 {
		code = strings.TrimSpace(parts[1])
	}

	// Minimal help
	if strings.Contains(code, "--help") || strings.Contains(code, "-h") {
		_, err := m.ReplyText(c, `<b>🧩 Eval Help</b>

<code>/eval &lt;Go code&gt;</code>
• Run Go code dynamically.
• If your code has "package" or "func", it runs as-is.
• Otherwise it runs inside a helper func with:
  <pre>M, R, Client, UBot, Ntg</pre>
• Supports prints and returns.

Examples:
<pre>/eval fmt.Println("Hi")
/eval return 5</pre>`, html)
		return err
	}

	if code == "" {
		_, err := m.ReplyText(
			c,
			"No code provided.\nUse <code>/eval --help</code> for usage info.",
			html,
		)
		return err
	}

	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		goPath = build.Default.GOPATH
	}
	var stdout, stderr bytes.Buffer
	i := interp.New(interp.Options{
		Stdout: &stdout,
		Stderr: &stderr,
		GoPath: goPath,
	})
	i.Use(stdlib.Symbols)

	var reply *td.Message
	if m.ReplyToMessageID() != 0 {
		reply, _ = m.GetRepliedMessage(c)
	}

	symbols := map[string]map[string]reflect.Value{
		"eval/eval": {
			"M":          reflect.ValueOf(m),
			"Client":     reflect.ValueOf(c),
			"Assistants": reflect.ValueOf(core.Assistants),
			"A":          reflect.ValueOf(core.Assistants),
			"R":          reflect.ValueOf(reply),
			"Message":    reflect.ValueOf(m),
		},
	}
	if err := i.Use(symbols); err != nil {
		logger.Errorf("failed to use custom symbols: %v", err)
	}
	ctx := context.Background()

	// Wrap snippet mode
	if !strings.Contains(code, "package ") && !strings.Contains(code, "func ") {
		code = fmt.Sprintf(`package main
import (
	e "eval/eval"
	"fmt"
)

func runSnippet() (res any) {
	m, msg, message, M := e.M, e.M, e.M, e.M
	r := e.R
	client, c, app, bot, Client := e.Client, e.Client, e.Client, e.Client, e.Client
	a, ass, Assistants, A := e.A, e.A, e.A, e.A

	_ = m; _ = msg; _ = message; _ = M
	_ = r; _ = client; _ = c; _ = app; _ = bot; _ = Client
	_ = a; _ = ass; _ = Assistants; _ = A
	_ = fmt.Println

	%s

	return res
}

func main() {
	if res := runSnippet(); res != nil {
		fmt.Println(res)
	}
}`, code)
	}

	result, err := i.EvalWithContext(ctx, code)
	if err != nil {
		_, err := m.ReplyText(c, fmt.Sprintf("<b>#EVALERR:</b> <code>%s</code>", err.Error()), html)
		return err
	}

	var output string
	if stdout.Len() > 0 {
		output = stdout.String()
	}
	if stderr.Len() > 0 {
		output += "\n" + stderr.String()
	}

	if result.IsValid() && result.Kind() != reflect.Invalid {
		val := result.Interface()
		rv := reflect.ValueOf(val)
		for rv.Kind() == reflect.Pointer {
			if rv.IsNil() {
				break
			}
			rv = rv.Elem()
		}
		if rv.IsValid() && rv.Interface() != nil {
			outVal := fmt.Sprintf("%v", rv.Interface())
			if output != "" {
				output += "\n"
			}
			output += fmt.Sprintf("Output: %s", outVal)
		}
	}

	if strings.TrimSpace(output) == "" {
		output = "<code>No Output</code>"
	}

	if len(output) > 4095 {
		file, _ := os.Create("output.txt")
		defer file.Close()
		io.WriteString(file, output)
		if _, rerr := m.ReplyDocument(
			c,
			&td.InputFileLocal{Path: file.Name()},
			&td.SendDocumentOpts{Caption: "Output"},
		); rerr != nil {
			return rerr
		}
		os.Remove(file.Name())
		return nil
	}
	_, rerr := m.ReplyText(
		c,
		fmt.Sprintf(
			"<b>#EVALOut:</b>\n<code>%s</code>",
			strings.TrimSpace(output),
		),
		html,
	)
	return rerr
}
