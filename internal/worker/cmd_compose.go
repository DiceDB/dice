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

// composeRename processes responses from multiple shards for a "Rename" operation.
// It iterates through all shard responses, checking for any errors. If an error is found
// in any shard response, it returns that error immediately. If all responses are successful,
// it returns an "OK" response to indicate that the Rename operation succeeded across all shards.
func composeRename(ctx context.Context, responses ...eval.EvalResponse) interface{} {
	for idx := range responses {
		if responses[idx].Error != nil {
			return responses[idx].Error
		}
	}

	return clientio.OK
}

// composeRename processes responses from multiple shards for a "Copy" operation.
// It iterates through all shard responses, checking for any errors. If an error is found
// in any shard response, it returns that error immediately. If all responses are successful,
// it returns an "OK" response to indicate that the Rename operation succeeded across all shards.

func composeCopy(ctx context.Context, responses ...eval.EvalResponse) interface{} {
	for idx := range responses {
		if responses[idx].Error != nil {
			return responses[idx].Error
		}
	}

	return clientio.OK
}

// composeMSet processes responses from multiple shards for an "MSet" operation
// (Multi-set operation). It loops through the responses to check if any shard returned an error.
// If an error is detected, it immediately returns that error. Otherwise, it returns "OK"
// to indicate that all "MSet" operations across shards were successful.
func composeMSet(_ context.Context, responses ...eval.EvalResponse) interface{} {
	for idx := range responses {
		if responses[idx].Error != nil {
			return responses[idx].Error
		}
	}

	return clientio.OK
}

// composeMGet processes responses from multiple shards for an "MGet" operation
// (Multi-get operation). It first sorts the responses by their SeqID to ensure the results
// are in the correct sequence. It then checks for any errors in the responses; if any error
// is encountered, it returns the error. If no errors are found, the function collects the
// results from all responses and returns them as a slice.
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
