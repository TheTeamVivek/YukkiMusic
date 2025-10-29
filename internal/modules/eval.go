/*
 * This file is part of YukkiMusic.
 *
 * YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
 * Copyright (C) 2025 TheTeamVivek
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it is useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 */

package modules

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"

	"github.com/TheTeamVivek/YukkiMusic/config"
	"github.com/TheTeamVivek/YukkiMusic/internal/core"
)

var evalLogger = gologging.GetLogger("Eval")

func evalCommandHandler(m *telegram.NewMessage) error {
	if m.SenderID() != config.OwnerID {
		return telegram.EndGroup
	}

	parts := strings.SplitN(m.RawText(true), " ", 2)
	var code string
	if len(parts) > 1 {
		code = strings.TrimSpace(parts[1])
	}

	// Minimal help
	if strings.Contains(code, "--help") || strings.Contains(code, "-h") {
		m.Reply(`<b>🧩 Eval Help</b>

<code>/eval &lt;Go code&gt;</code>
• Run Go code dynamically.
• If your code has "package" or "func", it runs as-is.
• Otherwise it runs inside a helper func with:
  <pre>M, R, Client, UBot, Ntg</pre>
• Supports prints and returns.

Examples:
<pre>/eval fmt.Println("Hi")
/eval return 5</pre>`)
		return nil
	}

	if code == "" {
		m.Reply("No code provided.\nUse <code>/eval --help</code> for usage info.")
		return nil
	}

	var stdout, stderr bytes.Buffer
	i := interp.New(interp.Options{
		Stdout: &stdout,
		Stderr: &stderr,
		GoPath: os.Getenv("GOPATH"),
		//	GoPath: build.Default.GOPATH,
	})
	i.Use(stdlib.Symbols)

	var reply *telegram.NewMessage
	if m.IsReply() {
		reply, _ = m.GetReplyMessage()
	}

	symbols := map[string]map[string]reflect.Value{
		"eval/eval": {
			"M":       reflect.ValueOf(m),
			"Client":  reflect.ValueOf(core.Bot),
			"UBot":    reflect.ValueOf(core.UBot),
			"R":       reflect.ValueOf(reply),
			"Message": reflect.ValueOf(m),
			"Ntg":     reflect.ValueOf(core.Ntg),
		},
	}
	if err := i.Use(symbols); err != nil {
		evalLogger.ErrorF("failed to use custom symbols: %v", err)
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
	call, ntg, Ntg := e.Ntg, e.Ntg, e.Ntg
	ub, UBot := e.UBot, e.UBot
	j := e.Client.JSON

	_ = m; _ = msg; _ = message; _ = M
	_ = r; _ = client; _ = c; _ = app; _ = bot; _ = Client
	_ = call; _ = ntg; _ = Ntg; _ = ub; _ = UBot
	_ = j; _ = fmt.Println

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
		m.Reply(fmt.Sprintf("<b>#EVALERR:</b> <code>%s</code>", err.Error()))
		return nil
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
		for rv.Kind() == reflect.Ptr {
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
		m.ReplyMedia(file.Name(), telegram.MediaOptions{Caption: "Output"})
		os.Remove(file.Name())
		return nil
	}

	m.Reply(fmt.Sprintf("<b>#EVALOut:</b>\n<code>%s</code>", strings.TrimSpace(output)))
	return nil
}
