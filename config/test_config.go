package config

import (
	"time"

	"github.com/dicedb/dice/internal/server/utils"
)

var (
	Test_Host string = "0.0.0.0"
	Test_Port int    = 7379
)

var Test_KeysLimit int = 10000 // default buffer size.

// Will evict EvictionRatio of keys whenever eviction runs
var (
	Test_EvictionRatio    float64 = 0.40
	Test_EvictionStrategy string  = "allkeys-lru"
	Test_AOFFile          string  = "./dice-master.aof"
)

// Network
var (
	Test_IOBufferLength    int = 512
	Test_IOBufferLengthMAX int = 50 * 1024
)

var (
	Test_ShardCronFrequency           = 1 * time.Second
	Test_ServerMultiplexerPollTimeout = 100 * time.Millisecond
	Test_ServerMaxClients             = 50
)

// Users and ACLs
var (
	// if RequirePass is set to an empty string, no authentication is required
	Test_RequirePass string = utils.EmptyStr
)
