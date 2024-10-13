package utils

import (
	"math"

	diceerrors "github.com/dicedb/dice/internal/errors"
)

// util function for incrementing a value with a int64 increment
// parses the value to int64 internally and returns
// error if any
func IncrementInt(val int64, incr int64) (int64, error) {
	if (incr < 0 && val < 0 && incr < (math.MinInt64-val)) ||
		(incr > 0 && val > 0 && incr > (math.MaxInt64-val)) {
		return -1, diceerrors.NewErr(diceerrors.IncrDecrOverflowErr)
	}
	return val + incr, nil
}

func IncrementFloat(val float64, incr float64) (float64, error) {
	if math.IsInf(val+incr, 1) || math.IsInf(val+incr, -1) {
		return -1, diceerrors.NewErr(diceerrors.IncrDecrOverflowErr)
	}
	return val + incr, nil
}
