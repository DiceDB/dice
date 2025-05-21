package geoUtil

import (
	"math"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/types"
	"github.com/mmcloughlin/geohash"
)

// Bit precision - Same as redis (https://github.com/redis/redis/blob/5d0d64b062c160093dc287ed5d18ec7c807873cf/src/geohash_helper.c#L213)
const BIT_PRECISION = 52

// Earth's radius in meters
const EARTH_RADIUS float64 = 6372797.560856

// The maximum/minumum projected coordinate value (in meters) in the Web Mercator projection (EPSG:3857)
// Earth’s equator: ~40,075 km → half of that = ~20,037 km
// The Mercator projection transforms the globe into a square map.
// MERCATOR_MAX is the extent of that square in meters.
const MERCATOR_MAX float64 = 20037726.37
const MERCATOR_MIN float64 = -20037726.37

// Limits from EPSG:900913 / EPSG:3785 / OSGEO:41001
const LAT_MIN float64 = -85.05112878
const LAT_MAX float64 = 85.05112878
const LONG_MIN float64 = -180
const LONG_MAX float64 = 180

type Neighbors struct {
	North     uint64
	NorthEast uint64
	East      uint64
	SouthEast uint64
	South     uint64
	SouthWest uint64
	West      uint64
	NorthWest uint64
	Center    uint64
}

func ArrayToNeighbors(arr []uint64) *Neighbors {
	neightbors := Neighbors{}
	if len(arr) < 8 {
		return &neightbors
	}

	neightbors.North = arr[0]
	neightbors.NorthEast = arr[1]
	neightbors.East = arr[2]
	neightbors.SouthEast = arr[3]
	neightbors.South = arr[4]
	neightbors.SouthWest = arr[5]
	neightbors.West = arr[6]
	neightbors.NorthWest = arr[7]
	neightbors.Center = arr[8]

	return &neightbors
}

func NeightborsToArray(neightbors *Neighbors) [9]uint64 {
	arr := [9]uint64{}

	arr[0] = neightbors.North
	arr[1] = neightbors.NorthEast
	arr[2] = neightbors.East
	arr[3] = neightbors.SouthEast
	arr[4] = neightbors.South
	arr[5] = neightbors.SouthWest
	arr[6] = neightbors.West
	arr[7] = neightbors.NorthWest
	arr[8] = neightbors.Center

	return arr
}

// Computes the min (inclusive) and max (exclusive) scores for a given hash box.
// Aligns the geohash bits to BIT_PRECISION score by left-shifting
func GetMaxAndMinHashForBoxHash(hash uint64, steps uint) (max, min uint64) {
	shift := BIT_PRECISION - (steps)
	base := hash << shift
	rangeSize := uint64(1) << shift
	min = base
	max = base + rangeSize - 1

	return max, min
}

func ValidateLonLat(longitude, latitude float64) error {
	if latitude > LAT_MAX || latitude < LAT_MIN || longitude > LONG_MAX || longitude < LONG_MIN {
		return errors.ErrInvalidLonLatPair(longitude, latitude)
	}
	return nil
}

// Encode given Lon and Lat to GEOHASH
func EncodeHash(longitude, latitude float64) uint64 {
	return geohash.EncodeIntWithPrecision(latitude, longitude, BIT_PRECISION)
}

// DecodeHash returns the latitude and longitude from a geo hash
func DecodeHash(hash uint64) (lon, lat float64) {
	lat, lon = geohash.DecodeIntWithPrecision(hash, BIT_PRECISION)
	return lon, lat
}

func GetDistance(lon1, lat1, lon2, lat2 float64) float64 {
	lon1r := DegToRad(lon1)
	lon2r := DegToRad(lon2)
	lat1r := DegToRad(lat1)
	lat2r := DegToRad(lat2)

	v := math.Sin((lon2r - lon1r) / 2)
	// if v == 0 we can avoid doing expensive math when lons are practically the same (This impl is same as redis)
	if v == 0.0 {
		return GetLatDistance(lat1r, lat2r)
	}

	u := math.Sin((lat2r - lat1r) / 2)

	a := u*u + math.Cos(lat1r)*math.Cos(lat2r)*v*v

	return 2.0 * EARTH_RADIUS * math.Asin(math.Sqrt(a))
}

func DegToRad(deg float64) float64 {
	return deg * math.Pi / 180
}

func RadToDeg(rad float64) float64 {
	return 180.0 * rad / math.Pi
}

func GetLatDistance(lat1, lat2 float64) float64 {
	return EARTH_RADIUS * math.Abs(DegToRad(lat2)-DegToRad(lat1))
}

// ConvertDistance converts a distance from meters to the desired unit
func ConvertDistance(distance float64, unit types.Param) (float64, error) {
	var result float64

	switch unit {
	case types.M:
		result = distance
	case types.KM:
		result = distance / 1000
	case types.MI:
		result = distance / 1609.34
	case types.FT:
		result = distance / 0.3048
	default:
		return 0, errors.ErrInvalidUnit(string(unit))
	}

	// Round to 5 decimal places
	return math.Round(result*10000) / 10000, nil
}

// ConvertToMeter converts a distance to meters from the given unit
func ConvertToMeter(distance float64, unit types.Param) (float64, error) {
	var result float64

	switch unit {
	case types.M:
		result = distance
	case types.KM:
		result = distance * 1000
	case types.MI:
		result = distance * 1609.34
	case types.FT:
		result = distance * 0.3048
	default:
		return 0, errors.ErrInvalidUnit(string(unit))
	}
	// Round to 4 decimal places
	return math.Round(result*10000) / 10000, nil
}

// Return the bounding box of the search circle
// bounds[0] - bounds[2] is the minimum and maximum longitude
// bounds[1] - bounds[3] is the minimum and maximum latitude.
// Refer to this link to understand this function in detail - https://www.notion.so/Geo-Bounding-Box-Research-1f6a37dc1a9a80e7ac43feeeab7215bb?pvs=4
// since the higher the latitude, the shorter the arc length, the box shape is as follows
//
//	  \-----------------/          --------               \-----------------/
//	   \               /         /          \              \               /
//	    \  (long,lat) /         / (long,lat) \              \  (long,lat) /
//	     \           /         /              \             /             \
//	       ---------          /----------------\           /---------------\
//	Northern Hemisphere       Southern Hemisphere         Around the equator
func GetBoundingBoxForRectangleWithHash(hash uint64, widht float64, height float64) *geohash.Box {

	boudingBox := geohash.Box{}

	widht /= 2
	height /= 2
	lon, lat := DecodeHash(hash)
	latDelta := RadToDeg(height / EARTH_RADIUS)
	lonDeltaTop := RadToDeg(widht / EARTH_RADIUS / math.Cos(DegToRad(lat+latDelta)))
	lonDeltaBottom := RadToDeg(widht / EARTH_RADIUS / math.Cos(DegToRad(lat-latDelta)))

	boudingBox.MinLat = lat - latDelta
	boudingBox.MaxLat = lat + latDelta

	if lat < 0 {
		boudingBox.MinLng = lon - lonDeltaBottom
		boudingBox.MaxLng = lon + lonDeltaBottom
	} else {
		boudingBox.MinLng = lon - lonDeltaTop
		boudingBox.MaxLng = lon + lonDeltaTop
	}

	return &boudingBox

}

func GetBoundingBoxForRectangleWithLonLat(lon, lat float64, widht float64, height float64) *geohash.Box {
	return GetBoundingBoxForRectangleWithHash(EncodeHash(lon, lat), widht, height)
}

// Find the step → Precision at which 9 cells (3x3 cells) can cover the entire area
func EstimateStepsByRadius(radius float64, latitude float64) uint {
	if radius == 0 {
		return 26
	}
	var step uint = 1
	for radius < MERCATOR_MAX {
		radius *= 2
		step++
	}
	step -= 2 /* Make sure range is included in most of the base cases. */

	// Wider range towards the poles
	if latitude > 66 || latitude < -66 {
		step--
		if latitude > 80 || latitude < -80 {
			step--
		}
	}

	/* Frame to valid range. */
	if step < 1 {
		step = 1
	}
	if step > 26 {
		step = 26
	}
	return step
}

func GetNeighborsForGeoSearchUsingRadius(geoShape GeoShape) (neighbors *Neighbors, steps uint) {
	lon, lat := geoShape.GetLonLat()
	width, height := geoShape.GetBoudingBoxWidhtAndHeight()
	radius := geoShape.GetRadius()

	// Create bounding box to validate/invalidate neighbors later
	boudingBox := GetBoundingBoxForRectangleWithLonLat(lon, lat, width, height)

	// as (mmcloughlin/geohash) requires total number of bits, not steps
	steps = 2 * EstimateStepsByRadius(radius, lat)

	centerHash := geohash.EncodeIntWithPrecision(lat, lon, steps)
	centerBox := geohash.BoundingBoxIntWithPrecision(centerHash, steps)
	neighborsArr := geohash.NeighborsIntWithPrecision(centerHash, steps)
	neighborsArr = append(neighborsArr, centerHash)
	neighbors = ArrayToNeighbors(neighborsArr)

	// Check if the step is enough at the limits of the covered area.
	// Decode each of the 8 neighbours to get max and min (lon, lat)
	// If North.maxLatitude < maxLatitude(from bouding box) then we have to reduce step to increase neighbour size
	// Do this for N, S, E, W
	northBox := geohash.BoundingBoxIntWithPrecision(neighbors.North, steps)
	eastBox := geohash.BoundingBoxIntWithPrecision(neighbors.East, steps)
	southBox := geohash.BoundingBoxIntWithPrecision(neighbors.South, steps)
	westBox := geohash.BoundingBoxIntWithPrecision(neighbors.West, steps)

	if northBox.MaxLat < boudingBox.MaxLat || southBox.MinLat > boudingBox.MinLat || eastBox.MaxLng < boudingBox.MaxLng || westBox.MinLng > boudingBox.MinLng {
		steps--
		centerHash = geohash.EncodeIntWithPrecision(lat, lon, steps)
		centerBox = geohash.BoundingBoxIntWithPrecision(centerHash, steps)
		neighborsArr := geohash.NeighborsIntWithPrecision(centerHash, steps)
		neighborsArr = append(neighborsArr, centerHash)
		neighbors = ArrayToNeighbors(neighborsArr)
	}

	// Update the center block as well
	neighbors.Center = centerHash

	// Exclude search areas that are useless
	// why not at step == 1? Because geohash cells are so large that excluding neighbors could miss valid points.
	if steps >= 2 {
		if centerBox.MinLat < boudingBox.MinLat {
			neighbors.South = 0
			neighbors.SouthWest = 0
			neighbors.SouthEast = 0
		}
		if centerBox.MaxLat > boudingBox.MaxLat {
			neighbors.North = 0
			neighbors.NorthEast = 0
			neighbors.NorthWest = 0
		}
		if centerBox.MinLng < boudingBox.MinLng {
			neighbors.West = 0
			neighbors.SouthWest = 0
			neighbors.NorthWest = 0
		}
		if centerBox.MaxLng > boudingBox.MaxLng {
			neighbors.East = 0
			neighbors.SouthEast = 0
			neighbors.NorthEast = 0
		}
	}
	return neighbors, steps
}
