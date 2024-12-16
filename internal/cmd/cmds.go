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

package cmd

import (
	"fmt"
	"strings"

	"github.com/dgryski/go-farm"
	"github.com/dicedb/dice/internal/object"
)

// DiceDBCmd represents a command structure to be executed
// within a DiceDB system. This struct emulates the way DiceDB commands
// are structured, including the command itself, additional arguments,
// and an optional object to store or manipulate.
type DiceDBCmd struct {
	// Cmd represents the command to execute (e.g., "SET", "GET", "DEL").
	// This is the main command keyword that specifies the action to perform
	// in DiceDB. For example:
	// - "SET": To store a value.
	// - "GET": To retrieve a value.
	// - "DEL": To delete a value.
	// - "EXPIRE": To set a time-to-live for a key.
	Cmd string

	// Args holds any additional parameters required by the command.
	// For example:
	// - If Cmd is "SET", Args might contain ["key", "value"].
	// - If Cmd is "EXPIRE", Args might contain ["key", "seconds"].
	// This slice allows flexible support for commands with variable arguments.
	Args []string

	// InternalObj is a pointer to an InternalObj, representing an optional data structure
	// associated with the command. This contains pointer to the underlying simple
	// types such as int, string or even complex types
	// like hashes, sets, or sorted sets, which are stored and manipulated as objects.
	// WARN: This parameter should be used with caution
	InternalObj *object.InternalObj
}

type RedisCmds struct {
	Cmds      []*DiceDBCmd
	RequestID uint32
}

// Repr returns a string representation of the command.
func (cmd *DiceDBCmd) Repr() string {
	return fmt.Sprintf("%s %s", cmd.Cmd, strings.Join(cmd.Args, " "))
}

// GetFingerprint returns a 32-bit fingerprint of the command and its arguments.
func (cmd *DiceDBCmd) GetFingerprint() uint32 {
	return farm.Fingerprint32([]byte(cmd.Repr()))
}

// GetKey Returns the key which the command operates on.
//
// TODO: This is a naive implementation which assumes that the first argument is the key.
// This is not true for all commands, however, for now this is only used by the watch manager,
// which as of now only supports a small subset of commands (all of which fit this implementation).
func (cmd *DiceDBCmd) GetKey() string {
	var c string
	if len(cmd.Args) > 0 {
		c = cmd.Args[0]
	}
	return c
}
