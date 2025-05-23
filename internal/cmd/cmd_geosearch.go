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

var cGEOSEARCH = &CommandMeta{
	Name:      "GEOSEARCH",
	Syntax:    "GEOSEARCH key longitude latitude radius <M | KM | FT | MI> [WITHCOORD] [WITHDIST] [WITHHASH]",
	HelpShort: "GEOSEARCH Returns the members within the borders of the area specified by a given shape",
	HelpLong: `
Return the members of a sorted set populated with geospatial information using GEOADD, which are within the borders of the area specified by a given shape.
This command extends the GEORADIUS command, so in addition to searching within circular areas, it supports searching within rectangular areas.

This command should be used in place of the deprecated GEORADIUS and GEORADIUSBYMEMBER commands.

The query's center point is provided by one of these mandatory options:

FROMMEMBER: Use the position of the given existing <member> in the sorted set.
FROMLONLAT: Use the given <longitude> and <latitude> position.
The query's shape is provided by one of these mandatory options:

BYRADIUS: Similar to GEORADIUS, search inside circular area according to given <radius>.
BYBOX: Search inside an axis-aligned rectangle, determined by <height> and <width>.
The command optionally returns additional information using the following options:

WITHDIST: Also return the distance of the returned items from the specified center point. The distance is returned in the same unit as specified for the radius or height and width arguments.
WITHCOORD: Also return the longitude and latitude of the matching items.
WITHHASH: Also return the raw geohash-encoded sorted set score of the item, in the form of a 52 bit unsigned integer. This is only useful for low level hacks or debugging and is otherwise of little interest for the general user.

Matching items are returned unsorted by default. To sort them, use one of the following two options:

ASC: Sort returned items from the nearest to the farthest, relative to the center point.
DESC: Sort returned items from the farthest to the nearest, relative to the center point.

All matching items are returned by default. To limit the results to the first N matching items,
use the COUNT <count> option. When the ANY option is used, the command returns as soon as enough matches are found.
This means that the results returned may not be the ones closest to the specified point,
but the effort invested by the server to generate them is significantly less. When ANY is not provided,
the command will perform an effort that is proportional to the number of items matching the specified area and sort them,
so to query very large areas with a very small COUNT option may be slow even if just a few results are returned.

NOTE:
1. If ANY is used with ASC or DESC then sorting is avoided.
2. IF COUNT is used without ASC or DESC then result is automatically sorted with ASC
	`,
	Examples: `
localhost:7379> GEOADD Delhi 77.2167 28.6315 CP 77.2295 28.6129 IndiaGate 77.1197 28.6412 Rajouri 77.1000 28.5562 Airport 77.1900 28.6517 KarolBagh
OK 5
localhost:7379> GEOSEARCH Delhi FROMLONLAT 77.1000 28.5562 BYRADIUS 10 km COUNT 2 ANY
OK
0) Airport
1) Rajouri
localhost:7379> GEOSEARCH Delhi FROMLONLAT 77.1000 28.5562 BYRADIUS 10 km COUNT 2 DESC
OK
0) Rajouri
1) Airport
localhost:7379> GEOSEARCH Delhi FROMLONLAT 77.1000 28.5562 BYRADIUS 20 km COUNT 2 DESC
OK
0) CP
1) IndiaGate
localhost:7379> GEOSEARCH Delhi FROMLONLAT 77.1000 28.5562 BYBOX 40 40 km COUNT 2 ASC
OK
0) Airport
1) Rajouri
localhost:7379> GEOSEARCH Delhi FROMMEMBER CP BYBOX 40 40 km COUNT 2 ASC
OK
0) IndiaGate
0) KarolBagh
	`,
	Eval:    evalGEOSEARCH,
	Execute: executeGEOSEARCH,
}

func init() {
	CommandRegistry.AddCommand(cGEOSEARCH)
}

func newGEOSEARCHRes(elements []*wire.GEOElement) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_GEOSEARCHRes{
				GEOSEARCHRes: &wire.GEOSEARCHRes{
					Elements: elements,
				},
			},
		},
	}
}

var (
	GEOSEARCHResNilRes = newGEOSEARCHRes([]*wire.GEOElement{})
)

func evalGEOSEARCH(c *Cmd, s *dsstore.Store) (*CmdRes, error) {
	if len(c.C.Args) < 5 {
		return GEOSEARCHResNilRes, errors.ErrWrongArgumentCount("GEOSEARCH")
	}

	key := c.C.Args[0]
	params, nonParams := parseParams(c.C.Args[1:])

	// Validate all the parameters
	if len(nonParams) < 2 {
		return GEOSEARCHResNilRes, errors.ErrInvalidSyntax("GEOSEARCH")
	}

	// Get sorted set else return
	var gr *types.GeoRegistry
	obj := s.Get(key)
	if obj == nil {
		return GEODISTResNilRes, nil
	}
	if obj.Type != object.ObjTypeSortedSet {
		return GEOSEARCHResNilRes, errors.ErrWrongTypeOperation
	}
	gr = obj.Value.(*types.GeoRegistry)

	geoElements, err := gr.GeoSearchElementsWithinShape(params, nonParams)

	if err != nil {
		return GEOSEARCHResNilRes, err
	}

	return newGEOSEARCHRes(geoElements), nil

}

func executeGEOSEARCH(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 5 {
		return GEOSEARCHResNilRes, errors.ErrWrongArgumentCount("GEOSEARCH")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalGEOSEARCH(c, shard.Thread.Store())
}
