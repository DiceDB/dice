package config

import (
	"time"

	"github.com/dicedb/dice/internal/server/utils"
)

var (
	TestHost string = "0.0.0.0"
	TestPort int    = 8739
)

var TestKeysLimit int = 10000 // default buffer size.

// Will evict EvictionRatio of keys whenever eviction runs
var (
	TestEvictionRatio    float64 = 0.40
	TestEvictionStrategy string  = "allkeys-lru"
	TestAOFFile          string  = "./dice-master.aof"
)

// Network
var (
	TestIOBufferLength    int = 512
	TestIOBufferLengthMAX int = 50 * 1024
)

var (
	TestShardCronFrequency           = 1 * time.Second
	TestServerMultiplexerPollTimeout = 100 * time.Millisecond
	TestServerMaxClients             = 50
)

// Users and ACLs
var (
	// if RequirePass is set to an empty string, no authentication is required
	TestRequirePass string = utils.EmptyStr
)
