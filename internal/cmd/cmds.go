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
	"github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

// nolint: stylecheck
const INFINITE_EXPIRATION = int64(-1)

type Cmd struct {
	C        *wire.Command
	IsReplay bool
	ClientID string
	Mode     string
}

func (c *Cmd) String() string {
	return fmt.Sprintf("%s %s", c.C.Cmd, strings.Join(c.C.Args, " "))
}

func (c *Cmd) Fingerprint() uint32 {
	return farm.Fingerprint32([]byte(c.String()))
}

func (c *Cmd) Key() string {
	if len(c.C.Args) > 0 {
		return c.C.Args[0]
	}
	return ""
}

type CmdRes struct {
	R        *wire.Response
	ClientID string
}

type DiceDBCommand struct {
	Name      string
	HelpShort string
	Eval      func(c *Cmd, s *store.Store) (*CmdRes, error)
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

func Execute(c *Cmd, s *store.Store) (*CmdRes, error) {
	// TODO: Replace this iteration with a HashTable lookup.
	for _, cmd := range commandRegistry.cmds {
		if cmd.Name != c.C.Cmd {
			continue
		}

		start := time.Now()
		resp, err := cmd.Eval(c, s)
		if err != nil {
			resp.R.Err = err.Error()
		}

		slog.Debug("command executed",
			slog.Any("cmd", c.String()),
			slog.String("client_id", c.ClientID),
			slog.String("mode", c.Mode),
			slog.Int("shard_id", s.ShardID),
			slog.Any("took_ns", time.Since(start).Nanoseconds()))
		return resp, err
	}
	return cmdResNil, fmt.Errorf("command '%s' not found", c.C.Cmd)
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
func (cmd *DiceDBCmd) Fingerprint() uint32 {
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

func errWrongTypeOperation(command string) error {
	return fmt.Errorf("wrong type operation for '%s' command", strings.ToUpper(command))
}

var cmdResNil = &CmdRes{R: &wire.Response{
	Value: &wire.Response_VNil{VNil: true},
}}

var cmdResOK = &CmdRes{R: &wire.Response{
	Value: &wire.Response_VStr{VStr: "OK"},
}}

var cmdResIntNegTwo = &CmdRes{R: &wire.Response{
	Value: &wire.Response_VInt{VInt: -2},
}}

var cmdResIntNegOne = &CmdRes{R: &wire.Response{
	Value: &wire.Response_VInt{VInt: -1},
}}