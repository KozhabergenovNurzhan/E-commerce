package logger

import (
	"log/slog"
	"os"
)

func New(service string, level slog.Level) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})
	return slog.New(handler).With(slog.String("service", service))
}
