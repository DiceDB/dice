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
