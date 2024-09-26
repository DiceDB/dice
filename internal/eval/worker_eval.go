package eval

// This file contains functions required by worker nodes to
// evaluate specific Redis-like commands (e.g., INFO, PING).
// These evaluation functions are exposed to the worker,
// allowing them to process commands and return appropriate responses.

import (
	"bytes"
	"fmt"

	"github.com/dicedb/dice/internal/clientio"
	diceerrors "github.com/dicedb/dice/internal/errors"
	dstore "github.com/dicedb/dice/internal/store"
)

// RespINFO generates a response for the INFO command.
// The INFO command provides the total number of keys per database in the keyspace.
// It creates a buffer that stores the keyspace information and formats it according to the Redis protocol.
// The function returns the encoded buffer as the response.
func RespINFO(args []string) []byte {
	var info []byte                      // Initialize a byte slice to hold the keyspace info.
	buf := bytes.NewBuffer(info)         // Create a buffer for efficiently writing the response.
	buf.WriteString("# Keyspace\r\n")    // Write the header "# Keyspace" to indicate the section for keyspace stats.
	for i := range dstore.KeyspaceStat { // Iterate through the KeyspaceStat map, which holds the number of keys per database.
		fmt.Fprintf(buf, "db%d:keys=%d,expires=0,avg_ttl=0\r\n", i, dstore.KeyspaceStat[i]["keys"]) // Format the keyspace info for each database.
	}
	return clientio.Encode(buf.String(), false) // Encode the result into a Redis protocol-compliant format and return the encoded bytes.
}

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
