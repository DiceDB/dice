// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/shardmanager"
	dsstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/internal/types"
	"github.com/dicedb/dicedb-go/wire"
)

var cGEOHASH = &CommandMeta{
	Name:      "GEOHASH",
	Syntax:    "GEOHASH key [member [member ...]]",
	HelpShort: "Returns valid Geohash strings representing the position of one or more elements in a sorted set value representing a geospatial index",
	HelpLong: `
	The command returns 11 characters Geohash strings, so no precision is lost compared to the Redis internal 52 bit representation
	The returned Geohashes have the following properties:
	1. They can be shortened removing characters from the right. It will lose precision but will still point to the same area.
	2. Strings with a similar prefix are nearby, but the contrary is not true, it is possible that strings with different prefixes are nearby too.
	3. It is possible to use them in geohash.org URLs such as http://geohash.org/<geohash-string>.
	`,
	Examples: `
localhost:7379> GEOADD Delhi 77.2167 28.6315 CP 77.2295 28.6129 IndiaGate 77.1197 28.6412 Rajouri 77.1000 28.5562 Airport 77.1900 28.6517 KarolBagh
OK 5
localhost:7379> GEOHASh Delhi CP IndiaGate
OK 
0) ttnfvh5qxd0
1) ttnfv2uf1z0
	`,
	Eval:    evalGEOHASH,
	Execute: executeGEOHASH,
}

func init() {
	CommandRegistry.AddCommand(cGEOHASH)
}

func newGEOHASHRes(hashArr []string) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_GEOHASHRes{
				GEOHASHRes: &wire.GEOHASHRes{
					Hashes: hashArr,
				},
			},
		},
	}
}

var (
	GEOHASHResNilRes = newGEOHASHRes([]string{})
)

func evalGEOHASH(c *Cmd, s *dsstore.Store) (*CmdRes, error) {
	if len(c.C.Args) < 2 {
		return GEOHASHResNilRes, errors.ErrWrongArgumentCount("GEOHASH")
	}

	key := c.C.Args[0]
	var gr *types.GeoRegistry
	obj := s.Get(key)
	if obj == nil {
		return GEOHASHResNilRes, nil
	}
	if obj.Type != object.ObjTypeGeoRegistry {
		return GEOHASHResNilRes, errors.ErrWrongTypeOperation
	}
	gr = obj.Value.(*types.GeoRegistry)

	hashArr := gr.Get11BytesHash(c.C.Args[1:])

	return newGEOHASHRes(hashArr), nil

}

func executeGEOHASH(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 2 {
		return GEOHASHResNilRes, errors.ErrWrongArgumentCount("GEOHASH")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalGEOHASH(c, shard.Thread.Store())
}
