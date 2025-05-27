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

var cGEODIST = &CommandMeta{
	Name:      "GEODIST",
	Syntax:    "GEODIST key member1 member2 [M | KM | FT | MI]",
	HelpShort: "GEODIST Return the distance between two members in the geospatial index represented by the sorted set.",
	HelpLong: `
GEODIST Return the distance between two members in the geospatial index represented by the sorted set.
If any of the member is null, this will return Nil Output

The unit must be one of the following, and defaults to meters:
- m for meters.
- km for kilometers.
- mi for miles.
- ft for feet.
	`,
	Examples: `
localhost:7379> GEOADD Delhi 77.2096 28.6145 centralDelhi 77.2167 28.6315 CP 77.2295 28.6129 IndiaGate
OK 3
localhost:7379> GEODIST Delhi CP IndiaGate km
OK 2.416700
	`,
	Eval:    evalGEODIST,
	Execute: executeGEODIST,
}

func init() {
	CommandRegistry.AddCommand(cGEODIST)
}

func newGEODISTRes(distance float64) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_GEODISTRes{
				GEODISTRes: &wire.GEODISTRes{
					Distance: distance,
				},
			},
		},
	}
}

var (
	GEODISTResNilRes = newGEODISTRes(0)
)

func evalGEODIST(c *Cmd, s *dsstore.Store) (*CmdRes, error) {
	if len(c.C.Args) < 3 || len(c.C.Args) > 4 {
		return GEODISTResNilRes, errors.ErrWrongArgumentCount("GEODIST")
	}

	key := c.C.Args[0]
	params, nonParams := parseParams(c.C.Args[1:])

	unit := types.GetUnitTypeFromParsedParams(params)
	if len(c.C.Args) == 4 && len(unit) == 0 {
		return GEODISTResNilRes, errors.ErrInvalidUnit(c.C.Args[3])
	} else if len(unit) == 0 {
		unit = types.M
	}

	var gr *types.GeoRegistry
	obj := s.Get(key)
	if obj == nil {
		return GEODISTResNilRes, nil
	}
	if obj.Type != object.ObjTypeGeoRegistry {
		return GEODISTResNilRes, errors.ErrWrongTypeOperation
	}
	gr = obj.Value.(*types.GeoRegistry)

	dist, err := gr.GetDistanceBetweenMembers(nonParams[0], nonParams[1], unit)

	if err != nil {
		return GEODISTResNilRes, err
	}

	return newGEODISTRes(dist), nil

}

func executeGEODIST(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 3 || len(c.C.Args) > 4 {
		return GEODISTResNilRes, errors.ErrWrongArgumentCount("GEODIST")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalGEODIST(c, shard.Thread.Store())
}
