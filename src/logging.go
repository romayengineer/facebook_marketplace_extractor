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

func LogDebug0(msg string, args ...any) {
	slog.Log(context.Background(), LD0, msg, args...)
}

func LogDebug1(msg string, args ...any) {
	slog.Log(context.Background(), LD1, msg, args...)
}

func LogDebug2(msg string, args ...any) {
	slog.Log(context.Background(), LD2, msg, args...)
}

func LogDebug3(msg string, args ...any) {
	slog.Log(context.Background(), LD3, msg, args...)
}

func LogInfo0(msg string, args ...any) {
	slog.Log(context.Background(), LI0, msg, args...)
}

func LogInfo1(msg string, args ...any) {
	slog.Log(context.Background(), LI1, msg, args...)
}

func LogInfo2(msg string, args ...any) {
	slog.Log(context.Background(), LI2, msg, args...)
}

func LogInfo3(msg string, args ...any) {
	slog.Log(context.Background(), LI3, msg, args...)
}

func LogWarn0(msg string, args ...any) {
	slog.Log(context.Background(), LW0, msg, args...)
}

func LogError0(msg string, args ...any) {
	slog.Log(context.Background(), LE0, msg, args...)
}
