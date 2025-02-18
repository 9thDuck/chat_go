package env

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

func GetEnvString(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func GetEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		val, err := strconv.Atoi(val)
		if err != nil {
			slog.Error(fmt.Sprintf("error getting val for key: \"%s\" from .env, err: %v", key, err))
			return fallback
		}
		return val
	}
	return fallback
}
