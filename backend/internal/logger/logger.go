package logger

import (
	"log/slog"
	"os"
	"strings"
)

func New(appEnv string) *slog.Logger {
	level := slog.LevelInfo
	if strings.EqualFold(appEnv, "development") {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
