package config

var Host string = "0.0.0.0"
var Port int = 7379

var KeysLimit int = 100

// Will evict EvictionRatio of keys whenever eviction runs
var EvictionRatio float64 = 0.40

var EvictionStrategy string = "allkeys-lru"
var AOFFile string = "./dice-master.aof"
