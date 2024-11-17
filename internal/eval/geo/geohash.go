package geo

import (
	"fmt"
	"strings"
)

type Coordinates struct {
	Longitude float64
	Latitude float64
}

type GoeHash struct {
	Hash string
	Neighbors []GoeHash
}

var (
	// base32 alphabets
	base32 []string = []string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	"b", "c", "d", "e", "f", "g", "h", "j", "k", "m",
	"n", "p", "q", "r", "s", "t", "u", "v", "w", "x",
	"y", "z",
	}

	// This coordinate measures north-south position and ranges from -90° (South Pole) to +90° (North Pole).
	latitudeRange = []int64{90,-90}

	// This coordinate measures east-west position and ranges from -180° (Western Hemisphere) to +180° (Eastern Hemisphere).
	longitudeRange = []int64{180,-180}

	//
	MIN_LATITUDE 	= -85.05112878

	//
	MAX_LATITUDE 	= 85.05112878

	//
	MIN_LONGITUDE = -180

	//
	MAX_LONGITUDE = 180


	ErrInvalidInput = func (lat, long float64) error {
		return fmt.Errorf("invalid positions provided (lat: %.4f, long: %.4f)", lat, long)
	}

	ErrInvalidHash = func (hash string) error {
		return fmt.Errorf("invalid positions provided (hash: %s)", hash)
	}
)


func Encode(lat, long float64, precision int) (string, error) {

	if precision <= 0 {
		precision = 12;
	}

	if long > float64(MAX_LONGITUDE) || long < float64(MIN_LONGITUDE) || lat < MIN_LATITUDE || lat > MAX_LATITUDE {
		return "", ErrInvalidInput(lat, long)
	}


	// convert them to thier binary equivalent.
	latBin := getBinaryString(latitudeRange,lat, precision)
	longBin := getBinaryString(longitudeRange, long, precision)

	fmt.Println("lat: ", latBin)
	fmt.Println("long: ", longBin)


	return "", nil

}

func Decode(hash string) (Coordinates, error) {

	return Coordinates{}, nil
}


func getBinaryString(cRange []int64, target float64, precision int) string {
	var stringBuilder = strings.Builder{}

	max := float64(cRange[0])
	min := float64(cRange[1])

	for i := 0; i < precision; i++ {
		mid := (float64(max) + float64(min)) / 2

		if(target == float64(mid)) {
			stringBuilder.WriteString("1")
			min = mid;
		}else if (target <= float64(mid)) {
			stringBuilder.WriteString("0")
			max = mid
		}else if (target >= float64(mid)) {
			stringBuilder.WriteString("1")
			min = mid
		}
	}

	return stringBuilder.String()
}