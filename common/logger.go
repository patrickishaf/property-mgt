package common

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

func CreateLogger() *slog.Logger {
	l := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(l)
	logger = l
	return logger
}
