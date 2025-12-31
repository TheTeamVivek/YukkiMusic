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
package locales

import (
	"embed"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/Laky-64/gologging"
	"gopkg.in/yaml.v3"

	"main/internal/config"
)

//go:embed *.yml
var locales embed.FS

var (
	loadedLocales = make(map[string]map[string]string)
)

type Arg map[string]any

func init() {
	files, err := locales.ReadDir(".")
	if err != nil {
		gologging.Fatal("Failed to read embedded locales:", err)
		return
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		lang := f.Name()[:len(f.Name())-len(path.Ext(f.Name()))]
		file, err := locales.ReadFile(f.Name())
		if err != nil {
			gologging.Fatal("Failed to read locale file:", f.Name(), err)
			continue
		}
		var locale map[string]string
		if err := yaml.Unmarshal(file, &locale); err != nil {
			gologging.Fatal("Failed to unmarshal locale file:", f.Name(), err)
			continue
		}
		loadedLocales[lang] = locale
	}
	if _, ok := loadedLocales[config.DefaultLang]; !ok {
		gologging.Fatal("Default language not found:", config.DefaultLang)
	}
	gologging.Info("Loaded", len(loadedLocales), "locales.")
}

func Get(lang, key string, values Arg) string {
	if _, ok := loadedLocales[lang]; !ok {
		lang = config.DefaultLang
	}

	val, ok := loadedLocales[lang][key]
	if !ok {
		val = loadedLocales[config.DefaultLang][key]
	}

	if values == nil {
		return val
	}

	var buf strings.Builder

	for k, v := range values {
		buf.Reset()
		fmt.Fprintf(&buf, "%v", v)
		val = strings.ReplaceAll(val, "{"+k+"}", buf.String())
	}

	return val
}

func GetAvailableLanguages() []string {
	var langs []string
	for lang := range loadedLocales {
		langs = append(langs, lang)
	}
	sort.Strings(langs)
	return langs
}
