package logger

import (
	"context"
	"os"
	"time"

	"log/slog"
)

//nolint:gochecknoglobals // protected by package
var l *slog.Logger

type loggerContextKeyType string

//nolint:gochecknoglobals // protected by package
var loggerContextKey loggerContextKeyType = "stdLogger"

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

func GetLoggerFromContext(ctx context.Context) *slog.Logger {
	loggerRaw := ctx.Value(loggerContextKey)
	log, ok := loggerRaw.(*slog.Logger)
	if !ok {
		log = GetLogger()
	}
	return log
}

func ContextWithLogger(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, log)
}
