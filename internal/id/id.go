// This file is part of DiceDB.
// Copyright (C) 2025  DiceDB (dicedb.io).
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

package id

import (
	"sync"
)

var mu sync.Mutex
var turn, cycle, counter uint32 = 0, 0, 0

var totalBits uint32 = 32
var turnBits uint32 = 4
var counterBits = totalBits - turnBits

var cycleMap []uint32

func init() {
	cycleMap = make([]uint32, 1<<turnBits)
}

func ExpandID(id uint32) uint64 {
	_id := uint64(id)
	_id |= uint64(cycleMap[id>>counterBits]) << counterBits
	return _id
}

// NextID returns a new unique ID
// TODO: Persisting the cycle on disk and reloading it when we start the server
func NextID() uint32 {
	mu.Lock()
	defer mu.Unlock()
	counter = (counter + 1) & ((1 << counterBits) - 1)
	if counter == 0 {
		cycle++
		turn = (turn + 1) & ((1 << turnBits) - 1)
		cycleMap[turn] = cycle
	}
	return (turn << counterBits) | counter
}
