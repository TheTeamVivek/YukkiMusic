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
 * This program is distributed in the hope that it will be useful,
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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
)

func shellHandle(m *telegram.NewMessage) error {
	if m.SenderID() != config.OwnerID {
		return telegram.EndGroup
	}
	cmd := m.Args()
	var cmd_args []string
	if cmd == "" {
		m.Reply("No command provided")
		return nil
	}

	if runtime.GOOS == "windows" {
		cmd = "cmd"
		cmd_args_b := strings.Split(m.Args(), " ")
		cmd_args = []string{"/C"}
		cmd_args = append(cmd_args, cmd_args_b...)
	} else {
		cmd = strings.Split(cmd, " ")[0]
		cmd_args = strings.Split(m.Args(), " ")
		cmd_args = append(cmd_args[:0], cmd_args[1:]...)
	}
	cmx := exec.Command(cmd, cmd_args...)
	var out bytes.Buffer
	cmx.Stdout = &out
	var errx bytes.Buffer
	cmx.Stderr = &errx
	err := cmx.Run()

	if errx.String() == "" && out.String() == "" {
		if err != nil {
			m.Reply("<code>Error:</code> <b>" + err.Error() + "</b>")
			return nil
		}
		m.Reply("<code>No Output</code>")
		return nil
	}

	if out.String() != "" {
		m.Reply(`<pre lang="bash">` + strings.TrimSpace(out.String()) + `</pre>`)
	} else {
		m.Reply(`<pre lang="bash">` + strings.TrimSpace(errx.String()) + `</pre>`)
	}
	return nil
}

// --------- Eval function ------------

const boiler_code_for_eval = `
package main

import "fmt"
import "github.com/amarnathcjd/gogram/telegram"
import "encoding/json"

%s

var msg_id int32 = %d

var client *telegram.Client
var ub *telegram.Client
var message *telegram.NewMessage
var m *telegram.NewMessage
var r *telegram.NewMessage
` + "var msg = `%s`\nvar snd = `%s`\nvar cht = `%s`\nvar chn = `%s`\nvar cch = `%s`" + `


func evalCode() {
        %s
}

func main() {
        var msg_o *telegram.MessageObj
        var snd_o *telegram.UserObj
        var cht_o *telegram.ChatObj
        var chn_o *telegram.Channel
        json.Unmarshal([]byte(msg), &msg_o)
        json.Unmarshal([]byte(snd), &snd_o)
        json.Unmarshal([]byte(cht), &cht_o)
        json.Unmarshal([]byte(chn), &chn_o)
        client, _ = telegram.NewClient(telegram.ClientConfig{
                StringSession: "%s",
        })

        client.Cache.ImportJSON([]byte(cch))

        client.Conn()
        ub, _ = telegram.NewClient(telegram.ClientConfig{
                StringSession: "%s",
        })

        ub.Conn()
        
        x := []telegram.User{}
        y := []telegram.Chat{}
        x = append(x, snd_o)
        if chn_o != nil {
                y = append(y, chn_o)
        }
        if cht_o != nil {
                y = append(y, cht_o)
        }
        client.Cache.UpdatePeersToCache(x, y)
        idx := 0
        if cht_o != nil {
                idx = int(cht_o.ID)
        }
        if chn_o != nil {
                idx = int(chn_o.ID)
        }
        if snd_o != nil && idx == 0 {
                idx = int(snd_o.ID)
        }

        messageX, err := client.GetMessages(idx, &telegram.SearchOption{
                IDs: int(msg_id),
        })

        if err != nil {
                fmt.Println(err)
        }

        message = &messageX[0]
        m = message
        r, _ = message.GetReplyMessage()

        fmt.Println("output-start")
        evalCode()
}

func packMessage(c *telegram.Client, message telegram.Message, sender *telegram.UserObj, channel *telegram.Channel, chat *telegram.ChatObj) *telegram.NewMessage {
        var (
                m = &telegram.NewMessage{}
        )
        switch message := message.(type) {
        case *telegram.MessageObj:
                m.ID = message.ID
                m.OriginalUpdate = message
                m.Message = message
                m.Client = c
        default:
                return nil
        }
        m.Sender = sender
        m.Chat = chat
        m.Channel = channel
        if m.Channel != nil && (m.Sender.ID == m.Channel.ID) {
                m.SenderChat = channel
        } else {
                m.SenderChat = &telegram.Channel{}
        }
        m.Peer, _ = c.GetSendablePeer(message.(*telegram.MessageObj).PeerID)

        /*if m.IsMedia() {
                FileID := telegram.PackBotFileID(m.Media())
                m.File = &telegram.CustomFile{
                        FileID: FileID,
                        Name:   getFileName(m.Media()),
                        Size:   getFileSize(m.Media()),
                        Ext:    getFileExt(m.Media()),
                }
        }*/
        return m
}
`

func resolveImports(code string) (string, []string) {
	var imports []string
	importsRegex := regexp.MustCompile(`import\s*\(([\s\S]*?)\)|import\s*\"([\s\S]*?)\"`)
	importsMatches := importsRegex.FindAllStringSubmatch(code, -1)
	for _, v := range importsMatches {
		if v[1] != "" {
			imports = append(imports, v[1])
		} else {
			imports = append(imports, v[2])
		}
	}
	code = importsRegex.ReplaceAllString(code, "")
	return code, imports
}

func evalHandle(m *telegram.NewMessage) error {
	if m.SenderID() != config.OwnerID {
		return telegram.EndGroup
	}
	code := ""
	if x := strings.Split(m.RawText(true), " "); len(x) < 2 {
		return telegram.EndGroup
	} else {
		code = strings.TrimSpace(strings.Join(x[1:], " "))
	}

	code, imports := resolveImports(code)

	if code == "" {
		return nil
	}

	defer os.Remove("tmp/eval.go")
	defer os.Remove("tmp/eval_out.txt")
	defer os.Remove("tmp")

	resp, isfile := performEval(code, m, imports)
	if isfile {
		if _, err := m.ReplyMedia(resp, &telegram.MediaOptions{Caption: "Output"}); err != nil {
			m.Reply("Error: " + err.Error())
		}
		return nil
	}
	resp = strings.TrimSpace(resp)

	if resp != "" {
		if _, err := m.Reply(resp); err != nil {
			m.Reply(err)
		}
	}
	return nil
}

func performEval(code string, m *telegram.NewMessage, imports []string) (string, bool) {
	msg_b, _ := json.Marshal(m.Message)
	snd_b, _ := json.Marshal(m.Sender)
	cnt_b, _ := json.Marshal(m.Chat)
	chn_b, _ := json.Marshal(m.Channel)
	cache_b, _ := m.Client.Cache.ExportJSON()
	var importStatement string = ""
	if len(imports) > 0 {
		importStatement = "import (\n"
		for _, v := range imports {
			importStatement += `"` + v + `"` + "\n"
		}
		importStatement += ")\n"
	}
	ass, aErr := core.Assistants.First()
	if aErr != nil {
		return fmt.Sprintf("Failed to get assistant: %v", aErr), false
	}
	code_file := fmt.Sprintf(boiler_code_for_eval, importStatement, m.ID, msg_b, snd_b, cnt_b, chn_b, cache_b, code, m.Client.ExportSession(), ass.Client.ExportSession())
	tmp_dir := "tmp"
	_, err := os.ReadDir(tmp_dir)
	if err != nil {
		err = os.Mkdir(tmp_dir, 0o755)
		if err != nil {
			fmt.Println(err)
		}
	}

	// defer os.Remove(tmp_dir)

	os.WriteFile(tmp_dir+"/eval.go", []byte(code_file), 0o644)
	cmd := exec.Command("go", "run", "tmp/eval.go")
	var stdOut bytes.Buffer
	cmd.Stdout = &stdOut
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr

	err = cmd.Run()
	if stdOut.String() == "" && stdErr.String() == "" {
		if err != nil {
			return fmt.Sprintf("<b>#EVALERR:</b> <code>%s</code>", err.Error()), false
		}
		return "<b>#EVALOut:</b> <code>No Output</code>", false
	}

	if stdOut.String() != "" {
		if len(stdOut.String()) > 4095 {
			os.WriteFile("tmp/eval_out.txt", stdOut.Bytes(), 0o644)
			return "tmp/eval_out.txt", true
		}

		strDou := strings.Split(stdOut.String(), "output-start")

		return fmt.Sprintf("<b>#EVALOut:</b> <code>%s</code>", strings.TrimSpace(strDou[1])), false
	}

	if stdErr.String() != "" {
		regexErr := regexp.MustCompile(`eval.go:\d+:\d+:`)
		errMsg := regexErr.Split(stdErr.String(), -1)
		if len(errMsg) > 1 {
			if len(errMsg[1]) > 4095 {
				os.WriteFile("tmp/eval_out.txt", []byte(errMsg[1]), 0o644)
				return "tmp/eval_out.txt", true
			}
			return fmt.Sprintf("<b>#EVALERR:</b> <code>%s</code>", strings.TrimSpace(errMsg[1])), false
		}
		return fmt.Sprintf("<b>#EVALERR:</b> <code>%s</code>", stdErr.String()), false
	}

	return "<b>#EVALOut:</b> <code>No Output</code>", false
}

func jsonHandle(m *telegram.NewMessage) error {
	var jsonString []byte
	if !m.IsReply() {
		if strings.Contains(m.Args(), "-s") {
			jsonString, _ = json.MarshalIndent(m.Sender, "", "  ")
		} else if strings.Contains(m.Args(), "-m") {
			jsonString, _ = json.MarshalIndent(m.Media(), "", "  ")
		} else if strings.Contains(m.Args(), "-c") {
			jsonString, _ = json.MarshalIndent(m.Channel, "", "  ")
		} else {
			jsonString, _ = json.MarshalIndent(m.OriginalUpdate, "", "  ")
		}
	} else {
		r, err := m.GetReplyMessage()
		if err != nil {
			m.Reply("<code>Error:</code> <b>" + err.Error() + "</b>")
			return nil
		}
		if strings.Contains(m.Args(), "-s") {
			jsonString, _ = json.MarshalIndent(r.Sender, "", "  ")
		} else if strings.Contains(m.Args(), "-m") {
			jsonString, _ = json.MarshalIndent(r.Media(), "", "  ")
		} else if strings.Contains(m.Args(), "-c") {
			jsonString, _ = json.MarshalIndent(r.Channel, "", "  ")
		} else if strings.Contains(m.Args(), "-f") {
			jsonString, _ = json.MarshalIndent(r.File, "", "  ")
		} else {
			jsonString, _ = json.MarshalIndent(r.OriginalUpdate, "", "  ")
		}
	}

	// find all "Data": "<base64>" and decode and replace with actual data
	dataFieldRegex := regexp.MustCompile(`"Data": "([a-zA-Z0-9+/]+={0,2})"`)
	dataFields := dataFieldRegex.FindAllStringSubmatch(string(jsonString), -1)
	for _, v := range dataFields {
		decoded, err := base64.StdEncoding.DecodeString(v[1])
		if err != nil {
			m.Reply("Error: " + err.Error())
			return nil
		}
		jsonString = []byte(strings.ReplaceAll(string(jsonString), v[0], `"Data": "`+string(decoded)+`"`))
	}

	if len(jsonString) > 4095 {
		defer os.Remove("message.json")
		tmpFile, err := os.Create("message.json")
		if err != nil {
			m.Reply("Error: " + err.Error())
			return nil
		}

		_, err = tmpFile.Write(jsonString)
		if err != nil {
			m.Reply("Error: " + err.Error())
			return nil
		}

		_, err = m.ReplyMedia(tmpFile.Name(), &telegram.MediaOptions{Caption: "Message JSON"})
		if err != nil {
			m.Reply("Error: " + err.Error())
		}
	} else {
		m.Reply("<pre lang='json'>" + string(jsonString) + "</pre>")
	}

	return nil
}
