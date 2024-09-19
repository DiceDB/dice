package eval

// Gather file is used by Worker to collect and process
// responses from each shard in the form of scatterResponse.
// Each Gather function takes input from available shard
// responses implements a command specific logic
// to handle responses from individual shards and returns
// response which client expects for each command

import (
	"fmt"

	diceerrors "github.com/dicedb/dice/internal/errors"
)

// GatherPING gathers responses from each shard thread
// If error is returned by any shard then dice returns error
// If no error is returned by any shard then dice returns
// response to client from first shard response
func GatherPING(responses ...EvalResponse) []byte {
	for idx := range responses {
		if responses[idx].Error != nil {
			return diceerrors.NewErrArity("PING")
		}
	}

	return []byte(fmt.Sprintf("+%s\r\n", responses[0].Result))
}
