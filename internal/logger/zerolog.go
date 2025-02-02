// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package logger

import (
	"context"
	"log/slog"
	"time"

	"github.com/rs/zerolog"
)

// ZerologHandler is a custom handler that adapts slog to zerolog
type ZerologHandler struct {
	logger *zerolog.Logger
}

// newZerologHandler creates a new ZerologHandler
func newZerologHandler(logger *zerolog.Logger) *ZerologHandler {
	return &ZerologHandler{
		logger: logger,
	}
}

// Handle implements the slog.Handler interface
//
//nolint:gocritic // The slog.Record struct triggers hugeParam, but we don't control the interface (it's a standard library one)
func (h *ZerologHandler) Handle(_ context.Context, record slog.Record) error {
	event := h.logger.WithLevel(toZerologLevel(record.Level))
	record.Attrs(func(attr slog.Attr) bool {
		addAttrToZerolog(attr, event)
		return true
	})
	event.Msg(record.Message)
	return nil
}

// Enabled implements the slog.Handler interface
func (h *ZerologHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.logger.GetLevel() <= toZerologLevel(level)
}

// WithAttrs adds attributes to the log event
func (h *ZerologHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	ctx := h.logger.With()
	for _, attr := range attrs {
		ctx = addAttrToZerolog(attr, ctx)
	}
	logger := ctx.Logger()
	return newZerologHandler(&logger)
}

// WithGroup returns the handler for a group of logs
func (h *ZerologHandler) WithGroup(name string) slog.Handler {
	logger := h.logger.With().Str("group", name).Logger()
	return newZerologHandler(&logger)
}

// addAttrToZerolog is a generic function to add a slog.Attr to either a zerolog.Event or zerolog.Context
func addAttrToZerolog[T interface {
	Str(string, string) T
	Int64(string, int64) T
	Uint64(string, uint64) T
	Float64(string, float64) T
	Bool(string, bool) T
	Err(error) T
	Time(string, time.Time) T
	Dur(string, time.Duration) T
	Interface(string, any) T
	AnErr(key string, err error) T
}](attr slog.Attr, target T) T {
	switch attr.Value.Kind() {
	case slog.KindBool:
		return target.Bool(attr.Key, attr.Value.Bool())
	case slog.KindDuration:
		return target.Dur(attr.Key, attr.Value.Duration())
	case slog.KindFloat64:
		return target.Float64(attr.Key, attr.Value.Float64())
	case slog.KindInt64:
		return target.Int64(attr.Key, attr.Value.Int64())
	case slog.KindString:
		return target.Str(attr.Key, attr.Value.String())
	case slog.KindTime:
		return target.Time(attr.Key, attr.Value.Time())
	case slog.KindUint64:
		return target.Uint64(attr.Key, attr.Value.Uint64())
	case slog.KindGroup:
		// For group, we need to recurse into the group's attributes
		group := attr.Value.Group()
		for _, groupAttr := range group {
			target = addAttrToZerolog(groupAttr, target)
		}
		return target
	case slog.KindLogValuer:
		// LogValuer is a special case that needs to be resolved
		resolved := attr.Value.Resolve()
		return addAttrToZerolog(slog.Attr{Key: attr.Key, Value: resolved}, target)
	default:
		switch v := attr.Value.Any().(type) {
		// error is a special case since zerlog has a dedicated method for it, but it's not part of the slog.Kind
		case error:
			return target.AnErr(attr.Key, v)
		default:
			return target.Interface(attr.Key, attr.Value)
		}
	}
}

// toZerologLevel maps slog levels to zerolog levels
func toZerologLevel(level slog.Level) zerolog.Level {
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
