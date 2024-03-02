package common

import (
	"os"
	"time"

	"log/slog"
)

//nolint:gochecknoglobals // protected by package
var l *slog.Logger

// GetLogger returns the configured logger
func GetLogger() *slog.Logger {
	if l == nil {
		opts := &slog.HandlerOptions{
			// Use the ReplaceAttr function on the handler options to reformat timestamps
			ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
				// check that we are handling the time key
				if a.Key != slog.TimeKey {
					return a
				}
				t := a.Value.Time()
				a.Value = slog.StringValue(t.UTC().Format(time.DateTime))
				a.Key = "ts"
				return a
			},
		}
		l = slog.New(slog.NewTextHandler(os.Stdout, opts))
		level := slog.LevelInfo
		if os.Getenv("DEBUG") != "" {
			level = slog.LevelDebug
		}
		slog.SetLogLoggerLevel(level)
		slog.SetDefault(l)
	}
	return l
}
