package worker

import (
	"context"
	"sort"

	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/eval"
)

// Gather file is used by Worker to collect and process responses
// from multiple shards. For commands that are executed across
// several shards (e.g., MultiShard commands), a Gather function
// is responsible for aggregating the results.
//
// Each Gather function takes input in the form of shard responses,
// applies command-specific logic to combine or process these
// individual shard responses, and returns the final response
// expected by the client.
//
// The result is a unified response that reflects the combined
// outcome of operations executed across multiple shards, ensuring
// that the client receives a single, cohesive result.
func composeRename(ctx context.Context, responses ...eval.EvalResponse) interface{} {
	for idx := range responses {
		if responses[idx].Error != nil {
			return responses[idx].Error
		}
	}

	return clientio.OK
}

func composeCopy(ctx context.Context, responses ...eval.EvalResponse) interface{} {
	for idx := range responses {
		if responses[idx].Error != nil {
			return responses[idx].Error
		}
	}

	return clientio.OK
}

func composeMSet(_ context.Context, responses ...eval.EvalResponse) interface{} {
	for idx := range responses {
		if responses[idx].Error != nil {
			return responses[idx].Error
		}
	}

	return clientio.OK
}

func composeMGet(_ context.Context, responses ...eval.EvalResponse) interface{} {
	sort.Slice(responses, func(i, j int) bool {
		return responses[i].SeqID < responses[j].SeqID
	})

	results := make([]interface{}, 0, len(responses))

	for idx := range responses {
		if responses[idx].Error != nil {
			return responses[idx].Error
		}

		results = append(results, responses[idx].Result)
	}

	return results
}
