package eval

import (
	"errors"
	"strconv"
	"strings"

	"github.com/dicedb/dice/internal/clientio"
	diceerrors "github.com/dicedb/dice/internal/errors"
	dstore "github.com/dicedb/dice/internal/store"
)

func evalBITPOS(args []string, store *dstore.Store) []byte {
	if len(args) < 2 || len(args) > 5 {
		return diceerrors.NewErrArity("BITPOS")
	}

	key := args[0]
	obj := store.Get(key)

	bitToFind, err := parseBitToFind(args[1])
	if err != nil {
		return diceerrors.NewErrWithMessage(err.Error())
	}

	if obj == nil {
		if bitToFind == 0 {
			return clientio.Encode(0, true)
		}

		return clientio.Encode(-1, true)
	}

	byteSlice, err := getValueAsByteSlice(obj)
	if err != nil {
		return diceerrors.NewErrWithMessage(err.Error())
	}

	start, end, rangeType, endRangeProvided, err := parseOptionalParams(args[2:], len(byteSlice))
	if err != nil {
		return diceerrors.NewErrWithMessage(err.Error())
	}

	result := getBitPos(byteSlice, bitToFind, start, end, rangeType, endRangeProvided)

	return clientio.Encode(result, true)
}

func parseBitToFind(arg string) (byte, error) {
	bitToFindInt, err := strconv.Atoi(arg)
	if err != nil {
		return 0, errors.New("value is not an integer or out of range")
	}

	if bitToFindInt != 0 && bitToFindInt != 1 {
		return 0, errors.New("the bit argument must be 1 or 0")
	}

	return byte(bitToFindInt), nil
}

func parseOptionalParams(args []string, byteLen int) (start, end int, rangeType string, endRangeProvided bool, err error) {
	start, end, rangeType = 0, byteLen-1, BYTE
	endRangeProvided = false

	if len(args) > 0 {
		start, err = strconv.Atoi(args[0])
		if err != nil {
			return 0, 0, "", false, errors.New("value is not an integer or out of range")
		}
	}

	if len(args) > 1 {
		end, err = strconv.Atoi(args[1])
		if err != nil {
			return 0, 0, "", false, errors.New("value is not an integer or out of range")
		}
		endRangeProvided = true
	}

	if len(args) > 2 {
		rangeType = strings.ToUpper(args[2])
		if rangeType != BYTE && rangeType != BIT {
			return 0, 0, "", false, errors.New("syntax error")
		}
	}
	return start, end, rangeType, endRangeProvided, err
}

func getBitPos(byteSlice []byte, bitToFind byte, start, end int, rangeType string, endRangeProvided bool) int {
	byteLen := len(byteSlice)
	bitLen := len(byteSlice) * 8

	var result int

	if rangeType == BYTE {
		// Adjust start and end for both BYTE and BIT ranges
		// This handles negative indices and ensures we're within bounds
		start, end = adjustBitPosSearchRange(start, end, byteLen)

		// If start is beyond end or byteLen, we can't find anything
		if start > end || start >= byteLen {
			return -1
		}

		result = getBitPosWithBitRange(byteSlice, bitToFind, start*8, end*8+7)
	} else {
		// Adjust start and end for both BYTE and BIT ranges
		// This handles negative indices and ensures we're within bounds
		start, end = adjustBitPosSearchRange(start, end, bitLen)

		// If start is beyond end or byteLen, we can't find anything
		if start > end || start >= bitLen {
			return -1
		}

		result = getBitPosWithBitRange(byteSlice, bitToFind, start, end)
	}

	// Special case: if we're looking for a 0 bit, didn't find it,
	// and no end range was provided, we return the first bit position
	// that's not part of the byte slice (i.e., the total bit length)
	if bitToFind == 0 && result == -1 && !endRangeProvided {
		return bitLen
	}

	return result
}

func adjustBitPosSearchRange(start, end, byteLen int) (newStart, newEnd int) {
	if start < 0 {
		start += byteLen
	}
	if end < 0 {
		end += byteLen
	}
	start = max(0, start)
	end = min(byteLen-1, end)

	return start, end
}

func getBitPosWithBitRange(byteSlice []byte, bitToFind byte, start, end int) int {
	for i := start; i <= end; i++ {
		// Calculate which byte and bit we're looking at
		byteIndex := i / 8
		// 7 - (i % 8) because we count bits from left to right in each byte
		bitIndex := 7 - (i % 8)

		// Check if this bit matches what we're looking for
		if ((byteSlice[byteIndex] >> bitIndex) & 1) == bitToFind {
			return i
		}
	}

	// Bit not found in the range
	return -1
}
