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

package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf16"

	td "github.com/AshokShau/gotdbot"
)

// ExtractURLs collects URLs from the given message and, if it is a reply,
// from the replied message too. Both message text and media captions are
// inspected, since Telegram entities cover both.
func ExtractURLs(c *td.Client, m *td.Message) ([]string, error) {
	if m == nil {
		return nil, fmt.Errorf("invalid message")
	}

	urls := collectURLs(m.Text(), m.Entities())
	urls = append(urls, collectURLs(m.Caption(), m.CaptionEntities())...)
	urls = append(urls, collectRichURLs(m)...)

	if m.ReplyToMessageID() <= 0 {
		return finalizeURLs(urls)
	}

	r, err := m.GetRepliedMessage(c)
	if err != nil {
		if len(urls) > 0 {
			return urls, fmt.Errorf("failed to fetch reply message: %w", err)
		}
		return nil, fmt.Errorf("failed to fetch reply message: %w", err)
	}

	urls = append(urls, collectURLs(r.Text(), r.Entities())...)
	urls = append(urls, collectURLs(r.Caption(), r.CaptionEntities())...)
	urls = append(urls, collectRichURLs(r)...)
	return finalizeURLs(urls)
}

// collectRichURLs extracts links from a rich message (messageRichMessage). Rich
// messages do not expose td.TextEntity values; their links are richTextUrl
// nodes nested inside page blocks (headings, paragraphs, tables, ...). The
// content is walked generically so every current and future block/rich-text
// type is covered, and only richTextUrl carries a URL.
func collectRichURLs(m *td.Message) []string {
	var msg *td.RichMessage
	switch content := m.Content.(type) {
	case td.MessageRichMessage:
		msg = content.Message
	case *td.MessageRichMessage:
		msg = content.Message
	}
	if msg == nil {
		return nil
	}

	var urls []string
	walkRichURLs(reflect.ValueOf(msg), &urls)
	return urls
}

var richTextURLType = reflect.TypeFor[td.RichTextUrl]()

func walkRichURLs(v reflect.Value, urls *[]string) {
	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsNil() {
			walkRichURLs(v.Elem(), urls)
		}
	case reflect.Interface:
		if !v.IsNil() {
			walkRichURLs(v.Elem(), urls)
		}
	case reflect.Struct:
		if v.Type() == richTextURLType {
			if u := v.FieldByName("Url").String(); u != "" {
				*urls = append(*urls, u)
			}
		}
		for _, field := range v.Fields() {
			walkRichURLs(field, urls)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			walkRichURLs(v.Index(i), urls)
		}
	}
}

func collectURLs(text string, entities []td.TextEntity) []string {
	urls := make([]string, 0, len(entities))

	for _, ent := range entities {
		switch e := ent.Type.(type) {
		case *td.TextEntityTypeUrl:
			if seg, ok := sliceUTF16(text, ent.Offset, ent.Length); ok {
				urls = append(urls, seg)
			}
		case *td.TextEntityTypeTextUrl:
			if e.Url != "" {
				urls = append(urls, e.Url)
			}
		}
	}
	return urls
}

// sliceUTF16 slices the given string using UTF-16 code unit offsets,
// which is what Telegram text entities use.
func sliceUTF16(s string, offset, length int32) (string, bool) {
	if offset < 0 || length < 0 {
		return "", false
	}
	u16 := utf16.Encode([]rune(s))
	start := int(offset)
	end := start + int(length)
	if end > len(u16) {
		return "", false
	}
	return string(utf16.Decode(u16[start:end])), true
}

func finalizeURLs(urls []string) ([]string, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("no URLs found")
	}

	seen := make(map[string]struct{}, len(urls))
	deduped := make([]string, 0, len(urls))
	for _, u := range urls {
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		deduped = append(deduped, u)
	}
	return deduped, nil
}

// ExtractUser extracts a user ID from a message.
// It supports replies, inline mentions, @username mentions and plain text IDs/usernames.
func ExtractUser(c *td.Client, m *td.Message) (int64, error) {
	if m == nil {
		return 0, fmt.Errorf("invalid message")
	}

	if m.ReplyToMessageID() > 0 {
		return extractUserFromReply(c, m)
	}

	uid, ok, err := extractUserFromEntities(c, m)
	if err != nil {
		return 0, err
	}
	if ok {
		return uid, nil
	}

	return extractUserFromPlainText(c, m.Text())
}

func extractUserFromReply(c *td.Client, m *td.Message) (int64, error) {
	r, err := c.GetMessage(m.ChatID(), m.ReplyToMessageID())
	if err != nil {
		return 0, fmt.Errorf("failed to fetch reply message: %w", err)
	}
	if r == nil || r.SenderId == nil {
		return 0, fmt.Errorf(
			"replied message's sender is not a user (may be anon admin)",
		)
	}
	u, ok := r.SenderId.(*td.MessageSenderUser)
	if !ok {
		return 0, fmt.Errorf(
			"replied message's sender is not a user (maybe channel/group)",
		)
	}
	return u.UserId, nil
}

func extractUserFromEntities(c *td.Client, m *td.Message) (int64, bool, error) {
	text := m.Text()
	for _, ent := range m.Entities() {
		switch e := ent.Type.(type) {
		case *td.TextEntityTypeMentionName:
			return e.UserId, true, nil
		case *td.TextEntityTypeMention:
			start := int(ent.Offset)
			end := start + int(ent.Length)
			if start < 0 || end > len(text) || start > end {
				continue
			}
			username := strings.TrimPrefix(text[start:end], "@")
			uid, err := resolveUserID(c, username)
			if err != nil {
				return 0, false, err
			}
			if uid == 0 {
				return 0, false, fmt.Errorf(
					"resolved peer is not a user (maybe channel/group)",
				)
			}
			return uid, true, nil
		}
	}
	return 0, false, nil
}

func extractUserFromPlainText(c *td.Client, text string) (int64, error) {
	parts := strings.Fields(text)
	if len(parts) < 2 {
		return 0, fmt.Errorf("no user identifier found")
	}

	idStr := strings.TrimPrefix(strings.TrimSpace(parts[1]), "@")
	if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
		return id, nil
	}

	uid, err := resolveUserID(c, idStr)
	if err != nil {
		return 0, err
	}
	if uid == 0 {
		return 0, fmt.Errorf(
			"resolved peer is not a user (maybe channel/group)",
		)
	}
	return uid, nil
}

// ExtractChat extracts a chat ID from a message.
// It supports plain text IDs/usernames and @username mentions.
func ExtractChat(c *td.Client, m *td.Message) (int64, error) {
	if m == nil {
		return 0, fmt.Errorf("invalid message")
	}

	parts := strings.Fields(m.Text())
	if len(parts) < 2 {
		return 0, fmt.Errorf("no chat identifier found")
	}

	target := strings.TrimPrefix(strings.TrimSpace(parts[1]), "@")
	if id, err := strconv.ParseInt(target, 10, 64); err == nil {
		return id, nil
	}

	chat, err := searchPublicChat(c, target)
	if err != nil {
		return 0, err
	}
	return chat.Id, nil
}

func resolveUserID(c *td.Client, username string) (int64, error) {
	chat, err := searchPublicChat(c, username)
	if err != nil {
		return 0, err
	}
	priv, ok := chat.Type.(*td.ChatTypePrivate)
	if !ok {
		return 0, nil
	}
	return priv.UserId, nil
}

func searchPublicChat(c *td.Client, username string) (*td.Chat, error) {
	chat, err := c.SearchPublicChat(username)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}
	if chat == nil {
		return nil, fmt.Errorf("resolved peer is not found")
	}
	return chat, nil
}
