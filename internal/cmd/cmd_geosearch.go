// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"strconv"

	"github.com/dicedb/dice/internal/errors"
	geoUtil "github.com/dicedb/dice/internal/geo"
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
GEOSEARCH Returns the members within the borders of the area specified by a given shape
For now this only supports FROMLONLAT & BYRADIUS - You have to give longitude, latitude and the radius of the area.
The command optionally returns additional information using the following options:

WITHDIST: Also return the distance of the returned items from the specified center point. The distance is returned in the same unit as specified for the radius or height and width arguments.
WITHCOORD: Also return the longitude and latitude of the matching items.
WITHHASH: Also return the raw geohash-encoded sorted set score of the item, in the form of a 52 bit unsigned integer. This is only useful for low level hacks or debugging and is otherwise of little interest for the general user.

The elements are considered to be ordered from the lowest to the highest distance.
	`,
	Examples: `
localhost:7379> GEOADD Delhi 77.2167 28.6315 CP 77.2295 28.6129 IndiaGate 77.1197 28.6412 Rajouri 77.1000 28.5562 Airport 77.1900 28.6517 KarolBagh
OK 5
localhost:7379> GEOSEARCH Delhi 77.1000 28.5562 10 km 
OK 
0) Airport
0) Rajouri
localhost:7379> GEOSEARCH Delhi 77.1000 28.5562 10 km WITHCOORD WITHDIST WITHHASH
OK 
0) 3631198180857159, 0.000300, (77.099997, 28.556200), Airport
0) 3631199276102297, 9.648000, (77.119700, 28.641200), Rajouri
localhost:7379> GEOSEARCH Delhi 77.1000 28.5562 10 unknownUnit
ERR invalid syntax for 'GEOSEARCH' command
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

func getUnitTypeFromParsedParams(params map[types.Param]string) types.Param {
	if params[types.M] != "" {
		return types.M
	} else if params[types.KM] != "" {
		return types.KM
	} else if params[types.MI] != "" {
		return types.MI
	} else if params[types.FT] != "" {
		return types.FT
	} else {
		return ""
	}
}

func evalGEOSEARCH(c *Cmd, s *dsstore.Store) (*CmdRes, error) {
	if len(c.C.Args) < 5 {
		return GEOSEARCHResNilRes, errors.ErrWrongArgumentCount("GEOSEARCH")
	}

	key := c.C.Args[0]
	params, nonParams := parseParams(c.C.Args[1:])

	if len(nonParams) > 3 {
		return GEOSEARCHResNilRes, errors.ErrInvalidSyntax("GEOSEARCH")
	}
	unit := getUnitTypeFromParsedParams(params)
	if len(unit) == 0 {
		return GEOSEARCHResNilRes, errors.ErrInvalidUnit(c.C.Args[3])
	}

	lon, errLon := strconv.ParseFloat(nonParams[0], 10)
	lat, errLat := strconv.ParseFloat(nonParams[1], 10)
	radius, errRad := strconv.ParseFloat(nonParams[2], 10)

	if errLon != nil || errLat != nil || errRad != nil {
		return GEOSEARCHResNilRes, errors.ErrInvalidNumberFormat
	}
	if err := geoUtil.ValidateLonLat(lon, lat); err != nil {
		return GEOSEARCHResNilRes, err
	}
	finalRadius, _ := geoUtil.ConvertToMeter(radius, unit)

	var ss *types.SortedSet
	obj := s.Get(key)
	if obj == nil {
		return GEODISTResNilRes, nil
	}

	if obj.Type != object.ObjTypeSortedSet {
		return GEOSEARCHResNilRes, errors.ErrWrongTypeOperation
	}
	ss = obj.Value.(*types.SortedSet)

	var withCoord, withDist, withHash bool = false, false, false

	if params[types.WITHCOORD] != "" {
		withCoord = true
	}
	if params[types.WITHDIST] != "" {
		withDist = true
	}
	if params[types.WITHHASH] != "" {
		withHash = true
	}

	boudingBox := geoUtil.GetBoundingBoxWithLonLat(lon, lat, finalRadius)

	minLon := boudingBox[0]
	minLat := boudingBox[1]
	maxLon := boudingBox[2]
	maxLat := boudingBox[3]

	minHash := geoUtil.EncodeHash(minLon, minLat)
	maxHash := geoUtil.EncodeHash(maxLon, maxLat)

	zElements := ss.ZRANGE(int(minHash), int(maxHash), true, false)
	geoElements := []*wire.GEOElement{}

	for _, ele := range zElements {

		eleLon, eleLat := geoUtil.DecodeHash(uint64(ele.Score))
		dist := geoUtil.GetDistance(eleLon, eleLat, lon, lat)

		if dist <= finalRadius {

			geoElement := wire.GEOElement{
				Member: ele.Member,
			}

			if withCoord {
				geoElement.Coordinates = &wire.GEOCoordinates{
					Longitude: eleLon,
					Latitude:  eleLat,
				}
			}
			if withDist {
				geoElement.Distance, _ = geoUtil.ConvertDistance(dist, unit)
			}
			if withHash {
				geoElement.Hash = uint64(ele.Score)
			}
			geoElements = append(geoElements, &geoElement)
		}

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
