package types

import (
	"math"

	"github.com/dicedb/dice/internal/errors"
	"github.com/mmcloughlin/geohash"
)

// This represents a shape in which we have to find nodes
type GeoShape interface {
	GetBoudingBoxWidhtAndHeight() (float64, float64)
	GetRadius() float64
	GetCoordinate() *GeoCoordinate
	GetDistanceIfWithinShape(coordinate *GeoCoordinate) (distance float64)
}

// GeoShapeCircle Implementation of GeoShape
type GeoShapeCircle struct {
	Radius           float64
	CenterCoordinate *GeoCoordinate
	Unit             Param
}

func (circle *GeoShapeCircle) GetBoudingBoxWidhtAndHeight() (float64, float64) {
	return circle.Radius * 2, circle.Radius * 2
}

func (circle *GeoShapeCircle) GetRadius() float64 {
	return circle.Radius
}

func (circle *GeoShapeCircle) GetCoordinate() *GeoCoordinate {
	return circle.CenterCoordinate
}

func (circle *GeoShapeCircle) GetDistanceIfWithinShape(coordinate *GeoCoordinate) (distance float64) {
	distance = coordinate.GetDistanceFromCoordinate(circle.CenterCoordinate)
	if distance > circle.Radius {
		return 0
	}
	distance, _ = ConvertDistance(distance, circle.Unit)
	return distance
}

func GetNewGeoShapeCircle(radius float64, centerCoordinate *GeoCoordinate, unit Param) (*GeoShapeCircle, error) {
	var convRadius float64
	var errRad error
	if radius <= 0 {
		return nil, errors.ErrGeneral("RADIUS should be > 0")
	}

	if convRadius, errRad = ConvertToMeter(radius, unit); errRad != nil {
		return nil, errRad
	}

	longitude, latitude := centerCoordinate.Longitude, centerCoordinate.Latitude

	var coord *GeoCoordinate
	var errCoord error
	if coord, errCoord = NewGeoCoordinateFromLonLat(longitude, latitude); errCoord != nil {
		return nil, errCoord
	}

	return &GeoShapeCircle{
		Radius:           convRadius,
		CenterCoordinate: coord,
		Unit:             unit,
	}, nil
}

// GeoShapeRectangle Implementation of GeoShape
type GeoShapeRectangle struct {
	Widht            float64
	Height           float64
	CenterCoordinate *GeoCoordinate
	Unit             Param
}

func (rec *GeoShapeRectangle) GetBoudingBoxWidhtAndHeight() (float64, float64) {
	return rec.Widht, rec.Height
}

func (rec *GeoShapeRectangle) GetRadius() float64 {
	radius := math.Sqrt((rec.Widht/2)*(rec.Widht/2) + (rec.Height/2)*(rec.Height/2))
	return radius
}

func (rec *GeoShapeRectangle) GetCoordinate() *GeoCoordinate {
	return rec.CenterCoordinate
}

func (rec *GeoShapeRectangle) GetDistanceIfWithinShape(coordinate *GeoCoordinate) (distance float64) {
	// latitude distance is less expensive to compute than longitude distance
	// so we check first for the latitude condition
	latDistance := coordinate.GetLatDistanceFromCoordinate(rec.CenterCoordinate)
	if latDistance > rec.Height/2 {
		return 0
	}

	// Creating a coordinate with same latitude, to get longitude distance
	sameLatitudeCoord, _ := NewGeoCoordinateFromLonLat(rec.CenterCoordinate.Longitude, coordinate.Latitude)
	lonDistance := coordinate.GetDistanceFromCoordinate(sameLatitudeCoord)
	if lonDistance > rec.Widht/2 {
		return 0
	}

	distance = coordinate.GetDistanceFromCoordinate(rec.CenterCoordinate)
	distance, _ = ConvertDistance(distance, rec.Unit)
	return distance
}

func GetNewGeoShapeRectangle(widht float64, height float64, centerCoordinate *GeoCoordinate, unit Param) (*GeoShapeRectangle, error) {
	var convWidth, convHeight float64
	var err error

	if widht <= 0 || height <= 0 {
		return nil, errors.ErrGeneral("HEIGHT, WIDTH should be > 0")
	}
	if convWidth, err = ConvertToMeter(widht, unit); err != nil {
		return nil, err
	}
	if convHeight, err = ConvertToMeter(height, unit); err != nil {
		return nil, err
	}

	longitude, latitude := centerCoordinate.Longitude, centerCoordinate.Latitude

	var coord *GeoCoordinate
	var errCoord error
	if coord, errCoord = NewGeoCoordinateFromLonLat(longitude, latitude); errCoord != nil {
		return nil, errCoord
	}

	return &GeoShapeRectangle{
		Widht:            convWidth,
		Height:           convHeight,
		CenterCoordinate: coord,
		Unit:             unit,
	}, nil
}

// Return the bounding box of the shape
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
func GetBoundingBoxForShape(geoShape GeoShape) *geohash.Box {

	coord := geoShape.GetCoordinate()
	lon, lat := coord.Longitude, coord.Latitude
	width, height := geoShape.GetBoudingBoxWidhtAndHeight()
	boudingBox := geohash.Box{}

	width /= 2
	height /= 2
	latDelta := RadToDeg(height / EARTH_RADIUS)
	lonDeltaTop := RadToDeg(width / EARTH_RADIUS / math.Cos(DegToRad(lat+latDelta)))
	lonDeltaBottom := RadToDeg(width / EARTH_RADIUS / math.Cos(DegToRad(lat-latDelta)))

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

// Find the step → Precision at which 9 cells (3x3 cells) can cover the entire given shape
func EstimatePrecisionForShapeCoverage(geoShape GeoShape) uint {

	radius := geoShape.GetRadius()
	coord := geoShape.GetCoordinate()
	_, latitude := coord.Longitude, coord.Latitude

	if radius == 0 {
		return 26
	}

	var step int = 1
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

	// as (mmcloughlin/geohash) requires total number of bits, not steps, so we multiply by 2
	return uint(step * 2)
}

func GetGeohashNeighborsWithinShape(geoShape GeoShape, boudingBox *geohash.Box) (neighbors *Neighbors) {
	coord := geoShape.GetCoordinate()
	lon, lat := coord.Longitude, coord.Latitude
	steps := uint(2 * EstimatePrecisionForShapeCoverage(geoShape))

	centerHash := geohash.EncodeIntWithPrecision(lat, lon, steps)
	centerBox := geohash.BoundingBoxIntWithPrecision(centerHash, steps)
	neighborsArr := geohash.NeighborsIntWithPrecision(centerHash, steps)
	neighborsArr = append(neighborsArr, centerHash)
	neighbors = CreateNeighborsFromArray(neighborsArr)

	// Check if the step is enough at the limits of the covered area.
	// Decode each of the 8 neighbours to get max and min (lon, lat)
	// If North.maxLatitude < maxLatitude(from bouding box) then we have to reduce step to increase neighbour size
	// Do this for N, S, E, W
	northBox := geohash.BoundingBoxIntWithPrecision(neighbors.North, steps)
	eastBox := geohash.BoundingBoxIntWithPrecision(neighbors.East, steps)
	southBox := geohash.BoundingBoxIntWithPrecision(neighbors.South, steps)
	westBox := geohash.BoundingBoxIntWithPrecision(neighbors.West, steps)

	if northBox.MaxLat < boudingBox.MaxLat || southBox.MinLat > boudingBox.MinLat || eastBox.MaxLng < boudingBox.MaxLng || westBox.MinLng > boudingBox.MinLng {
		steps -= 2
		centerHash = geohash.EncodeIntWithPrecision(lat, lon, steps)
		centerBox = geohash.BoundingBoxIntWithPrecision(centerHash, steps)
		neighborsArr := geohash.NeighborsIntWithPrecision(centerHash, steps)
		neighborsArr = append(neighborsArr, centerHash)
		neighbors = CreateNeighborsFromArray(neighborsArr)
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

	// Set Steps in neighbors
	neighbors.Steps = steps
	return neighbors
}
