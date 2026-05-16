package main

import (
	"flag"
	"log/slog"
	"os"
	"strings"
)

type Flags struct {
	action   string
	logLevel string
}

func NewFlags() Flags {
	action := flag.String("action", "search", "Action to perform: search, process_data, get_details")
	logLevel := flag.String("log-level", "debug", "Log level: debug, info, warn, error")
	flag.Parse()

	var level slog.Level
	switch strings.ToLower(*logLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})))

	return Flags{
		action:   *action,
		logLevel: *logLevel,
	}
}
