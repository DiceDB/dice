// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/dgryski/go-farm"
	"github.com/dicedb/dice/internal/object"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/wire"
)

type Cmd struct {
	C        *wire.Command
	ThreadID string
}

type CmdRes struct {
	R        *wire.Response
	ThreadID string
}

type DiceDBCommand struct {
	Name      string
	HelpShort string
	Eval      func(c *Cmd, s *dstore.Store) (*CmdRes, error)
}

type CmdRegistry struct {
	cmds []*DiceDBCommand
}

func Total() int {
	return len(commandRegistry.cmds)
}

func (r *CmdRegistry) AddCommand(cmd *DiceDBCommand) {
	r.cmds = append(r.cmds, cmd)
}

var commandRegistry CmdRegistry = CmdRegistry{
	cmds: []*DiceDBCommand{},
}

func Execute(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	for _, cmd := range commandRegistry.cmds {
		if cmd.Name == c.C.Cmd {
			start := time.Now()
			resp, err := cmd.Eval(c, s)
			if err != nil {
				resp.R.Err = err.Error()
			}

			slog.Debug("command executed",
				slog.Any("cmd", c.C.Cmd),
				slog.String("args", strings.Join(c.C.Args, " ")),
				slog.String("thread_id", c.ThreadID),
				slog.Int("shard_id", s.ShardID),
				slog.Any("took_ns", time.Since(start).Nanoseconds()))
			return resp, err
		}
	}
	return cmdResNil, errors.New("command not found")
}

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

	// InternalObjs is a pointer to list of InternalObjs, representing an optional data structure
	// associated with the command. This contains pointer to the underlying simple
	// types such as int, string or even complex types
	// like hashes, sets, or sorted sets, which are stored and manipulated as objects.
	// WARN: This parameter should be used with caution
	InternalObjs []*object.InternalObj
}

type DiceDBCmds struct {
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

// Key Returns the key which the command operates on.
//
// TODO: This is a naive implementation which assumes that the first argument is the key.
// This is not true for all commands, however, for now this is only used by the watch manager,
// which as of now only supports a small subset of commands (all of which fit this implementation).
func (cmd *DiceDBCmd) Key() string {
	var c string
	if len(cmd.Args) > 0 {
		c = cmd.Args[0]
	}
	return c
}

func errWrongArgumentCount(command string) error {
	return fmt.Errorf("wrong number of arguments for '%s' command", strings.ToUpper(command))
}

var errUnknownObjectType = errors.New("unknown object type")

//nolint:unparam
func errInvalidSyntax(command string) error {
	return fmt.Errorf("invalid syntax for '%s' command", strings.ToUpper(command))
}

//nolint:unparam
func errInvalidValue(command, param string) error {
	return fmt.Errorf("invalid value for a parameter in '%s' command for %s parameter", strings.ToUpper(command), strings.ToUpper(param))
}

var cmdResNil = &CmdRes{R: &wire.Response{
	Value: &wire.Response_VNil{VNil: true},
}}

var cmdResOK = &CmdRes{R: &wire.Response{
	Value: &wire.Response_VStr{VStr: "OK"},
}}
