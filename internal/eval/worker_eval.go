package eval

// These evaluation functions are exposed to the worker, without
// making any contact with shards allowing them to process
// commands and return appropriate responses.

import (
	"github.com/dicedb/dice/internal/clientio"
	diceerrors "github.com/dicedb/dice/internal/errors"
)

// RespPING evaluates the PING command and returns the appropriate response.
// If no arguments are provided, it responds with "PONG" (standard behavior).
// If an argument is provided, it returns the argument as the response.
// If more than one argument is provided, it returns an arity error.
func RespPING(args []string) []byte {
	// Check for incorrect number of arguments (arity error).
	if len(args) >= 2 {
		return diceerrors.NewErrArity("PING") // Return an error if more than one argument is provided.
	}

	// If no arguments are provided, return the standard "PONG" response.
	if len(args) == 0 {
		return clientio.Encode("PONG", true) // Encode "PONG" with an indication that this is a simple string response.
	}
	// If one argument is provided, return the argument as the response.
	return clientio.Encode(args[0], false) // Encode the argument and return it as the response.
}
