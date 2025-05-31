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

var cGEOPOS = &CommandMeta{
	Name:      "GEOPOS",
	Syntax:    "GEOPOS key [member [member ...]]",
	HelpShort: "Return the positions (longitude,latitude) of all the specified members of the geospatial index.",
	HelpLong: `
		Given a sorted set representing a geospatial index, populated using the GEOADD command, it is often useful to obtain back the coordinates of specified member. 
		When the geospatial index is populated via GEOADD the coordinates are converted into a 52 bit geohash, 
		so the coordinates returned may not be exactly the ones used in order to add the elements, but small errors may be introduced.
	`,
	Examples: `
localhost:7379> GEOADD Delhi 77.2167 28.6315 CP 77.2295 28.6129 IndiaGate 77.1197 28.6412 Rajouri 77.1000 28.5562 Airport 77.1900 28.6517 KarolBagh
OK 5
localhost:7379> GEOPOS Delhi CP nonEx
OK 
0) 77.216700, 28.631498
1) (nil)
	`,
	Eval:    evalGEOPOS,
	Execute: executeGEOPOS,
}

func init() {
	CommandRegistry.AddCommand(cGEOPOS)
}

func newGEOPOSRes(coordinates []*types.GeoCoordinate) *CmdRes {
	reponseCoordinates := make([]*wire.GEOCoordinates, len(coordinates))
	for i, coord := range coordinates {
		if coord == nil {
			continue
		}
		reponseCoordinates[i] = &wire.GEOCoordinates{
			Longitude: coord.Longitude,
			Latitude:  coord.Latitude,
		}
	}
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_GEOPOSRes{
				GEOPOSRes: &wire.GEOPOSRes{
					Coordinates: reponseCoordinates,
				},
			},
		},
	}
}

var (
	GEOPOSResNilRes = newGEOPOSRes([]*types.GeoCoordinate{})
)

func evalGEOPOS(c *Cmd, s *dsstore.Store) (*CmdRes, error) {
	if len(c.C.Args) < 2 {
		return GEOPOSResNilRes, errors.ErrWrongArgumentCount("GEOPOS")
	}

	key := c.C.Args[0]
	var gr *types.GeoRegistry
	obj := s.Get(key)
	if obj == nil {
		return GEOPOSResNilRes, nil
	}
	if obj.Type != object.ObjTypeGeoRegistry {
		return GEOPOSResNilRes, errors.ErrWrongTypeOperation
	}
	gr = obj.Value.(*types.GeoRegistry)
	coordinates := gr.GetCoordinates(c.C.Args[1:])
	return newGEOPOSRes(coordinates), nil
}

func executeGEOPOS(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 2 {
		return GEOPOSResNilRes, errors.ErrWrongArgumentCount("GEOPOS")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalGEOPOS(c, shard.Thread.Store())
}
