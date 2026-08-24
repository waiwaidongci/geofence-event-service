package logging

import (
	"log/slog"
	"os"
)

func New(level string) *slog.Logger {
	var value slog.Level
	if level == "debug" {
		value = slog.LevelDebug
	} else if level == "warn" {
		value = slog.LevelWarn
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: value}))
}
