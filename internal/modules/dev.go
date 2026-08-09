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
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	td "github.com/AshokShau/gotdbot"

	"yukkimusic/config"
)

func init() {
	helpTexts["/sh"] = `<i>Execute shell commands on server.</i>

<u>Usage:</u>
<b>/sh [command]</b> — Run shell command

<b>🔒 Restrictions:</b>
• <b>Owner only</b> command

<b>⚠️ Warning:</b>
Direct system access - extremely powerful.`
	helpTexts["/bash"] = helpTexts["/sh"]
	helpTexts["/shell"] = helpTexts["/sh"]

	helpTexts["/eval"] = helpTexts["/ev"]
	helpTexts["/ev"] = `<i>Execute Go code dynamically (eval mode).</i>

<u>Usage:</u>
<b>/eval [code]</b> — Run Go code

<b>🔒 Restrictions:</b>
• <b>Owner only</b> command

<b>⚠️ Warning:</b>
Powerful command - use with caution.`

	helpTexts["/json"] = `<i>Get JSON representation of message/user/chat.</i>

<u>Usage:</u>
<b>/json</b> — Current message JSON
<b>/json -s</b> — Sender JSON
<b>/json -c</b> — Chat JSON
<b>/json -m</b> — Media JSON
<b>/json [reply] -f</b> — File JSON

<b>💡 Use Case:</b>
Debugging and development.`

	helpTexts["/logs"] = `<i>Send current bot logs file.</i>

<u>Usage:</u>
<b>/logs</b> — Send current log file

<b>🔒 Restrictions:</b>
• <b>Sudo users only</b>`
}

func logsHandler(c *td.Client, m *td.Message) error {
	chatID := m.ChatID()
	logFile := "logs.txt"

	info, err := os.Stat(logFile)
	if err != nil || info.Size() == 0 {
		m.ReplyText(c, F(chatID, "logs_empty"), nil)
		return nil
	}

	_, err = m.ReplyDocument(c, &td.InputFileLocal{Path: logFile}, nil)
	if err != nil {
		m.ReplyText(c, err.Error(), nil)
	}

	return nil
}

func shellHandler(c *td.Client, m *td.Message) error {
	if m.SenderID() != config.OwnerID {
		return nil
	}
	cmd := m.Args()
	var cmdArgs []string
	if cmd == "" {
		m.ReplyText(c, "No command provided", nil)
		return nil
	}

	if runtime.GOOS == "windows" {
		cmd = "cmd"
		cmdArgs = append([]string{"/C"}, strings.Split(m.Args(), " ")...)
	} else {
		parts := strings.Split(cmd, " ")
		cmd = parts[0]
		cmdArgs = parts[1:]
	}
	cmx := exec.Command(cmd, cmdArgs...)
	var out bytes.Buffer
	cmx.Stdout = &out
	var errx bytes.Buffer
	cmx.Stderr = &errx
	err := cmx.Run()

	html := &td.SendTextMessageOpts{ParseMode: "HTML"}

	if errx.String() == "" && out.String() == "" {
		if err != nil {
			m.ReplyText(c, "<code>Error:</code> <b>"+err.Error()+"</b>", html)
			return nil
		}
		m.ReplyText(c, "<code>No Output</code>", html)
		return nil
	}

	if out.String() != "" {
		m.ReplyText(
			c,
			`<pre lang="bash">`+strings.TrimSpace(out.String())+`</pre>`,
			html,
		)
	} else {
		m.ReplyText(c, `<pre lang="bash">`+strings.TrimSpace(errx.String())+`</pre>`, html)
	}
	return nil
}

func jsonHandle(c *td.Client, m *td.Message) error {
	var jsonString []byte
	html := &td.SendTextMessageOpts{ParseMode: "HTML"}
	if m.ReplyToMessageID() == 0 {
		switch {
		case strings.Contains(m.Args(), "-s"):
			if u, err := c.GetUser(m.SenderID()); err == nil && u != nil {
				jsonString, _ = json.MarshalIndent(u, "", "  ")
			}
		case strings.Contains(m.Args(), "-m"):
			jsonString, _ = json.MarshalIndent(m.Content, "", "  ")
		case strings.Contains(m.Args(), "-c"):
			if ch, err := c.GetChat(m.ChatID()); err == nil && ch != nil {
				jsonString, _ = json.MarshalIndent(ch, "", "  ")
			}
		default:
			jsonString, _ = json.MarshalIndent(m, "", "  ")
		}
	} else {
		r, err := m.GetRepliedMessage(c)
		if err != nil || r == nil {
			m.ReplyText(c, "<code>Error:</code> <b>could not fetch replied message</b>", html)
			return nil
		}
		switch {
		case strings.Contains(m.Args(), "-s"):
			if u, err := c.GetUser(r.SenderID()); err == nil && u != nil {
				jsonString, _ = json.MarshalIndent(u, "", "  ")
			}
		case strings.Contains(m.Args(), "-m"):
			jsonString, _ = json.MarshalIndent(r.Content, "", "  ")
		case strings.Contains(m.Args(), "-c"):
			if ch, err := c.GetChat(r.ChatID()); err == nil && ch != nil {
				jsonString, _ = json.MarshalIndent(ch, "", "  ")
			}
		case strings.Contains(m.Args(), "-f"):
			jsonString, _ = json.MarshalIndent(r.Content, "", "  ")
		default:
			jsonString, _ = json.MarshalIndent(r, "", "  ")
		}
	}

	// find all "Data": "<base64>" and decode and replace with actual data
	dataFieldRegex := regexp.MustCompile(`"Data": "([a-zA-Z0-9+/]+={0,2})"`)
	dataFields := dataFieldRegex.FindAllStringSubmatch(string(jsonString), -1)
	for _, v := range dataFields {
		decoded, err := base64.StdEncoding.DecodeString(v[1])
		if err != nil {
			m.ReplyText(c, "Error: "+err.Error(), html)
			return nil
		}
		jsonString = []byte(
			strings.ReplaceAll(
				string(jsonString),
				v[0],
				`"Data": "`+string(decoded)+`"`,
			),
		)
	}

	if len(jsonString) > 4095 {
		defer os.Remove("message.json")
		tmpFile, err := os.Create("message.json")
		if err != nil {
			m.ReplyText(c, "Error: "+err.Error(), html)
			return nil
		}

		_, err = tmpFile.Write(jsonString)
		if err != nil {
			m.ReplyText(c, "Error: "+err.Error(), html)
			return nil
		}

		_, err = m.ReplyDocument(
			c,
			&td.InputFileLocal{Path: tmpFile.Name()},
			&td.SendDocumentOpts{Caption: "Message JSON"},
		)
		if err != nil {
			m.ReplyText(c, "Error: "+err.Error(), html)
		}
	} else {
		m.ReplyText(c, "<pre lang='json'>"+string(jsonString)+"</pre>", html)
	}

	return nil
}
