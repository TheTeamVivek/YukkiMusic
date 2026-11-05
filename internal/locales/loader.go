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

package locales

import (
	"embed"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

	"yukkimusic/internal/utils"

	"gopkg.in/yaml.v3"

	"yukkimusic/config"
)

//go:embed *.yml
var localesFS embed.FS

var (
	loadedLocales = make(map[string]map[string]string)
	placeholderRe = regexp.MustCompile(`\{(\w+)\}`)
)

const baseLang = "en"

type Arg map[string]any

func Load() error {
	files, err := localesFS.ReadDir(".")
	if err != nil {
		return fmt.Errorf("read locales dir: %w", err)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		lang := strings.TrimSuffix(f.Name(), path.Ext(f.Name()))

		data, err := localesFS.ReadFile(f.Name())
		if err != nil {
			return fmt.Errorf("read locale %s: %w", f.Name(), err)
		}

		var locale map[string]string
		if err := yaml.Unmarshal(data, &locale); err != nil {
			return fmt.Errorf("parse locale %s: %w", f.Name(), err)
		}

		loadedLocales[lang] = locale
	}

	if _, ok := loadedLocales[config.DefaultLang]; !ok {
		return fmt.Errorf("default language %s not found", config.DefaultLang)
	}

	return nil
}

func lookup(lang, key string) (string, bool) {
	for _, l := range []string{lang, config.DefaultLang, baseLang} {
		if val, ok := loadedLocales[l][key]; ok {
			return val, true
		}
	}

	utils.Log.Errorf("Missing translation → lang=%s key=%s", lang, key)
	return "", false
}

func fill(s string, values Arg) string {
	if len(values) == 0 {
		return s
	}
	return placeholderRe.ReplaceAllStringFunc(s, func(m string) string {
		if v, ok := values[m[1:len(m)-1]]; ok {
			return fmt.Sprint(v)
		}
		return m
	})
}

func Get(lang, key string, values Arg) string {
	val, ok := lookup(lang, key)
	if !ok {
		return fmt.Sprintf("[%s]", key)
	}
	return fill(expandEmojis(val), values)
}

func GetButton(lang, key string, values Arg) (string, int64) {
	val, ok := lookup(lang, key)
	if !ok {
		return fmt.Sprintf("[%s]", key), 0
	}
	text, icon := buttonEmoji(val)
	return fill(text, values), icon
}

func GetAvailableLanguages() []string {
	langs := make([]string, 0, len(loadedLocales))

	for lang := range loadedLocales {
		langs = append(langs, lang)
	}

	sort.Strings(langs)
	return langs
}
