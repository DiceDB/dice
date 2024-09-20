package server

// Gather file is used by Worker to collect and process
// responses from each shard in the form of scatterResponse.
// Each Gather function takes input from available shard
// responses implements a command specific logic
// to handle responses from individual shards and returns
// response which client expects for each command
