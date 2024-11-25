package errors

import (
	"errors"
	"fmt"
	"strings"
)

const (
	ArityErr               = "wrong number of arguments for '%s' command"
	SyntaxErr              = "syntax error"
	ExpiryErr              = "invalid expire time in '%s' command"
	IntOrOutOfRangeErr     = "value is not an integer or out of range"
	IntOrFloatErr          = "value is not an integer or a float"
	ValOutOfRangeErr       = "value is out of range"
	IncrDecrOverflowErr    = "increment or decrement would overflow"
	NoKeyErr               = "no such key"
	ErrDefault             = "-ERR %s"
	WrongTypeErr           = "-WRONGTYPE Operation against a key holding the wrong kind of value"
	WrongTypeHllErr        = "-WRONGTYPE Key is not a valid HyperLogLog string value."
	InvalidHllErr          = "-INVALIDOBJ Corrupted HLL object detected"
	IOThreadNotFoundErr    = "io-thread with ID %s not found"
	JSONPathValueTypeErr   = "-WRONGTYPE wrong type of path value - expected string but found %s"
	HashValueNotIntegerErr = "hash value is not an integer"
	InternalServerError    = "-ERR: Internal server error, unable to process command"
	InvalidFloatErr        = "-ERR value is not a valid float"
	InvalidIntErr          = "-ERR value is not a valid integer"
	InvalidBitfieldType    = "-ERR Invalid bitfield type. Use something like i16 u8. Note that u64 is not supported but i64 is."
	BitfieldOffsetErr      = "-ERR bit offset is not an integer or out of range"
	OverflowTypeErr        = "-ERR Invalid OVERFLOW type specified"
	WrongKeyTypeErr        = "-ERR Existing key has wrong Dice type"
	NoKeyExistsErr         = "-ERR Could not perform this operation on a key that doesn't exist"
)

type DiceError struct {
	message error
}

func newDiceErr(message string) *DiceError {
	return &DiceError{
		message: errors.New(message),
	}
}

func (d *DiceError) toEncodedMessage() []byte {
	return []byte(fmt.Sprintf("%s\r\n", d.message.Error()))
}

func NewErr(message string) error {
	return newDiceErr(message).message
}

// NewErrWithMessage If the error code is already passed in the string,
// the error code provided is used, otherwise the string "-ERR " for the generic
// error code is automatically added. Note that 's' must NOT end with \r\n.
func NewErrWithMessage(errMsg string) []byte {
	// If the string already starts with "-..." then the error code
	// is provided by the caller. Otherwise, we use "-ERR".
	if errMsg == "" || errMsg[0] != '-' {
		errMsg = fmt.Sprintf(ErrDefault, errMsg)
	}

	return newDiceErr(errMsg).toEncodedMessage()
}

func NewErrWithFormattedMessage(errMsgFmt string, args ...interface{}) []byte {
	if len(args) > 0 {
		errMsgFmt = fmt.Sprintf(errMsgFmt, args...)
	}

	return NewErrWithMessage(errMsgFmt)
}

func NewErrArity(cmdName string) []byte {
	return NewErrWithFormattedMessage(ArityErr, strings.ToLower(cmdName))
}

func NewErrExpireTime(cmdName string) []byte {
	return NewErrWithFormattedMessage(ExpiryErr, strings.ToLower(cmdName))
}
