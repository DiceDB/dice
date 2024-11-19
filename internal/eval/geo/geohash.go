package geo

import (
	"fmt"
	"math"
)

type Coordinates struct {
	Long float64
	Lat float64
}

type Neighbours struct {
	SouthWest string
	South string
	SouthEast string
	West string
	East string
	NorthWest string
	North string
	NorthEast string
}

const (
	MIN_LONGITUDE = -180.0
	MAX_LONGITUDE = 180.0
	MIN_LATITUDE  = -90.0
	MAX_LATITUDE  = 90.0
)


var (
	// base32 alphabets
	base32 = [32]byte{
	'0', '1', '2', '3', '4', '5', '6', '7',
	'8', '9', 'b', 'c', 'd', 'e', 'f', 'g',
	'h', 'j', 'k', 'm', 'n', 'p', 'q', 'r',
	's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	}

	ErrInvalidInput = func (lat, long float64) error {
		return fmt.Errorf("invalid positions provided (lat: %.4f, long: %.4f)", lat, long)
	}

	ErrInvalidHash = func (hash string) error {
		return fmt.Errorf("invalid positions provided (hash: %s)", hash)
	}
)


func EncodeGeoHash(long, lat float64, precision int) (string, error) {
	// validate coordinates provided
	if long < MIN_LONGITUDE || long > MAX_LONGITUDE || lat < MIN_LATITUDE || lat > MAX_LATITUDE {
		return "", fmt.Errorf("latitude or longitude out of valid range")
	}

	// Local copies of ranges to prevent global modification
	latRange := []float64{-90.0, 90.0}
	longRange := []float64{-180.0, 180.0}

	xbits := make([]bool, 0, precision*5/2)
	ybits := make([]bool, 0, precision*5/2)

	for i := 0; i < precision*5/2; i++ {
		if long >= longRange[1] {
			xbits = append(xbits, true)
			longRange[0] = longRange[1]
		} else {
			xbits = append(xbits, false)
			longRange[1] = (longRange[0] + longRange[1]) / 2.0
		}

		if lat >= latRange[1] {
			ybits = append(ybits, true)
			latRange[0] = latRange[1]
		} else {
			ybits = append(ybits, false)
			latRange[1] = (latRange[0] + latRange[1]) / 2.0
		}
	}

	// Interleave xbits and ybits, but ensure not to go out of bounds
	zbits := make([]bool, 0, 64)
	for i := 0; i < len(xbits) && i < len(ybits); i++ {
		zbits = append(zbits, xbits[i])
		zbits = append(zbits, ybits[i])
	}

	// If zbits is not a multiple of 5, pad with false (0) bits to make it divisible by 5
	for len(zbits)%5 != 0 {
		zbits = append(zbits, false)
	}

	// Convert 5-bit chunks to BASE32_CODES characters
	var geohash []byte
	for i := 0; i < len(zbits); i += 5 {
		var bitChunk uint8
		for j := 0; j < 5; j++ {
			if zbits[i+j] {
				bitChunk |= 1 << (4 - j)
			}
		}
		geohash = append(geohash, base32[bitChunk])
	}

	return string(geohash), nil
}

func DecodeGeoHash(geohash string) Coordinates {
	xs := []float64{-180.0, 180.0}
	ys := []float64{-90.0, 90.0}
	isEven := true

	for i := 0; i < len(geohash); i++ {
		c := geohash[i]
		val := getBase32Value(c)

		for j := 4; j >= 0; j-- {
			bit := (val >> j) & 1
			if isEven {
				if bit == 1 {
					xs[0] = (xs[0] + xs[1]) / 2
				} else {
					xs[1] = (xs[0] + xs[1]) / 2
				}
			} else {
				if bit == 1 {
					ys[0] = (ys[0] + ys[1]) / 2
				} else {
					ys[1] = (ys[0] + ys[1]) / 2
				}
			}
			isEven = !isEven
		}
	}

	return Coordinates{
		Long: (xs[0] + xs[1]) / 2,
		Lat: (ys[0] + ys[1]) / 2,
	}
}

func GetNeighbours(geohash string) (Neighbours, error) {
	coord := DecodeGeoHash(geohash)

	lat, lng := coord.Lat, coord.Long

	// Define latitude and longitude offsets for each direction
	latOffset := []float64{1, -1, 1, 0, 0, -1, 1, -1}
	lngOffset := []float64{0, 0, 1, -1, 1, 0, 1, -1}

	directions := []string{"SW", "S", "SE", "W", "E", "NW", "N", "NE"}

	neighbors := Neighbours{}

	
	for i, direction := range directions {
		newLat := lat + latOffset[i]*getLatError(len(geohash))
		newLng := lng + lngOffset[i]*getLngError(len(geohash))
		neighborGeohash, err := EncodeGeoHash(newLat, newLng, len(geohash))
		if err != nil {
			return neighbors, err
		}

		// Set the appropriate field in the struct
		switch direction {
		case "SW":
			neighbors.SouthWest = neighborGeohash
		case "S":
			neighbors.South = neighborGeohash
		case "SE":
			neighbors.SouthEast = neighborGeohash
		case "W":
			neighbors.West = neighborGeohash
		case "E":
			neighbors.East = neighborGeohash
		case "NW":
			neighbors.NorthWest = neighborGeohash
		case "N":
			neighbors.North = neighborGeohash
		case "NE":
			neighbors.NorthEast = neighborGeohash
		}
	}

	return neighbors, nil
}


func getBase32Value(c byte) uint8 {
	for i, code := range base32 {
		if code == c {
			return uint8(i)
		}
	}
	return 0
}

func getLatError(length int) float64 {
	return 90.0 / math.Pow(2, float64(length*5/2))
}

// Function to get longitude error based on geohash length
func getLngError(length int) float64 {
	return 180.0 / math.Pow(2, float64((length*5+1)/2))
}