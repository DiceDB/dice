package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/dicedb/dice/core/diceerrors"

	"github.com/dicedb/dice/core/bit"
)

func evalBITPOS(args []string, store *Store) []byte {
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
			return Encode(0, true)
		}

		return Encode(-1, true)
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

	return Encode(result, true)
}

func parseBitToFind(arg string) (byte, error) {
	bitToFindInt, err := strconv.Atoi(arg)
	if err != nil {
		return 0, errors.New("value is not an integer or out of range")
	}

	if bitToFindInt != 0 && bitToFindInt != 1 {
		//nolint: stylecheck
		return 0, errors.New("The bit argument must be 1 or 0")
	}

	return byte(bitToFindInt), nil
}

func getValueAsByteSlice(obj *Obj) ([]byte, error) {
	oType, oEnc := ExtractTypeEncoding(obj)
	switch oType {
	case ObjTypeInt:
		return []byte(strconv.FormatInt(obj.Value.(int64), 10)), nil
	case ObjTypeString:
		return getStringValueAsByteSlice(obj, oEnc)
	// TODO: Have this case as SETBIT stores values encoded as byte arrays. Need to investigate this further.
	case ObjTypeByteArray:
		return getByteArrayValueAsByteSlice(obj)
	default:
		return nil, fmt.Errorf("ERR unsopported type")
	}
}

func getStringValueAsByteSlice(obj *Obj, oEnc uint8) ([]byte, error) {
	switch oEnc {
	case ObjEncodingInt:
		intVal, ok := obj.Value.(int64)
		if !ok {
			return nil, errors.New("expected integer value but got another type")
		}

		return []byte(strconv.FormatInt(intVal, 10)), nil
	case ObjEncodingEmbStr, ObjEncodingRaw:
		strVal, ok := obj.Value.(string)
		if !ok {
			return nil, errors.New("expected string value but got another type")
		}

		return []byte(strVal), nil
	default:
		return nil, fmt.Errorf("unsupported encoding type: %d", oEnc)
	}
}

func getByteArrayValueAsByteSlice(obj *Obj) ([]byte, error) {
	byteArray, ok := obj.Value.(*ByteArray)
	if !ok {
		return nil, errors.New("expected byte array value but got another type")
	}

	return byteArray.data, nil
}

func parseOptionalParams(args []string, byteLen int) (start, end int, rangeType string, endRangeProvided bool, err error) {
	start, end, rangeType = 0, byteLen-1, bit.BYTE
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
		if rangeType != bit.BYTE && rangeType != bit.BIT {
			return 0, 0, "", false, errors.New("syntax error")
		}
	}
	return start, end, rangeType, endRangeProvided, err
}

func getBitPos(byteSlice []byte, bitToFind byte, start, end int, rangeType string, endRangeProvided bool) int {
	byteLen := len(byteSlice)
	bitLen := len(byteSlice) * 8
	if rangeType == bit.BIT {
		// For BIT range, if start is beyond the total bits, it's invalid
		if start > bitLen {
			return -1
		}
	}

	// Adjust start and end for both BYTE and BIT ranges
	// This handles negative indices and ensures we're within bounds
	start, end = adjustBitPosSearchRange(start, end, byteLen)

	// If start is beyond end or byteLen, we can't find anything
	if start > end || start >= byteLen {
		return -1
	}

	var result int
	if rangeType == bit.BYTE {
		result = getBitPosWithByteRange(byteSlice, bitToFind, start, end)
	} else {
		// Convert byte range to bit range
		// We multiply by 8 because each byte has 8 bits
		start *= 8
		// The +7 ensures we include all bits in the last byte
		// min() ensures we don't go beyond the actual bit length
		end = min(end*8+7, bitLen-1)
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

//nolint: gocritic
func adjustBitPosSearchRange(start, end, byteLen int) (int, int) {
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

func getBitPosWithByteRange(byteSlice []byte, bitToFind byte, start, end int) int {
	for i := start; i <= end; i++ {
		for j := 0; j < 8; j++ {
			// Check each bit in the byte from left to right
			// We use 7-j because bit 7 is the leftmost (most significant) bit
			if ((byteSlice[i] >> (7 - j)) & 1) == bitToFind {
				// Return the bit position (i*8 gives us the byte offset in bits)
				return i*8 + j
			}
		}
	}

	// Bit not found in the range
	return -1
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
