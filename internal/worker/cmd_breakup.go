package worker

// Breakup file is used by Worker to split commands that need to be executed
// across multiple shards. For commands that operate on multiple keys or
// require distribution across shards (e.g., MultiShard commands), a Breakup
// function is invoked to break the original command into multiple smaller
// commands, each targeted at a specific shard.
//
// Each Breakup function takes the input command, identifies the relevant keys
// and their corresponding shards, and generates a set of commands that are
// individually sent to the respective shards. This ensures that commands can
// be executed in parallel across shards, allowing for horizontal scaling
// and distribution of data processing.
//
// The result is a list of commands, one for each shard, which are then
// scattered to the shard threads for execution.
