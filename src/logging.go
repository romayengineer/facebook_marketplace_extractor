package main

import (
	"context"
	"log/slog"
)

const (
	LD0 slog.Level = -4
	LD1 slog.Level = -3
	LD2 slog.Level = -2
	LD3 slog.Level = -1
	LI0 slog.Level = 0
	LI1 slog.Level = 1
	LI2 slog.Level = 2
	LI3 slog.Level = 3
	LW0 slog.Level = 4
	LE0 slog.Level = 8
)

func Log(level slog.Level, msg string, args ...any) {
	slog.Log(context.Background(), level, msg, args...)
}
