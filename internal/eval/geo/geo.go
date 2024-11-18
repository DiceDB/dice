package geo

import (
	"math"

	"github.com/dicedb/dice/internal/errors"
	"github.com/mmcloughlin/geohash"
)

// Earth's radius in meters
const earthRadius float64 = 6372797.560856

// Bit precision for geohash - picked up to match redis
const bitPrecision = 52

const mercatorMax = 20037726.37

const (
	minLat = -85.05112878
	maxLat = 85.05112878
	minLon = -180
	maxLon = 180
)

type Unit string

const (
	Meters     Unit = "m"
	Kilometers Unit = "km"
	Miles      Unit = "mi"
	Feet       Unit = "ft"
)

func DegToRad(deg float64) float64 {
	return math.Pi * deg / 180.0
}

func RadToDeg(rad float64) float64 {
	return 180.0 * rad / math.Pi
}

func GetDistance(
	lon1,
	lat1,
	lon2,
	lat2 float64,
) float64 {
	lon1r := DegToRad(lon1)
	lon2r := DegToRad(lon2)
	v := math.Sin((lon2r - lon1r) / 2)
	// if v == 0 we can avoid doing expensive math when lons are practically the same
	if v == 0.0 {
		return GetLatDistance(lat1, lat2)
	}

	lat1r := DegToRad(lat1)
	lat2r := DegToRad(lat2)
	u := math.Sin((lat2r - lat1r) / 2)

	a := u*u + math.Cos(lat1r)*math.Cos(lat2r)*v*v

	return 2.0 * earthRadius * math.Asin(math.Sqrt(a))
}

func GetLatDistance(lat1, lat2 float64) float64 {
	return earthRadius * math.Abs(DegToRad(lat2)-DegToRad(lat1))
}

// EncodeHash returns a geo hash for a given coordinate, and returns it in float64 so it can be used as score in a zset
func EncodeHash(
	latitude,
	longitude float64,
) float64 {
	h := geohash.EncodeIntWithPrecision(latitude, longitude, bitPrecision)

	return float64(h)
}

// DecodeHash returns the latitude and longitude from a geo hash
// The hash should be a float64, as it is used as score in a zset
func DecodeHash(hash float64) (lat, lon float64) {
	lat, lon = geohash.DecodeIntWithPrecision(uint64(hash), bitPrecision)

	return lat, lon
}

// ConvertDistance converts a distance from meters to the desired unit
func ConvertDistance(
	distance float64,
	unit string,
) (converted float64, err []byte) {
	switch Unit(unit) {
	case Meters:
		return distance, nil
	case Kilometers:
		return distance / 1000, nil
	case Miles:
		return distance / 1609.34, nil
	case Feet:
		return distance / 0.3048, nil
	default:
		return 0, errors.NewErrWithMessage("ERR unsupported unit provided. please use m, km, ft, mi")
	}
}

// ToMeters converts a distance and its unit to meters
func ToMeters(distance float64, unit string) (float64, bool) {
	switch Unit(unit) {
	case Meters:
		return distance, true
	case Kilometers:
		return distance * 1000, true
	case Miles:
		return distance * 1609.34, true
	case Feet:
		return distance * 0.3048, true
	default:
		return 0, false
	}
}

func geohashEstimateStepsByRadius(radius, lat float64) uint8 {
	if radius == 0 {
		return 26
	}

	step := 1
	for radius < mercatorMax {
		radius *= 2
		step++
	}
	step -= 2 // Make sure range is included in most of the base cases.

	/* Note from the redis implementation:
		Wider range towards the poles... Note: it is possible to do better
	     than this approximation by computing the distance between meridians
	     	at this latitude, but this does the trick for now. */
	if lat > 66 || lat < -66 {
		step--
		if lat > 80 || lat < -80 {
			step--
		}
	}

	if step < 1 {
		step = 1
	}
	if step > 26 {
		step = 26
	}

	return uint8(step)
}

// Area returns the geohashes of the area covered by a circle with a given radius. It returns the center hash
// and the 8 surrounding hashes. The second return value is the number of steps used to cover the area.
func Area(centerHash, radius float64) ([9]uint64, uint8) {
	var result [9]uint64

	centerLat, centerLon := DecodeHash(centerHash)

	steps := geohashEstimateStepsByRadius(radius, centerLat)

	centerRadiusHash := geohash.EncodeIntWithPrecision(centerLat, centerLon, uint(steps)*2)

	neighbors := geohash.NeighborsIntWithPrecision(centerRadiusHash, uint(steps)*2)
	area := geohash.BoundingBoxInt(centerRadiusHash)

	/* Check if the step is enough at the limits of the covered area.
	 * Sometimes when the search area is near an edge of the
	 * area, the estimated step is not small enough, since one of the
	 * north / south / west / east square is too near to the search area
	 * to cover everything. */
	north := geohash.BoundingBoxInt(neighbors[0])
	east := geohash.BoundingBoxInt(neighbors[2])
	south := geohash.BoundingBoxInt(neighbors[4])
	west := geohash.BoundingBoxInt(neighbors[6])

	decreaseStep := false
	if north.MaxLat < maxLat || south.MinLat < minLat || east.MaxLng < maxLon || west.MinLng < minLon {
		decreaseStep = true
	}

	if steps > 1 && decreaseStep {
		steps--
		centerRadiusHash = geohash.EncodeIntWithPrecision(centerLat, centerLon, uint(steps)*2)
		neighbors = geohash.NeighborsIntWithPrecision(centerRadiusHash, uint(steps)*2)
		area = geohash.BoundingBoxInt(centerRadiusHash)
	}

	// exclude useless areas
	if steps >= 2 {
		if area.MinLat < minLat {
			neighbors[3] = 0 // south east
			neighbors[4] = 0 // south
			neighbors[5] = 0 // south west
		}

		if area.MaxLat > maxLat {
			neighbors[0] = 0 // north
			neighbors[1] = 0 // north east
			neighbors[7] = 0 // north west
		}

		if area.MinLng < minLon {
			neighbors[5] = 0 // south west
			neighbors[6] = 0 // west
			neighbors[7] = 0 // north west
		}

		if area.MaxLng > maxLon {
			neighbors[1] = 0 // north east
			neighbors[2] = 0 // east
			neighbors[3] = 0 // south east
		}
	}

	result[0] = centerRadiusHash
	for i := 0; i < len(neighbors); i++ {
		result[i+1] = neighbors[i]
	}

	return result, steps
}

// HashMinMax returns the min and max hashes for a given hash and steps. This can be used to get the range of hashes
// that a given hash and a radius (steps) will cover.
func HashMinMax(hash uint64, steps uint8) (uint64, uint64) {
	min := geohashAlign52Bits(hash, steps)
	hash++
	max := geohashAlign52Bits(hash, steps)
	return min, max
}

func geohashAlign52Bits(hash uint64, steps uint8) uint64 {
	hash <<= (52 - steps*2)
	return hash
}
