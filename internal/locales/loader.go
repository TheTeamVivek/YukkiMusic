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
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"main/internal/config"
)

//go:embed *.yml
var localesFS embed.FS

var loadedLocales = make(map[string]map[string]string)

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

func Get(lang, key string, values Arg) string {
	locale, ok := loadedLocales[lang]
	if !ok {
		locale = loadedLocales[config.DefaultLang]
	}

	val, ok := locale[key]
	if !ok {
		val = loadedLocales[config.DefaultLang][key]
	}

	if values == nil {
		return val
	}

	for k, v := range values {
		val = strings.ReplaceAll(val, "{"+k+"}", fmt.Sprint(v))
	}

	return val
}

func GetAvailableLanguages() []string {
	langs := make([]string, 0, len(loadedLocales))

	for lang := range loadedLocales {
		langs = append(langs, lang)
	}

	sort.Strings(langs)
	return langs
}
