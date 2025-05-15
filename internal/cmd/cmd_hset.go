// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

type SSMap map[string]string

var cHSET = &CommandMeta{
	Name:      "HSET",
	Syntax:    "HSET key field value [field value ...]",
	HelpShort: "HSET sets field value in the string-string map stored at key",
	HelpLong: `
HSET sets the field and value for the key in the string-string map.

The command returns the number of fields that were added.
	`,
	Examples: `
localhost:7379> HSET k1 f1 v1
OK 1
localhost:7379> HSET k1 f1 v1 f2 v2 f3 v3
OK 2
	`,
	Eval:    evalHSET,
	Execute: executeHSET,
}

func init() {
	CommandRegistry.AddCommand(cHSET)
}

func newHSETRes(count int64) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_HSETRes{
				HSETRes: &wire.HSETRes{
					Count: count,
				},
			},
		},
	}
}

var (
	HSETResNilRes = newHSETRes(0)
)

func evalHSET(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	key := c.C.Args[0]

	var m SSMap
	var countFieldsAdded int64

	obj := s.Get(key)
	if obj != nil {
		if err := object.AssertType(obj.Type, object.ObjTypeSSMap); err != nil {
			return HSETResNilRes, errors.ErrWrongTypeOperation
		}
		m = obj.Value.(SSMap)
	} else {
		m = make(SSMap)
	}

	// kvs is the list of key-value pairs to set in the SSMap
	// key and value are alternating elements in the list
	kvs := c.C.Args[1:]
	if (len(kvs) & 1) == 1 {
		return HSETResNilRes, errors.ErrWrongArgumentCount("HSET")
	}

	for i := 0; i < len(kvs); i += 2 {
		k, v := kvs[i], kvs[i+1]
		if _, ok := m[k]; !ok {
			countFieldsAdded++
		}
		m[k] = v
	}

	obj = s.NewObj(m, -1, object.ObjTypeSSMap)
	s.Put(key, obj)

	return newHSETRes(countFieldsAdded), nil
}

func executeHSET(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 3 {
		return HSETResNilRes, errors.ErrWrongArgumentCount("HSET")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalHSET(c, shard.Thread.Store())
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
