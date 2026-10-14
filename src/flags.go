package main

import (
	"flag"
	"log/slog"
	"os"
	"strings"
)

type Flags struct {
	action        string
	keywords      string
	logLevel      string
	scrollSeconds int
	delayMs       int
}

func NewFlags() Flags {
	action := flag.String("action", "search", "Action to perform: search, process_data, get_details, save")
	keywords := flag.String("keywords", "", "Keywords to search for (required when -action is search)")
	logLevel := flag.String("log-level", "debug", "Log level: debug, info, warn, error")
	scrollSeconds := flag.Int("scroll-seconds", 60, "Max seconds to scroll search results before opening a product")
	delayMs := flag.Int("delay-ms", 0, "Delay in milliseconds between replayed description requests (action=pull_description)")
	flag.Parse()

	if strings.ToLower(*action) == "search" && *keywords == "" {
		LogFatal("keywords flag is required when action is search")
	}

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
		action:        *action,
		keywords:      *keywords,
		logLevel:      *logLevel,
		scrollSeconds: *scrollSeconds,
		delayMs:       *delayMs,
	}
}
