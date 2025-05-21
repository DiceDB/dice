package geoUtil

import (
	"math"

	"github.com/dicedb/dice/internal/types"
)

// This represents a shape in which we have to find nodes
type GeoShape interface {
	GetBoudingBoxWidhtAndHeight() (float64, float64)
	GetRadius() float64
	GetLonLat() (float64, float64)
	GetDistanceIfWithinShape(lon float64, lat float64) (distance float64)
}

// GeoShapeCircle Implementation of GeoShape
type GeoShapeCircle struct {
	Radius    float64
	Longitude float64
	Latitude  float64
	Unit      types.Param
}

func (circle *GeoShapeCircle) GetBoudingBoxWidhtAndHeight() (float64, float64) {
	return circle.Radius * 2, circle.Radius * 2
}

func (circle *GeoShapeCircle) GetRadius() float64 {
	return circle.Radius
}

func (circle *GeoShapeCircle) GetLonLat() (float64, float64) {
	return circle.Longitude, circle.Latitude
}

func (circle *GeoShapeCircle) GetDistanceIfWithinShape(lon float64, lat float64) (distance float64) {
	distance = GetDistance(lon, lat, circle.Longitude, circle.Latitude)
	if distance > circle.Radius {
		return 0
	}
	return distance
}

func GetNewGeoShapeCircle(radius float64, longitude float64, latitude float64, unit types.Param) (*GeoShapeCircle, error) {
	var convRadius float64
	var err error
	if convRadius, err = ConvertToMeter(radius, unit); err != nil {
		return nil, err
	}
	return &GeoShapeCircle{
		Radius:    convRadius,
		Longitude: longitude,
		Latitude:  latitude,
		Unit:      unit,
	}, nil
}

// GeoShapeRectangle Implementation of GeoShape
type GeoShapeRectangle struct {
	Widht     float64
	Height    float64
	Longitude float64
	Latitude  float64
	Unit      types.Param
}

func (rec *GeoShapeRectangle) GetBoudingBoxWidhtAndHeight() (float64, float64) {
	return rec.Widht, rec.Height
}

func (rec *GeoShapeRectangle) GetRadius() float64 {
	radius := math.Sqrt((rec.Widht/2)*(rec.Widht/2) + (rec.Height/2)*(rec.Height/2))
	return radius
}

func (rec *GeoShapeRectangle) GetLonLat() (float64, float64) {
	return rec.Longitude, rec.Latitude
}

func (rec *GeoShapeRectangle) GetDistanceIfWithinShape(lon float64, lat float64) (distance float64) {
	// latitude distance is less expensive to compute than longitude distance
	// so we check first for the latitude condition
	latDistance := GetLatDistance(lat, rec.Latitude)
	if latDistance > rec.Height/2 {
		return 0
	}

	lonDistance := GetDistance(lon, lat, rec.Longitude, lat)
	if lonDistance > rec.Widht/2 {
		return 0
	}

	distance = GetDistance(lon, lat, rec.Longitude, rec.Latitude)

	return distance
}

func GetNewGeoShapeRectangle(widht float64, height float64, longitude float64, latitude float64, unit types.Param) (*GeoShapeRectangle, error) {
	var convWidth, convHeight float64
	var err error
	if convWidth, err = ConvertToMeter(widht, unit); err != nil {
		return nil, err
	}
	if convHeight, err = ConvertToMeter(height, unit); err != nil {
		return nil, err
	}
	return &GeoShapeRectangle{
		Widht:     convWidth,
		Height:    convHeight,
		Longitude: longitude,
		Latitude:  latitude,
		Unit:      unit,
	}, nil
}
