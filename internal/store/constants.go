// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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
	HGetAll			 string = "HGETALL"
)
