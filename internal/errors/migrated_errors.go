package errors

import (
	"errors"
	"fmt"
	"strings"
)

// Package errors provides error definitions and utility functions for handling
// common DiceDB error scenarios within the application. This package centralizes
// error messages to ensure consistency and clarity when interacting with DiceDB
// commands and responses.

// Standard error variables for various DiceDB-related error conditions.
var (
	ErrAuthFailed                 = errors.New("AUTH failed")                                                            // Indicates authentication failure.
	ErrIntegerOutOfRange          = errors.New("ERR value is not an integer or out of range")                            // Represents a value that is either not an integer or is out of allowed range.
	ErrInvalidNumberFormat        = errors.New("ERR value is not an integer or a float")                                 // Signals that a value provided is not in a valid integer or float format.
	ErrValueOutOfRange            = errors.New("ERR value is out of range")                                              // Indicates that a value is beyond the permissible range.
	ErrOverflow                   = errors.New("ERR increment or decrement would overflow")                              // Signifies that an increment or decrement operation would exceed the limits.
	ErrSyntax                     = errors.New("ERR syntax error")                                                       // Represents a syntax error in a DiceDB command.
	ErrKeyNotFound                = errors.New("ERR no such key")                                                        // Indicates that the specified key does not exist.
	ErrWrongTypeOperation         = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")      // Signals an operation attempted on a key with an incompatible type.
	ErrInvalidHyperLogLogKey      = errors.New("WRONGTYPE Key is not a valid HyperLogLog string value")                  // Indicates that a key is not a valid HyperLogLog value.
	ErrCorruptedHyperLogLogObject = errors.New("INVALIDOBJ Corrupted HLL object detected")                               // Signals detection of a corrupted HyperLogLog object.
	ErrInvalidJSONPathType        = errors.New("WRONGTYPE wrong type of path value - expected string but found integer") // Represents an invalid type for a JSON path.
	ErrInvalidExpireTimeValue     = errors.New("ERR invalid expire time")                                                // Indicates that the provided expiration time is invalid.
	ErrHashValueNotInteger        = errors.New("ERR hash value is not an integer")                                       // Signifies that a hash value is expected to be an integer.
	ErrInternalServer             = errors.New("ERR Internal server error, unable to process command")                   // Represents a generic internal server error.
	ErrAuth                       = errors.New("AUTH <password> called without any password configured for the default user. Are you sure your configuration is correct?")
	ErrAborted                    = errors.New("server received ABORT command")
	ErrEmptyCommand               = errors.New("empty command")
	ErrInvalidIPAddress           = errors.New("invalid IP address")
	ErrInvalidFingerprint         = errors.New("invalid fingerprint")
	ErrKeyDoesNotExist            = errors.New("ERR could not perform this operation on a key that doesn't exist")

	// Error generation functions for specific error messages with dynamic parameters.
	ErrWrongArgumentCount = func(command string) error {
		return fmt.Errorf("ERR wrong number of arguments for '%s' command", strings.ToLower(command)) // Indicates an incorrect number of arguments for a given command.
	}
	ErrInvalidExpireTime = func(command string) error {
		return fmt.Errorf("ERR invalid expire time in '%s' command", strings.ToLower(command)) // Represents an invalid expiration time for a specific command.
	}

	ErrInvalidElementPeekCount = func(max int) error {
		return fmt.Errorf("ERR number of elements to peek should be a positive number less than %d", max) // Signals an invalid count for elements to peek.
	}

	ErrGeneral = func(err string) error {
		return fmt.Errorf("ERR %s", err) // General error format for various commands.
	}

	ErrFormatted = func(errMsg string, opts ...any) error {
		return ErrGeneral(fmt.Sprintf(errMsg, opts...))
	}
	ErrIOThreadNotFound = func(id string) error {
		return fmt.Errorf("ERR io-thread with ID %s not found", id) // Indicates that an io-thread with the specified ID does not exist.
	}

	ErrJSONPathNotFound = func(path string) error {
		return fmt.Errorf("ERR Path '%s' does not exist", path) // Represents an error where the specified JSON path cannot be found.
	}

	ErrUnsupportedEncoding = func(encoding int) error {
		return fmt.Errorf("ERR unsupported encoding: %d", encoding) // Indicates that an unsupported encoding type was provided.
	}

	ErrUnexpectedType = func(expectedType string, actualType interface{}) error {
		return fmt.Errorf("ERR expected %s but got another type: %s", expectedType, actualType) // Signals an unexpected type received when an integer was expected.
	}

	ErrUnexpectedJSONPathType = func(expectedType string, actualType interface{}) error {
		return fmt.Errorf("ERR wrong type of path value - expected %s but found %s", expectedType, actualType) // Signals an unexpected type received when an integer was expected.
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
