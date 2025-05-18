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

// Limits from EPSG:900913 / EPSG:3785 / OSGEO:41001
const LAT_MIN float64 = -85.05112878
const LAT_MAX float64 = 85.05112878
const LONG_MIN float64 = -180
const LONG_MAX float64 = 180

func ValidateLonLat(longitude, latitude float64) error {
	if latitude > LAT_MAX || latitude < LAT_MIN || longitude > LONG_MAX || longitude < LONG_MIN {
		return errors.ErrInvalidLonLatPair(longitude, latitude)
	}
	return nil
}

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
	return EARTH_RADIUS * math.Abs(lat2-lat1)
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

	// Round to 4 decimal places
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
func GetBoundingBoxWithHash(hash uint64, radius float64) [4]float64 {

	boudingBox := [4]float64{}
	lon, lat := DecodeHash(hash)
	latDelta := RadToDeg(radius / EARTH_RADIUS)
	lonDeltaTop := RadToDeg(radius / EARTH_RADIUS / math.Cos(DegToRad(lat+latDelta)))
	lonDeltaBottom := RadToDeg(radius / EARTH_RADIUS / math.Cos(DegToRad(lat-latDelta)))

	boudingBox[1] = lat - latDelta
	boudingBox[3] = lat + latDelta

	if lat < 0 {
		boudingBox[0] = lon - lonDeltaBottom
		boudingBox[2] = lon + lonDeltaBottom
	} else {
		boudingBox[0] = lon - lonDeltaTop
		boudingBox[2] = lon + lonDeltaTop
	}

	return boudingBox

}

func GetBoundingBoxWithLonLat(lon, lat float64, radius float64) [4]float64 {
	return GetBoundingBoxWithHash(EncodeHash(lon, lat), radius)
}

func parseParams() {

}
