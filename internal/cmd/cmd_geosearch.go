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
	"github.com/emirpasic/gods/queues/priorityqueue"
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

	// Get sorted set else return
	var ss *types.SortedSet
	obj := s.Get(key)
	if obj == nil {
		return GEODISTResNilRes, nil
	}
	if obj.Type != object.ObjTypeSortedSet {
		return GEOSEARCHResNilRes, errors.ErrWrongTypeOperation
	}
	ss = obj.Value.(*types.SortedSet)

	// Validate all the parameters
	if len(nonParams) < 2 {
		return GEOSEARCHResNilRes, errors.ErrInvalidSyntax("GEOSEARCH")
	}
	unit := getUnitTypeFromParsedParams(params)
	if len(unit) == 0 {
		return GEOSEARCHResNilRes, errors.ErrInvalidUnit(string(unit))
	}

	// Return error if both FROMLONLAT & FROMMEMBER are set
	if params[types.FROMLONLAT] != "" && params[types.FROMMEMBER] != "" {
		return GEOSEARCHResNilRes, errors.ErrInvalidSetOfOptions(string(types.FROMLONLAT), string(types.FROMMEMBER))
	}

	// Return error if none of FROMLONLAT & FROMMEMBER are set
	if params[types.FROMLONLAT] == "" && params[types.FROMMEMBER] == "" {
		return GEOSEARCHResNilRes, errors.ErrNeedOneOfTheOptions(string(types.FROMLONLAT), string(types.FROMMEMBER))
	}

	// Return error if both BYBOX & BYRADIUS are set
	if params[types.BYBOX] != "" && params[types.BYRADIUS] != "" {
		return GEOSEARCHResNilRes, errors.ErrInvalidSetOfOptions(string(types.BYBOX), string(types.BYRADIUS))
	}

	// Return error if none of BYBOX & BYRADIUS are set
	if params[types.BYBOX] == "" && params[types.BYRADIUS] == "" {
		return GEOSEARCHResNilRes, errors.ErrNeedOneOfTheOptions(string(types.BYBOX), string(types.BYRADIUS))
	}

	// Return error if ANY is used without COUNT
	if params[types.ANY] != "" && params[types.COUNT] == "" {
		return GEOSEARCHResNilRes, errors.ErrGeneral("ANY argument requires COUNT argument")
	}

	// Return error if Both ASC & DESC are used
	if params[types.ASC] != "" && params[types.DESC] != "" {
		return GEOSEARCHResNilRes, errors.ErrGeneral("Use one of ASC or DESC")
	}

	// Fetch Longitute and Latitude based on FROMLONLAT & FROMMEMBER param
	var lon, lat float64
	var errLon, errLat error

	// Fetch Longitute and Latitude from params
	if params[types.FROMLONLAT] != "" {
		if len(nonParams) < 2 {
			return GEOSEARCHResNilRes, errors.ErrWrongArgumentCount("GEOSEARCH")
		}
		lon, errLon = strconv.ParseFloat(nonParams[0], 10)
		lat, errLat = strconv.ParseFloat(nonParams[1], 10)

		if errLon != nil || errLat != nil {
			return GEOSEARCHResNilRes, errors.ErrInvalidNumberFormat
		}

		// Adjust the nonParams array for further operations
		nonParams = nonParams[2:]
	}

	// Fetch Longitute and Latitude from member
	if params[types.FROMMEMBER] != "" {
		if len(nonParams) < 1 {
			return GEOSEARCHResNilRes, errors.ErrWrongArgumentCount("GEOSEARCH")
		}
		member := nonParams[0]
		node := ss.GetByKey(member)
		if node == nil {
			return GEOSEARCHResNilRes, errors.ErrMemberNotFoundInSortedSet(member)
		}
		hash := node.Score()
		lon, lat = geoUtil.DecodeHash(uint64(hash))

		// Adjust the nonParams array for further operations
		nonParams = nonParams[1:]
	}

	// Validate Longitute and Latitude
	if err := geoUtil.ValidateLonLat(lon, lat); err != nil {
		return GEOADDResNilRes, err
	}

	// Create shape based on BYBOX or BYRADIUS param
	var searchShape geoUtil.GeoShape

	// Create shape from BYBOX
	if params[types.BYBOX] != "" {
		if len(nonParams) < 2 {
			return GEOSEARCHResNilRes, errors.ErrWrongArgumentCount("GEOSEARCH")
		}

		var width, height float64
		var errWidth, errHeight error
		width, errWidth = strconv.ParseFloat(nonParams[0], 10)
		height, errHeight = strconv.ParseFloat(nonParams[1], 10)

		if errWidth != nil || errHeight != nil {
			return GEOSEARCHResNilRes, errors.ErrInvalidNumberFormat
		}
		if height <= 0 || width <= 0 {
			return GEOSEARCHResNilRes, errors.ErrGeneral("HEIGHT, WIDTH should be > 0")
		}

		searchShape, _ = geoUtil.GetNewGeoShapeRectangle(width, height, lon, lat, unit)

		// Adjust the nonParams array for further operations
		nonParams = nonParams[2:]
	}

	// Create shape from BYRADIUS
	if params[types.BYRADIUS] != "" {
		if len(nonParams) < 1 {
			return GEOSEARCHResNilRes, errors.ErrWrongArgumentCount("GEOSEARCH")
		}

		var radius float64
		var errRad error
		radius, errRad = strconv.ParseFloat(nonParams[0], 10)

		if errRad != nil {
			return GEOSEARCHResNilRes, errors.ErrInvalidNumberFormat
		}
		if radius <= 0 {
			return GEOSEARCHResNilRes, errors.ErrGeneral("RADIUS should be > 0")
		}

		searchShape, _ = geoUtil.GetNewGeoShapeCircle(radius, lon, lat, unit)

		// Adjust the nonParams array for further operations
		nonParams = nonParams[1:]
	}

	// Get COUNT based on Params
	var count int = -1
	var errCount error

	// Check for COUNT to limit the output
	if params[types.COUNT] != "" {
		if len(nonParams) < 1 {
			return GEOSEARCHResNilRes, errors.ErrWrongArgumentCount("GEOSEARCH")
		}
		count, errCount = strconv.Atoi(nonParams[0])
		if errCount != nil {
			return GEOSEARCHResNilRes, errors.ErrInvalidNumberFormat
		}
		if count <= 0 {
			return GEOSEARCHResNilRes, errors.ErrGeneral("COUNT must be > 0")
		}

		// Adjust the nonParams array for further operations
		nonParams = nonParams[1:]
	}

	// If all the params are not used till now
	// Means there're some unknown param
	if len(nonParams) != 0 {
		return GEOSEARCHResNilRes, errors.ErrUnknownOption(nonParams[0])
	}

	// Check for ANY option
	var anyOption bool = false
	if params[types.ANY] != "" {
		anyOption = true
	}

	// Check for Sorting Key ASC or DESC (-1 = DESC, 0 = NoSort, 1 = ASC)
	var sortType float64 = 0
	if params[types.ASC] != "" {
		sortType = 1
	}
	if params[types.DESC] != "" {
		sortType = -1
	}

	// COUNT without ordering does not make much sense (we need to sort in order to return the closest N entries)
	// Note that this is not needed for ANY option
	if count != -1 && sortType == 0 && !anyOption {
		sortType = 1
	}

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

	// Find Neighbors from the shape
	neighbors, steps := geoUtil.GetNeighborsForGeoSearchUsingRadius(searchShape)
	neighborsArr := geoUtil.NeightborsToArray(neighbors)

	// HashMap of all the nodes (we are making map for deduplication)
	geoElementMap := map[string]*wire.GEOElement{}
	totalElements := 0

	// Find all the elements in the neighbor and the center block
	for _, neighbor := range neighborsArr {

		// Discarded neighbors
		if neighbor == 0 {
			continue
		}

		// If ANY option is used and totalElements == count
		// Break the loop and Return the current result
		if anyOption && count == totalElements {
			break
		}

		maxHash, minHash := geoUtil.GetMaxAndMinHashForBoxHash(neighbor, steps)

		zElements := ss.ZRANGE(int(minHash), int(maxHash), true, false)

		for _, ele := range zElements {

			eleLon, eleLat := geoUtil.DecodeHash(uint64(ele.Score))
			dist := searchShape.GetDistanceIfWithinShape(eleLon, eleLat)

			if anyOption && totalElements == count {
				break
			}

			if dist != 0 {

				distConverted, _ := geoUtil.ConvertDistance(dist, unit)
				geoElement := wire.GEOElement{
					Member: ele.Member,
					Coordinates: &wire.GEOCoordinates{
						Longitude: eleLon,
						Latitude:  eleLat,
					},
					Distance: distConverted,
					Hash:     uint64(ele.Score),
				}
				geoElementMap[ele.Member] = &geoElement
				totalElements++
			}

		}
	}

	// Convert map to array
	geoElements := []*wire.GEOElement{}

	for _, ele := range geoElementMap {
		geoElements = append(geoElements, ele)
	}

	// Let count be the total elements we need
	if count == -1 {
		count = totalElements
	}

	// Return unsorted result if ANY is used or sortType = 0
	if anyOption || sortType == 0 {
		filterDimensionsBasedOnFlags(geoElements, withCoord, withDist, withHash)
		return newGEOSEARCHRes(geoElements), nil
	}

	// Comparator function for MaxHeap
	// If ASC is set -> we use MaxHeap -> To Pop out the largest element if LEN > COUNT
	// If DESC is set -> we use MinHeap -> To Pop out the smallest element if LEN > COUNT
	// So Reverse the final array
	cmp := func(a, b interface{}) int {
		distance1 := a.(*wire.GEOElement).Distance
		distance2 := b.(*wire.GEOElement).Distance
		if distance1*sortType < distance2*sortType {
			return 1
		} else if distance1*sortType > distance2*sortType {
			return -1
		}
		return 0
	}

	// Create a priority Queue to store the 'COUNT' results
	pq := priorityqueue.NewWith(cmp)

	for _, ele := range geoElements {
		pq.Enqueue(ele)
		if pq.Size() > count {
			pq.Dequeue()
		}
	}

	// Final result Arr
	resultGeoElements := []*wire.GEOElement{}

	// Transfer elements from priority Queue to Arr
	for pq.Size() > 0 {
		queueEle, _ := pq.Dequeue()
		geoEle := queueEle.(*wire.GEOElement)
		resultGeoElements = append(resultGeoElements, geoEle)
	}

	// Reverse the output array Because
	// If ASC is set -> we use MaxHeap -> Which will give use DESC array
	// If DESC is set -> we use MinHeap -> Which will give use ASC array
	// So Reverse the final array
	for i, j := 0, len(resultGeoElements)-1; i < j; i, j = i+1, j-1 {
		resultGeoElements[i], resultGeoElements[j] = resultGeoElements[j], resultGeoElements[i]
	}

	filterDimensionsBasedOnFlags(resultGeoElements, withCoord, withDist, withHash)

	return newGEOSEARCHRes(resultGeoElements), nil

}

func executeGEOSEARCH(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 5 {
		return GEOSEARCHResNilRes, errors.ErrWrongArgumentCount("GEOSEARCH")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalGEOSEARCH(c, shard.Thread.Store())
}

func filterDimensionsBasedOnFlags(geoElements []*wire.GEOElement, withCoord, withDist, withHash bool) {
	for _, ele := range geoElements {
		if !withCoord {
			ele.Coordinates = nil
		}

		if !withDist {
			ele.Distance = 0
		}

		if !withHash {
			ele.Hash = 0
		}
	}
}
