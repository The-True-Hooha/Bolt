package logger

import (
	"log/slog"
	"os"
)

var (
	handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	Logger = slog.New(handler)
)


func Debug(msg string, args ...any) {
	Logger.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	Logger.Info(msg, args...)
}


func Warn(msg string, args ...any) {
	Logger.Warn(msg, args...)
}


func Error(msg string, args ...any) {
	Logger.Error(msg, args...)
}