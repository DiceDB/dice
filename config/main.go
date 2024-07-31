package config

var (
	Host string = "0.0.0.0"
	Port int    = 7379
)

var KeysLimit int = 100

// Will evict EvictionRatio of keys whenever eviction runs
var (
	EvictionRatio    float64 = 0.40
	EvictionStrategy string  = "allkeys-lru"
	AOFFile          string  = "./dice-master.aof"
)

// Network
var (
	IOBufferLength    int = 512
	IOBufferLengthMAX int = 50 * 1024
)
