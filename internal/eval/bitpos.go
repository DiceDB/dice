// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package eval

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/store"
)

type RespType int

var RespNIL = []byte("$-1\r\n")
var RespOK = []byte("+OK\r\n")
var RespQueued = []byte("+QUEUED\r\n")
var RespZero = []byte(":0\r\n")
var RespOne = []byte(":1\r\n")
var RespMinusOne = []byte(":-1\r\n")
var RespMinusTwo = []byte(":-2\r\n")
var RespEmptyArray = []byte("*0\r\n")

const (
	NIL                RespType = iota // Represents an empty or null response.
	OK                                 // Represents a successful "OK" response.
	CommandQueued                      // Represents that a command has been queued for execution.
	IntegerZero                        // Represents the integer value zero in RESP format.
	IntegerOne                         // Represents the integer value one in RESP format.
	IntegerNegativeOne                 // Represents the integer value negative one in RESP format.
	IntegerNegativeTwo                 // Represents the integer value negative two in RESP format.
	EmptyArray                         // Represents an empty array in RESP format.
)

func Encode(value interface{}, isSimple bool) []byte {
	if isSimple {
		return []byte(fmt.Sprintf(":%v\r\n", value))
	}
	return []byte(fmt.Sprintf(":%v\r\n", value))
}

func evalBITPOS(args []string, st *store.Store) *EvalResponse {
	if len(args) < 2 || len(args) > 5 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("BITPOS"),
		}
	}

	key := args[0]
	obj := st.Get(key)

	bitToFind, err := parseBitToFind(args[1])
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(err.Error()),
		}
	}

	if obj == nil {
		if bitToFind == 0 {
			return &EvalResponse{
				Result: IntegerZero,
				Error:  nil,
			}
		}

		return &EvalResponse{
			Result: IntegerNegativeOne,
			Error:  nil,
		}
	}

	byteSlice, err := getValueAsByteSlice(obj)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  err,
		}
	}

	start, end, rangeType, endRangeProvided, err := parseOptionalParams(args[2:], len(byteSlice))
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  err,
		}
	}

	result := getBitPos(byteSlice, bitToFind, start, end, rangeType, endRangeProvided)

	return &EvalResponse{
		Result: result,
		Error:  nil,
	}
}

func parseBitToFind(arg string) (byte, error) {
	bitToFindInt, err := strconv.Atoi(arg)
	if err != nil {
		return 0, diceerrors.ErrIntegerOutOfRange
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
			return 0, 0, "", false, diceerrors.ErrIntegerOutOfRange
		}
	}

	if len(args) > 1 {
		end, err = strconv.Atoi(args[1])
		if err != nil {
			return 0, 0, "", false, diceerrors.ErrIntegerOutOfRange
		}
		endRangeProvided = true
	}

	if len(args) > 2 {
		rangeType = strings.ToUpper(args[2])
		if rangeType != BYTE && rangeType != BIT {
			return 0, 0, "", false, diceerrors.ErrInvalidSyntax("BITPOS")
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
