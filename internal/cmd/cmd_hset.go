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

type HashMap map[string]string

var cHSET = &CommandMeta{
	Name:      "HSET",
	Syntax:    "HSET key field value [field value ...]",
	HelpShort: "HSET sets field value in the hash stored at key",
	HelpLong: `
HSET sets the field and value for the key in args.

Returns "OK" if the field value was set or updated to hash key. Returns (nil) if the field value was not set or updated to hash key.
	`,
	Examples: `
localhost:7379> HSET k1 f1 v1
OK OK
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

func (h HashMap) Get(k string) (*string, bool) {
	value, ok := h[k]
	if !ok {
		return nil, false
	}
	return &value, true
}

func (h HashMap) Set(k, v string) (*string, bool) {
	value, ok := h[k]
	if ok {
		oldValue := value
		h[k] = v
		return &oldValue, true
	}

	h[k] = v
	return nil, false
}

func hashMapBuilder(keyValuePairs []string, currentHashMap HashMap) (HashMap, int64, error) {
	var hmap HashMap
	var numKeysNewlySet int64

	if currentHashMap == nil {
		hmap = make(HashMap)
	} else {
		hmap = currentHashMap
	}

	iter := 0
	argLength := len(keyValuePairs)

	for iter <= argLength-1 {
		if iter >= argLength-1 || iter+1 > argLength-1 {
			return hmap, -1, errors.ErrWrongArgumentCount("HSET")
		}

		k := keyValuePairs[iter]
		v := keyValuePairs[iter+1]

		_, present := hmap.Set(k, v)
		if !present {
			numKeysNewlySet++
		}
		iter += 2
	}

	return hmap, numKeysNewlySet, nil
}

func evalHSET(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	key := c.C.Args[0]
	obj := s.Get(key)

	var hashMap HashMap

	if obj != nil {
		if err := object.AssertType(obj.Type, object.ObjTypeHashMap); err != nil {
			return cmdResNil, errors.ErrWrongTypeOperation
		}
		hashMap = obj.Value.(HashMap)
	}

	keyValuePairs := c.C.Args[1:]

	hashMap, numKeys, err := hashMapBuilder(keyValuePairs, hashMap)
	if err != nil {
		return cmdResNil, err
	}

	obj = s.NewObj(hashMap, -1, object.ObjTypeHashMap)
	s.Put(key, obj)

	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{VInt: numKeys},
	}}, nil
}

func executeHSET(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 3 {
		return cmdResNil, errors.ErrWrongArgumentCount("HSET")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalHSET(c, shard.Thread.Store())
}
