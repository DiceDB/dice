package core

import (
	"sync"
)

var mu sync.Mutex
var cycle uint32 = 0
var counter uint32 = 0
var turn uint32 = 0

var totalBits int = 32
var turnBits int = 8
var counterBits int = totalBits - turnBits

func ExpandID(id uint32) uint64 {
	return uint64(cycle)<<counterBits | uint64(counter)
}

func NextID() uint32 {
	mu.Lock()
	defer mu.Unlock()

	counter = (counter + 1) & ((1 << counterBits) - 1)
	cycle += (1 - min(counter, 1))
	turn += (1 - min(counter, 1))
	turn &= ((1 << turnBits) - 1)
	return uint32(turn)<<counterBits | uint32(counter)
}
