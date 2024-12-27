// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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
