// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
)

type SSMap map[string]string

var cHSET = &CommandMeta{
	Name:      "HSET",
	Syntax:    "HSET key field value [field value ...]",
	HelpShort: "HSET sets field value in the string-string map stored at key",
	HelpLong: `
HSET sets the field and value for the key in the string-string map.

Returns "OK" if the field value was set or updated in the map. Returns (nil) if the
field value was not set or updated.
	`,
	Examples: `
localhost:7379> HSET k1 f1 v1
OK 1
localhost:7379> HSET k1 f1 v1 f2 v2 f3 v3
OK 2
localhost:7379> HGET k1 f1
OK v1
localhost:7379> HGET k2 f1
OK (nil)
	`,
	Eval:    evalHSET,
	Execute: executeHSET,
}

func init() {
	CommandRegistry.AddCommand(cHSET)
}

// Get returns the value for the key in the SSMap.
// Returns false if the key does not exist.
// Returns the value if the key exists.
func (h SSMap) Get(k string) (string, bool) {
	value, ok := h[k]
	if !ok {
		return "", false
	}
	return value, true
}

// Set sets the value v for the key k in the SSMap.
// Returns the old value if the key exists.
// The bool return value indicates if the key was already present in the SSMap.
func (h SSMap) Set(k, v string) (string, bool) {
	value, ok := h[k]
	if ok {
		oldValue := value
		h[k] = v
		return oldValue, true
	}

	h[k] = v
	return "", false
}

func evalHSET(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	key := c.C.Args[0]

	var m SSMap
	var newFields int64

	obj := s.Get(key)
	if obj != nil {
		if err := object.AssertType(obj.Type, object.ObjTypeSSMap); err != nil {
			return cmdResNil, errors.ErrWrongTypeOperation
		}
		m = obj.Value.(SSMap)
	} else {
		m = make(SSMap)
	}

	// kvs is the list of key-value pairs to set in the SSMap
	// key and value are alternating elements in the list
	kvs := c.C.Args[1:]
	if (len(kvs) & 1) == 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("HSET")
	}

	for i := 0; i < len(kvs); i += 2 {
		k, v := kvs[i], kvs[i+1]
		if _, ok := m[k]; !ok {
			newFields++
		}
		m[k] = v
	}

	obj = s.NewObj(m, -1, object.ObjTypeSSMap)
	s.Put(key, obj)

	return cmdResInt(newFields), nil
}

func executeHSET(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 3 {
		return cmdResNil, errors.ErrWrongArgumentCount("HSET")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalHSET(c, shard.Thread.Store())
}
