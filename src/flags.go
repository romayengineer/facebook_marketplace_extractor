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
	titleKeywords string
}

func NewFlags() Flags {
	action := flag.String("action", "search", "Action to perform: search, process_data, get_details, save")
	keywords := flag.String("keywords", "", "Keywords to search for (required when -action is search)")
	titleKeywords := flag.String("title-keywords", "", "Keywords to pull descriptions from (required when -action is pull_description)")
	logLevel := flag.String("log-level", "debug", "Log level: debug, info, warn, error")
	flag.Parse()

	if strings.ToLower(*action) == "search" && *keywords == "" {
		LogFatal("keywords flag is required when action is search")
	}

	if strings.ToLower(*action) == "pull_description" && *titleKeywords == "" {
		LogFatal("title-keywords flag is required when action is pull_description")
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
		titleKeywords: *titleKeywords,
	}
}
