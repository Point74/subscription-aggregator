package logger

import (
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(handler)
}
