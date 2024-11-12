package eval

import (
	"crypto/rand"
	"math"
	"math/big"
	"strconv"

	diceerrors "github.com/dicedb/dice/internal/errors"
)

// util function for incrementing a value with a int64 increment
// parses the value to int64 internally and returns
// error if any
func IncrementInt(val, incr int64) (int64, error) {
	if (incr < 0 && val < 0 && incr < (math.MinInt64-val)) ||
		(incr > 0 && val > 0 && incr > (math.MaxInt64-val)) {
		return -1, diceerrors.NewErr(diceerrors.IncrDecrOverflowErr)
	}
	return val + incr, nil
}

func IncrementFloat(val, incr float64) (float64, error) {
	if math.IsInf(val+incr, 1) || math.IsInf(val+incr, -1) {
		return -1, diceerrors.NewErr(diceerrors.IncrDecrOverflowErr)
	}
	return val + incr, nil
}

func hashRandomFields(hs Hash, count int, withValues bool) *[]string {
	items := hs.Items()

	var results []string
	resultSet := make(map[string]struct{})

	abs := func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}

	for i := 0; i < abs(count); i++ {
		if count > 0 && len(resultSet) == len(items) {
			break
		}

		randomIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(items))))
		randomField := items[randomIndex.Int64()][0]

		if count > 0 {
			if _, exists := resultSet[randomField]; exists {
				i--
				continue
			}
			resultSet[randomField] = struct{}{}
		}

		results = append(results, randomField)
		if withValues {
			results = append(results, items[randomIndex.Int64()][1])
		}
	}
	return &results
}

func hashIncrementValue(oldVal *string, incr int64) (int64, error) {
	intOldVal, err := strconv.ParseInt(*oldVal, 10, 64)
	if err != nil {
		return -1, diceerrors.NewErr(diceerrors.HashValueNotIntegerErr)
	}
	newVal, err := IncrementInt(intOldVal, incr)
	if err != nil {
		return -1, err
	}
	return newVal, nil
}

func hashIncrementFloatValue(oldVal *string, incr float64) (*string, error) {
	floatOldVal, err := strconv.ParseFloat(*oldVal, 64)
	if err != nil {
		return nil, diceerrors.ErrInvalidNumberFormat
	}
	newVal, err := IncrementFloat(floatOldVal, incr)
	if err != nil {
		return nil, diceerrors.ErrGeneral(err.Error())
	}
	newValStr := strconv.FormatFloat(newVal, 'f', -1, 64)
	return &newValStr, nil
}
