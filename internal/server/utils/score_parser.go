// utils/score_parser.go
package utils

import (
	"math"
	"strconv"
)

// ParseScore converts a string to a float64, handling special cases like -inf and +inf.
func ParseScore(input string) (float64, error) {
	switch input {
	case "-inf":
		return math.Inf(-1), nil
	case "+inf":
		return math.Inf(1), nil
	default:
		return strconv.ParseFloat(input, 64)
	}
}
