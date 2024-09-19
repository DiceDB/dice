package logger

import (
	"context"
	"log/slog"

	"github.com/rs/zerolog"
)

// ZerologHandler is a custom handler that adapts slog to zerolog
type ZerologHandler struct {
	logger zerolog.Logger
}

// newZerologHandler creates a new ZerologHandler
func newZerologHandler(logger zerolog.Logger) *ZerologHandler {
	return &ZerologHandler{logger: logger}
}

// Handle implements the slog.Handler interface
func (h *ZerologHandler) Handle(ctx context.Context, record slog.Record) error {
	event := h.logger.WithLevel(mapLevel(record.Level))

	record.Attrs(func(attr slog.Attr) bool {
		switch attr.Value.Kind() {
		case slog.KindString:
			event = event.Str(attr.Key, attr.Value.String())
		case slog.KindInt64:
			event = event.Int64(attr.Key, attr.Value.Int64())
		case slog.KindFloat64:
			event = event.Float64(attr.Key, attr.Value.Float64())
		case slog.KindBool:
			event = event.Bool(attr.Key, attr.Value.Bool())
		default:
			// Log unknown types generically
			event = event.Interface(attr.Key, attr.Value.Any())
		}
		return true
	})

	event.Msg(record.Message)
	return nil
}

// mapLevel maps slog levels to zerolog levels
func mapLevel(level slog.Level) zerolog.Level {
	switch {
	case level == slog.LevelDebug:
		return zerolog.DebugLevel
	case level == slog.LevelWarn:
		return zerolog.WarnLevel
	case level == slog.LevelError:
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

// Enabled implements the slog.Handler interface
func (h *ZerologHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.logger.GetLevel() <= mapLevel(level)
}

// WithAttrs adds attributes to the log event
func (h *ZerologHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	logger := h.logger
	for _, attr := range attrs {
		logger = logger.With().Interface(attr.Key, attr.Value).Logger()
	}
	return newZerologHandler(logger)
}

// WithGroup returns the handler for a group of logs
func (h *ZerologHandler) WithGroup(name string) slog.Handler {
	logger := h.logger.With().Str("group", name).Logger()
	return newZerologHandler(logger)
}
