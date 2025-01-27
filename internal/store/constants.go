// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package store

const (
	Set              string = "SET"
	Del              string = "DEL"
	Get              string = "GET"
	Rename           string = "RENAME"
	ZAdd             string = "ZADD"
	ZRange           string = "ZRANGE"
	Replace          string = "REPLACE"
	Smembers         string = "SMEMBERS"
	JSONGet          string = "JSON.GET"
	PFADD            string = "PFADD"
	PFCOUNT          string = "PFCOUNT"
	PFMERGE          string = "PFMERGE"
	KEYSPERSHARD     string = "KEYSPERSHARD"
	Evict            string = "EVICT"
	SingleShardSize  string = "SINGLEDBSIZE"
	SingleShardTouch string = "SINGLETOUCH"
	SingleShardKeys  string = "SINGLEKEYS"
	FlushDB          string = "FLUSHDB"
)
