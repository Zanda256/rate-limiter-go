package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
)

// TraceIDFunc represents a function that can return the trace id from
// the specified context.
type TraceIDFunc func(ctx context.Context) string

// Logger represents a logger for logging information.
type Logger struct {
	handler     slog.Handler
	traceIDFunc TraceIDFunc
}

// New constructs a new log for application use.
func New(w io.Writer, minLevel Level, serviceName string, traceIDFunc TraceIDFunc) *Logger {
	return new(w, minLevel, serviceName, traceIDFunc, Events{})
}

func new(w io.Writer, minLevel Level, serviceName string, traceIDFunc TraceIDFunc, events Events) *Logger {

	// Convert the file name to just the name.ext when this key/value will
	// be logged.
	f := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			if source, ok := a.Value.Any().(*slog.Source); ok {
				v := fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line)
				return slog.Attr{Key: "file", Value: slog.StringValue(v)}
			}
		}

		return a
	}

	// Construct the slog JSON handler for use.
	handler := slog.Handler(slog.NewJSONHandler(w, &slog.HandlerOptions{AddSource: true, Level: slog.Level(minLevel), ReplaceAttr: f}))

	// If events are to be processed, wrap the JSON handler around the custom
	// log handler.
	if events.Debug != nil || events.Info != nil || events.Warn != nil || events.Error != nil {
		handler = newLogHandler(handler, events)
	}

	// Attributes to add to every log.
	attrs := []slog.Attr{
		{Key: "service", Value: slog.StringValue(serviceName)},
	}

	// Add those attributes and capture the final handler.
	handler = handler.WithAttrs(attrs)

	return &Logger{
		handler:     handler,
		traceIDFunc: traceIDFunc,
	}
}
