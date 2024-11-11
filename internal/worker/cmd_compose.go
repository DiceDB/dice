package worker

import (
	"sort"

	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/ops"
)

// This file contains functions used by the Worker to handle and process responses
// from multiple shards during distributed operations. For commands that are executed
// across several shards, such as MultiShard commands, dedicated functions are responsible
// for aggregating and managing the results.
//
// Each function takes a variable number of shard responses as input, applies command-specific
// logic to evaluate or combine these responses, and returns the final outcome to the client.
//
// The goal is to provide a unified response that accurately reflects the result of operations
// executed across multiple shards, ensuring the client receives a single, coherent result.
//

// composeRename processes responses from multiple shards for a "Rename" operation.
// It iterates through all shard responses, checking for any errors. If an error is found
// in any shard response, it returns that error immediately. If all responses are successful,
// it returns an "OK" response to indicate that the Rename operation succeeded across all shards.
func composeRename(responses ...ops.StoreResponse) interface{} {
	for idx := range responses {
		if responses[idx].EvalResponse.Error != nil {
			return responses[idx].EvalResponse.Error
		}
	}

	return clientio.OK
}

// composeCopy processes responses from multiple shards for a "Copy" operation.
// It iterates through all shard responses, checking for any errors. If an error is found
// in any shard response, it returns that error immediately. If all responses are successful,
// it returns an "OK" response to indicate that the Copy operation succeeded across all shards.
func composeCopy(responses ...ops.StoreResponse) interface{} {
	for idx := range responses {
		if responses[idx].EvalResponse.Error != nil {
			return responses[idx].EvalResponse.Error
		}
	}

	return clientio.OK
}

// composeMSet processes responses from multiple shards for an "MSet" operation
// (Multi-set operation). It loops through the responses to check if any shard returned an error.
// If an error is detected, it immediately returns that error. Otherwise, it returns "OK"
// to indicate that all "MSet" operations across shards were successful.
func composeMSet(responses ...ops.StoreResponse) interface{} {
	for idx := range responses {
		if responses[idx].EvalResponse.Error != nil {
			return responses[idx].EvalResponse.Error
		}
	}

	return clientio.OK
}

// composeMGet processes responses from multiple shards for an "MGet" operation
// (Multi-get operation). It first sorts the responses by their SeqID to ensure the results
// are in the correct sequence. It then checks for any errors in the responses; if any error
// is encountered, it returns the error. If no errors are found, the function collects the
// results from all responses and returns them as a slice.
func composeMGet(responses ...ops.StoreResponse) interface{} {
	sort.Slice(responses, func(i, j int) bool {
		return responses[i].SeqID < responses[j].SeqID
	})

	results := make([]interface{}, 0, len(responses))

	for idx := range responses {
		if responses[idx].EvalResponse.Error != nil {
			return responses[idx].EvalResponse.Error
		}

		results = append(results, responses[idx].EvalResponse.Result)
	}

	return results
}
