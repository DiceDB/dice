package mocks

import (
	"context"
	"log/slog"
)

// SlogNoopHandler is a no-op implementation of slog.Handler
type SlogNoopHandler struct{}

func (h SlogNoopHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (h SlogNoopHandler) Handle(context.Context, slog.Record) error { return nil }
func (h SlogNoopHandler) WithAttrs([]slog.Attr) slog.Handler        { return h }
func (h SlogNoopHandler) WithGroup(string) slog.Handler             { return h }
