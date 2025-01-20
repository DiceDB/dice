// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

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
