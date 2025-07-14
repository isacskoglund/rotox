// Package config provides configuration loading and logger setup utilities.
//
// This package handles standardized logger creation with configurable levels
// and formats, and includes tracing integration for observability.
package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/isacskoglund/rotox/internal/tracing"
)

// LoggerFromConfig creates a configured slog.Logger based on the provided level and format.
// Supported levels: debug, info, warn, error
// Supported formats: json, text
// The logger includes tracing integration for request correlation.
func LoggerFromConfig(
	level string,
	format string,
) (*slog.Logger, error) {
	lvl, ok := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}[level]
	if !ok {
		return nil, fmt.Errorf("invalid log level: %s", level)
	}

	var handler slog.Handler
	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	case "text":
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	default:
		return nil, fmt.Errorf("invalid log format: %s", format)
	}
	handler = tracing.NewHandler(handler)
	return slog.New(handler), nil
}
