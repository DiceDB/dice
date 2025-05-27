// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"strconv"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/shardmanager"
	dsstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/internal/types"
	"github.com/dicedb/dicedb-go/wire"
)

var cGEOADD = &CommandMeta{
	Name:      "GEOADD",
	Syntax:    "GEOADD key [NX | XX] [CH] longitude latitude member [longitude latitude member ...]",
	HelpShort: "GEOADD adds all the specified GEO members with the specified longitude & latitude pair to the sorted set stored at key",
	HelpLong: `
GEOADD adds all the specified GEO members with the specified longitude & latitude pair to the sorted set stored at key
The command takes arguments in the standard format x,y so the longitude must be specified before the latitude. 
There are limits to the coordinates that can be indexed: areas very near to the poles are not indexable.

The exact limits, as specified by EPSG:900913 / EPSG:3785 / OSGEO:41001 are the following:

- Valid longitudes are from -180 to 180 degrees.
- Valid latitudes are from -85.05112878 to 85.05112878 degrees.

This has similar options as ZADD
- NX: Only add new elements and do not update existing elements
- XX: Only update existing elements and do not add new elements
`,
	Examples: `
localhost:7379> GEOADD Delhi NX 77.2096 28.6145 "Central Delhi"
OK 1
localhost:7379> GEOADD Delhi 77.2167 28.6315 CP 77.2295 28.6129 IndiaGate 77.1197 28.6412 Rajouri 77.1000 28.5562 Airport 77.1900 28.6517 KarolBagh
OK 5
localhost:7379> GEOADD Delhi NX 77.2096 280 "Central Delhi"
ERR Invalid Longitude, Latitude pair ('77.209600', '280.000000')! Check the range in Docs
	`,
	Eval:    evalGEOADD,
	Execute: executeGEOADD,
}

func init() {
	CommandRegistry.AddCommand(cGEOADD)
}

func newGEOADDRes(count int64) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_GEOADDRes{
				GEOADDRes: &wire.GEOADDRes{
					Count: count,
				},
			},
		},
	}
}

var (
	GEOADDResNilRes = newGEOADDRes(0)
)

func evalGEOADD(c *Cmd, s *dsstore.Store) (*CmdRes, error) {
	if len(c.C.Args) < 4 {
		return GEOADDResNilRes, errors.ErrWrongArgumentCount("GEOADD")
	}

	key := c.C.Args[0]
	params, nonParams := parseParams(c.C.Args[1:])

	if len(nonParams)%3 != 0 {
		return GEOADDResNilRes, errors.ErrWrongArgumentCount("GEOADD")
	}

	var gr *types.GeoRegistry
	obj := s.Get(key)
	if obj == nil {
		gr = types.NewGeoRegistry()
	} else {
		if obj.Type != object.ObjTypeGeoRegistry {
			return GEOADDResNilRes, errors.ErrWrongTypeOperation
		}
		gr = obj.Value.(*types.GeoRegistry)
	}

	GeoCoordinates, members := []*types.GeoCoordinate{}, []string{}
	for i := 0; i < len(nonParams); i += 3 {
		lon, errLon := strconv.ParseFloat(nonParams[i], 10)
		lat, errLat := strconv.ParseFloat(nonParams[i+1], 10)
		if errLon != nil || errLat != nil {
			return GEOADDResNilRes, errors.ErrInvalidNumberFormat
		}
		coordinate, err := types.NewGeoCoordinateFromLonLat(lon, lat)
		if err != nil {
			return GEOADDResNilRes, err
		}
		GeoCoordinates = append(GeoCoordinates, coordinate)
		members = append(members, nonParams[i+2])
	}

	count, err := gr.Add(GeoCoordinates, members, params)
	if err != nil {
		return GEOADDResNilRes, err
	}

	s.Put(key, s.NewObj(gr, -1, object.ObjTypeGeoRegistry), dsstore.WithPutCmd(dsstore.ZAdd))
	return newGEOADDRes(count), nil
}

func executeGEOADD(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 4 {
		return GEOADDResNilRes, errors.ErrWrongArgumentCount("GEOADD")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalGEOADD(c, shard.Thread.Store())
}
