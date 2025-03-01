// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

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
	CmdHandlerNotFoundErr  = "command handler with ID %s not found"
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

// Package errors provides error definitions and utility functions for handling
// common DiceDB error scenarios within the application. This package centralizes
// error messages to ensure consistency and clarity when interacting with DiceDB
// commands and responses.

// Standard error variables for various DiceDB-related error conditions.
var (
	ErrAuthFailed                 = errors.New("AUTH failed")                                                            // Indicates authentication failure.
	ErrIntegerOutOfRange          = errors.New("value is not an integer or out of range")                                // Represents a value that is either not an integer or is out of allowed range.
	ErrInvalidNumberFormat        = errors.New("value is not an integer or a float")                                     // Signals that a value provided is not in a valid integer or float format.
	ErrValueOutOfRange            = errors.New("value is out of range")                                                  // Indicates that a value is beyond the permissible range.
	ErrOverflow                   = errors.New("increment or decrement would overflow")                                  // Signifies that an increment or decrement operation would exceed the limits.
	ErrKeyNotFound                = errors.New("no such key")                                                            // Indicates that the specified key does not exist.
	ErrWrongTypeOperation         = errors.New("wrongtype operation against a key holding the wrong kind of value")      // Signals an operation attempted on a key with an incompatible type.
	ErrInvalidHyperLogLogKey      = errors.New("WRONGTYPE Key is not a valid HyperLogLog string value")                  // Indicates that a key is not a valid HyperLogLog value.
	ErrCorruptedHyperLogLogObject = errors.New("INVALIDOBJ Corrupted HLL object detected")                               // Signals detection of a corrupted HyperLogLog object.
	ErrInvalidJSONPathType        = errors.New("WRONGTYPE wrong type of path value - expected string but found integer") // Represents an invalid type for a JSON path.
	ErrHashValueNotInteger        = errors.New("hash value is not an integer")                                           // Signifies that a hash value is expected to be an integer.
	ErrInternalServer             = errors.New("internal server error, unable to process command")                       // Represents a generic internal server error.
	ErrAuth                       = errors.New("AUTH <password> called without any password configured for the default user. Are you sure your configuration is correct?")
	ErrAborted                    = errors.New("server received ABORT command")
	ErrEmptyCommand               = errors.New("empty command")
	ErrInvalidIPAddress           = errors.New("invalid IP address")
	ErrInvalidFingerprint         = errors.New("invalid fingerprint")
	ErrKeyDoesNotExist            = errors.New("could not perform this operation on a key that doesn't exist")
	ErrKeyExists                  = errors.New("key exists")
	ErrUnknownObjectType          = errors.New("unknown object type")

	ErrInvalidValue = func(command, param string) error {
		return fmt.Errorf("invalid value for a parameter in '%s' command for %s parameter", strings.ToUpper(command), strings.ToUpper(param))
	}

	ErrInvalidSyntax = func(command string) error {
		return fmt.Errorf("invalid syntax for '%s' command", strings.ToUpper(command))
	}

	// Error generation functions for specific error messages with dynamic parameters.
	ErrWrongArgumentCount = func(command string) error {
		return fmt.Errorf("wrong number of arguments for '%s' command", strings.ToUpper(command))
	}
	ErrInvalidExpireTime = func(command string) error {
		return fmt.Errorf("invalid expire time in '%s' command", strings.ToUpper(command)) // Represents an invalid expiration time for a specific command.
	}

	ErrInvalidElementPeekCount = func(max int) error {
		return fmt.Errorf("number of elements to peek should be a positive number less than %d", max) // Signals an invalid count for elements to peek.
	}

	ErrGeneral = func(err string) error {
		return fmt.Errorf("%s", err) // General error format for various commands.
	}

	ErrFormatted = func(errMsg string, opts ...any) error {
		return ErrGeneral(fmt.Sprintf(errMsg, opts...))
	}
	ErrIOThreadNotFound = func(id string) error {
		return fmt.Errorf("io-thread with ID %s not found", id) // Indicates that an io-thread with the specified ID does not exist.
	}

	ErrJSONPathNotFound = func(path string) error {
		return fmt.Errorf("path '%s' does not exist", path) // Represents an error where the specified JSON path cannot be found.
	}

	ErrUnsupportedEncoding = func(encoding int) error {
		return fmt.Errorf("unsupported encoding: %d", encoding) // Indicates that an unsupported encoding type was provided.
	}

	ErrUnexpectedType = func(expectedType string, actualType interface{}) error {
		return fmt.Errorf("expected %s but got another type: %s", expectedType, actualType) // Signals an unexpected type received when an integer was expected.
	}

	ErrUnexpectedJSONPathType = func(expectedType string, actualType interface{}) error {
		return fmt.Errorf("wrong type of path value - expected %s but found %s", expectedType, actualType) // Signals an unexpected type received when an integer was expected.
	}

	ErrUnknownCmd = func(cmd string) error {
		return fmt.Errorf("ERROR unknown command '%v'", cmd) // Indicates that an unsupported encoding type was provided.
	}
)

type PreProcessError struct {
	Result interface{}
}

func (e *PreProcessError) Error() string {
	return fmt.Sprintf("%v", e.Result)
}

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
