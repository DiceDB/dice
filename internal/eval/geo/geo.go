// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package geo

import (
	"math"

	diceerrors "github.com/dicedb/dice/internal/errors"
)

// Earth's radius in meters
const earthRadius float64 = 6372797.560856

// Bit precision for geohash
const bitPrecision = 52
// Bit precision steps for geohash - picked up to match redis
const maxSteps = 26

// Bit precision for geohash string
const bitPrecisionString = 10
const mercatorMax = 20037726.37

const (
	/* These are constraints from EPSG:900913 / EPSG:3785 / OSGEO:41001 */
	/* We can't geocode at the north/south pole. */
	globalMinLat = -85.05112878
	globalMaxLat = 85.05112878
	globalMinLon = -180.0
	globalMaxLon = 180.0
)

type Unit string

const (
	Meters     Unit = "m"
	Kilometers Unit = "km"
	Miles      Unit = "mi"
	Feet       Unit = "ft"
)

// DegToRad converts degrees to radians.
func DegToRad(deg float64) float64 {
	return math.Pi * deg / 180.0
}

// RadToDeg converts radians to degrees.
func RadToDeg(rad float64) float64 {
	return 180.0 * rad / math.Pi
}

// GetDistance calculates the distance between two geographical points specified by their longitude and latitude.
func GetDistance(lon1, lat1, lon2, lat2 float64) float64 {
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

// GetLatDistance calculates the distance between two latitudes.
func GetLatDistance(lat1, lat2 float64) float64 {
	return earthRadius * math.Abs(DegToRad(lat2)-DegToRad(lat1))
}

// EncodeHash returns a geo hash for a given coordinate, and returns it in float64 so it can be used as score in a zset.
func EncodeHash(latitude, longitude float64) float64 {
	h := encodeHash(longitude, latitude, maxSteps)
	h = align52Bits(h, maxSteps)

	return float64(h)
}

// encodeHash encodes the latitude and longitude into a geohash with the specified number of steps.
func encodeHash(longitude, latitude float64, steps uint8) uint64 {
	latOffset := (latitude - globalMinLat) / (globalMaxLat - globalMinLat)
	longOffset := (longitude - globalMinLon) / (globalMaxLon - globalMinLon)

	latOffset *= float64(uint64(1) << steps)
	longOffset *= float64(uint64(1) << steps)
	return interleave64(uint32(latOffset), uint32(longOffset))
}

// DecodeHash returns the latitude and longitude from a geo hash.
// The hash should be a float64, as it is used as score in a sorted set.
func DecodeHash(hash float64) (lat, lon float64) {
	return decodeHash(uint64(hash), maxSteps)
}

// decodeHash decodes the geohash into latitude and longitude with the specified number of steps.
func decodeHash(hash uint64, steps uint8) (lat float64, lon float64) {
	hashSep := deinterleave64(hash)

	latScale := globalMaxLat - globalMinLat
	longScale := globalMaxLon - globalMinLon

	ilato := uint32(hashSep)       // lat part
	ilono := uint32(hashSep >> 32) // lon part

	// divide by 2**step.
	// Then, for 0-1 coordinate, multiply times scale and add
	// to the min to get the absolute coordinate.
	minLat := globalMinLat + (float64(ilato)*1.0/float64(uint64(1)<<steps))*latScale
	maxLat := globalMinLat + (float64(ilato+1)*1.0/float64(uint64(1)<<steps))*latScale
	minLon := globalMinLon + (float64(ilono)*1.0/float64(uint64(1)<<steps))*longScale
	maxLon := globalMinLon + (float64(ilono+1)*1.0/float64(uint64(1)<<steps))*longScale

	lon = (minLon + maxLon) / 2
	if lon > globalMaxLon {
		lon = globalMaxLon
	}
	if lon < globalMinLon {
		lon = globalMinLon
	}

	lat = (minLat + maxLat) / 2
	if lat > globalMaxLat {
		lat = globalMaxLat
	}
	if lat < globalMinLat {
		lat = globalMinLat
	}

	return lat, lon
}

// ConvertDistance converts a distance from meters to the desired unit.
func ConvertDistance(distance float64, unit string) (float64, error) {
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
		return 0, diceerrors.ErrUnsupportedUnit
	}
}

// ToMeters converts a distance and its unit to meters.
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

// geohashEstimateStepsByRadius estimates the number of steps required to cover a radius at a given latitude.
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

// boundingBox returns the bounding box for a given latitude, longitude and radius.
func boundingBox(lat, lon, radius float64) (float64, float64, float64, float64) {
	latDelta := RadToDeg(radius / earthRadius)
	lonDeltaTop := RadToDeg(radius / earthRadius / math.Cos(DegToRad(lat+latDelta)))
	lonDeltaBottom := RadToDeg(radius / earthRadius / math.Cos(DegToRad(lat-latDelta)))

	isSouthernHemisphere := false
	if lat < 0 {
		isSouthernHemisphere = true
	}

	minLon := lon - lonDeltaTop
	if isSouthernHemisphere {
		minLon = lon - lonDeltaBottom
	}

	maxLon := lon + lonDeltaTop
	if isSouthernHemisphere {
		maxLon = lon + lonDeltaBottom
	}

	minLat := lat - latDelta
	maxLat := lat + latDelta

	return minLon, minLat, maxLon, maxLat
}

// Area returns the geohashes of the area covered by a circle with a given radius. It returns the center hash
// and the 8 surrounding hashes. The second return value is the number of steps used to cover the area.
func Area(centerHash, radius float64) ([9]uint64, uint8) {
	var result [9]uint64

	centerLat, centerLon := decodeHash(uint64(centerHash), maxSteps)
	minLon, minLat, maxLon, maxLat := boundingBox(centerLat, centerLon, radius)
	steps := geohashEstimateStepsByRadius(radius, centerLat)
	centerRadiusHash := encodeHash(centerLon, centerLat, steps)

	neighbors := geohashNeighbors(uint64(centerRadiusHash), steps)
	area := areaBySteps(centerRadiusHash, steps)

	/* Check if the step is enough at the limits of the covered area.
	 * Sometimes when the search area is near an edge of the
	 * area, the estimated step is not small enough, since one of the
	 * north / south / west / east square is too near to the search area
	 * to cover everything. */
	north := areaBySteps(neighbors[0], steps)
	south := areaBySteps(neighbors[4], steps)
	east := areaBySteps(neighbors[2], steps)
	west := areaBySteps(neighbors[6], steps)

	decreaseStep := false
	if north.Lat.Max < maxLat || south.Lat.Min > minLat || east.Lon.Max < maxLon || west.Lon.Min > minLon {
		decreaseStep = true
	}

	if steps > 1 && decreaseStep {
		steps--
		centerRadiusHash = encodeHash(centerLat, centerLon, steps)
		neighbors = geohashNeighbors(centerRadiusHash, steps)
		area = areaBySteps(centerRadiusHash, steps)
	}

	// exclude useless areas
	if steps >= 2 {
		if area.Lat.Min < minLat {
			neighbors[3] = 0 // south east
			neighbors[4] = 0 // south
			neighbors[5] = 0 // south west
		}

		if area.Lat.Max > maxLat {
			neighbors[0] = 0 // north
			neighbors[1] = 0 // north east
			neighbors[7] = 0 // north west
		}

		if area.Lon.Min < minLon {
			neighbors[5] = 0 // south west
			neighbors[6] = 0 // west
			neighbors[7] = 0 // north west
		}

		if area.Lon.Max > maxLon {
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
	min := align52Bits(hash, steps)
	hash++
	max := align52Bits(hash, steps)
	return min, max
}

// align52Bits aligns the hash to 52 bits.
func align52Bits(hash uint64, steps uint8) uint64 {
	hash <<= (52 - steps*2)
	return hash
}

type hashRange struct {
	Min float64
	Max float64
}

type hashArea struct {
	Lat hashRange
	Lon hashRange
}

// deinterleave64 deinterleaves a 64-bit integer.
func deinterleave64(interleaved uint64) uint64 {
	x := interleaved & 0x5555555555555555
	y := (interleaved >> 1) & 0x5555555555555555

	x = (x | (x >> 1)) & 0x3333333333333333
	y = (y | (y >> 1)) & 0x3333333333333333

	x = (x | (x >> 2)) & 0x0f0f0f0f0f0f0f0f
	y = (y | (y >> 2)) & 0x0f0f0f0f0f0f0f0f

	x = (x | (x >> 4)) & 0x00ff00ff00ff00ff
	y = (y | (y >> 4)) & 0x00ff00ff00ff00ff

	x = (x | (x >> 8)) & 0x0000ffff0000ffff
	y = (y | (y >> 8)) & 0x0000ffff0000ffff

	x = (x | (x >> 16)) & 0x00000000ffffffff
	y = (y | (y >> 16)) & 0x00000000ffffffff

	return (y << 32) | x
}

// interleave64 interleaves two 32-bit integers into a 64-bit integer.
func interleave64(xlo, ylo uint32) uint64 {
	B := []uint64{
		0x5555555555555555,
		0x3333333333333333,
		0x0F0F0F0F0F0F0F0F,
		0x00FF00FF00FF00FF,
		0x0000FFFF0000FFFF,
	}
	S := []uint{1, 2, 4, 8, 16}

	x := uint64(xlo)
	y := uint64(ylo)

	x = (x | (x << S[4])) & B[4]
	y = (y | (y << S[4])) & B[4]

	x = (x | (x << S[3])) & B[3]
	y = (y | (y << S[3])) & B[3]

	x = (x | (x << S[2])) & B[2]
	y = (y | (y << S[2])) & B[2]

	x = (x | (x << S[1])) & B[1]
	y = (y | (y << S[1])) & B[1]

	x = (x | (x << S[0])) & B[0]
	y = (y | (y << S[0])) & B[0]

	return x | (y << 1)
}

// areaBySteps calculates the area covered by a hash at a given number of steps.
func areaBySteps(hash uint64, steps uint8) *hashArea {
	hashSep := deinterleave64(hash)

	latScale := globalMaxLat - globalMinLat
	longScale := globalMaxLon - globalMinLon

	ilato := uint32(hashSep)       // lat part
	ilono := uint32(hashSep >> 32) // lon part

	// divide by 2**step.
	// Then, for 0-1 coordinate, multiply times scale and add
	// to the min to get the absolute coordinate.
	area := &hashArea{}
	area.Lat.Min = globalMinLat + (float64(ilato)/float64(uint64(1)<<steps))*latScale
	area.Lat.Max = globalMinLat + (float64(ilato+1)/float64(uint64(1)<<steps))*latScale
	area.Lon.Min = globalMinLon + (float64(ilono)/float64(uint64(1)<<steps))*longScale
	area.Lon.Max = globalMinLon + (float64(ilono+1)/float64(uint64(1)<<steps))*longScale
	return area
}

// geohashMoveX moves the geohash in the x direction.
func geohashMoveX(hash uint64, steps uint8, d int8) uint64 {
	if d == 0 {
		return hash
	}

	x := hash & 0xaaaaaaaaaaaaaaaa
	y := hash & 0x5555555555555555

	zz := 0x5555555555555555 >> (64 - steps*2)

	if d > 0 {
		x = x + uint64(zz+1)
	} else {
		x = x | uint64(zz)
		x = x - uint64(zz+1)
	}

	x &= (0xaaaaaaaaaaaaaaaa >> (64 - steps*2))
	return x | y
}

// geohashMoveY moves the geohash in the y direction.
func geohashMoveY(hash uint64, steps uint8, d int8) uint64 {
	x := hash & 0xaaaaaaaaaaaaaaaa
	y := hash & 0x5555555555555555

	zz := uint64(0xaaaaaaaaaaaaaaaa) >> (64 - steps*2)

	if d > 0 {
		y = y + (zz + 1)
	} else {
		y = y | zz
		y = y - (zz + 1)
	}

	y &= (0x5555555555555555 >> (64 - steps*2))
	return x | y
}

// geohashNeighbors returns the geohash neighbors of a given hash with a given number of steps.
func geohashNeighbors(hash uint64, steps uint8) [8]uint64 {
	neighbors := [8]uint64{}

	neighbors[0] = geohashMoveY(hash, steps, 1)                           // North
	neighbors[1] = geohashMoveX(geohashMoveY(hash, steps, 1), steps, 1)   // North-East
	neighbors[2] = geohashMoveX(hash, steps, 1)                           // East
	neighbors[3] = geohashMoveX(geohashMoveY(hash, steps, -1), steps, 1)  // South-East
	neighbors[4] = geohashMoveY(hash, steps, -1)                          // South
	neighbors[5] = geohashMoveX(geohashMoveY(hash, steps, -1), steps, -1) // South-West
	neighbors[6] = geohashMoveX(hash, steps, -1)                          // West
	neighbors[7] = geohashMoveX(geohashMoveY(hash, steps, 1), steps, -1)  // North-West

	return neighbors
}
