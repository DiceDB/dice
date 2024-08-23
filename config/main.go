package config

import "github.com/dicedb/dice/internal/constants"

var (
	Host string = "0.0.0.0"
	Port int    = 7379
)

var KeysLimit int = 10000 // default buffer size.

// Will evict EvictionRatio of keys whenever eviction runs
var (
	EvictionRatio    float64 = 0.40
	EvictionStrategy string  = "allkeys-lru"
	AOFFile          string  = "./dice-master.aof"
	TempAOFFile      string  = "./tmp-dice-master.aof"
)

// Network
var (
	IOBufferLength    int = 512
	IOBufferLengthMAX int = 50 * 1024
)

// Users and ACLs
var (
	// If requirepass is set to an empty string, no authentication is required
	RequirePass string = constants.EmptyStr
)
