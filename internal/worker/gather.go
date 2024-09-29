package worker

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
