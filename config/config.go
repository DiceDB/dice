package config

import (
	"time"

	"github.com/dicedb/dice/internal/constants"
)

var (
	Host string = "0.0.0.0"
	Port int    = 7379
)

var KeysLimit int = 10000 // default buffer size.

const (
	SIMPLE_FIRST    = "simple-first"
	ALL_KEYS_RANDOM = "allkeys-random"
	ALL_KEYS_LRU    = "allkeys-lru"
	ALL_KEYS_LFU    = "allkeys-lfu"
)

// Will evict EvictionRatio of keys whenever eviction runs
var (
	EvictionRatio    float64 = 0.40
	EvictionStrategy string  = ALL_KEYS_LFU
	AOFFile          string  = "./dice-master.aof"
)

// Network
var (
	IOBufferLength    int = 512
	IOBufferLengthMAX int = 50 * 1024
)

var (
	ShardCronFrequency           = 1 * time.Second
	ServerMultiplexerPollTimeout = 100 * time.Millisecond
	ServerMaxClients             = 20000
)

// Users and ACLs
var (
	// if RequirePass is set to an empty string, no authentication is required
	RequirePass string = constants.EmptyStr
)
