package eval

// Scatter file is used by shardThread to evaluate
// individual commands on each shard. Scatter functions
// returns response as interface and error using
// a structure for each command, which eventiually passed
// till worker from each shard response

import (
	"fmt"
)

type EvalScatterResponse struct {
	Result interface{} // Result of the Store operation, for now the type is set to []byte, but this can change in the future.
	Error  error
}

// ScatterPING function is used by shardThread to
// evaluate PING command
func ScatterPING(args []string) EvalScatterResponse {
	if len(args) >= 2 {
		return EvalScatterResponse{Result: nil, Error: fmt.Errorf("PING")}
	}

	if len(args) == 0 {
		return EvalScatterResponse{Result: "PONG", Error: nil}
	} else {
		return EvalScatterResponse{Result: args[0], Error: nil}
	}
}
