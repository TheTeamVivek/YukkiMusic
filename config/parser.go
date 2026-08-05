package config

import (
	"os"
	"strconv"
	"strings"
)

func fallback[T any](def []T) T {
	if len(def) > 0 {
		return def[0]
	}

	var zero T
	return zero
}

func getEnv(key string, def ...string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback(def)
}

func getEnvBool(key string, def ...bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback(def)
	}

	b, err := strconv.ParseBool(v)
	if err != nil {
		logger.FatalF("invalid boolean value for %s: %q", key, v)
	}

	return b
}

func getEnvInt(key string, def ...int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback(def)
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		logger.FatalF("invalid integer value for %s: %q", key, v)
	}

	return n
}

func getEnvInt64(key string, def ...int64) int64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback(def)
	}

	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		logger.FatalF("invalid int64 value for %s: %q", key, v)
	}

	return n
}

func getEnvStrings(key string, def ...[]string) []string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback(def)
	}

	fields := strings.FieldsFunc(v, func(r rune) bool {
		switch r {
		case ',', ';', ' ':
			return true
		default:
			return false
		}
	})

	out := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field != "" {
			out = append(out, field)
		}
	}

	if len(out) == 0 {
		return fallback(def)
	}

	return out
}

func getEnvInt64s(key string, def ...[]int64) []int64 {
	fields := getEnvStrings(key)
	if len(fields) == 0 {
		return fallback(def)
	}

	out := make([]int64, 0, len(fields))
	for _, field := range fields {
		n, err := strconv.ParseInt(field, 10, 64)
		if err != nil {
			logger.WarnF("ignoring invalid %s value %q", key, field)
			continue
		}
		out = append(out, n)
	}

	if len(out) == 0 {
		return fallback(def)
	}

	return out
}
