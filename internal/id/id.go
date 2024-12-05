package id

import (
	"sync/atomic"
)

const (
	totalBits   uint32 = 32                                // Total bits for the ID
	turnBits    uint32 = 4                                 // Bits for turn (allows up to 16 turns before wrap-around
	workerBits  uint32 = 4                                 // Bits allocated for worker ID (supports up to 16 workers, 2^4)
	counterBits uint32 = totalBits - workerBits - turnBits // Remaining bits for counter (24 bits)
)

var (
	counter  []uint32        // Per-worker counter
	turn     []uint32        // Per-worker turn counter
	cycleMap [][]uint32      // Per-worker cycle map
	maxTurns = 1 << turnBits // Max turns before wrap-around
)

func init() {
	numWorkers := 1 << workerBits
	counter = make([]uint32, numWorkers)
	turn = make([]uint32, numWorkers)
	cycleMap = make([][]uint32, numWorkers)

	for i := range cycleMap {
		cycleMap[i] = make([]uint32, maxTurns)
	}
}

// NextID generates a lock-free, unique ID per worker.
// TODO: Persist the cycleMap to disk for server restart.
func NextID(workerID uint32) uint32 {
	if workerID >= (1 << workerBits) {
		panic("workerID exceeds limit for allocated bits")
	}

	// Increment the counter for this worker and mask it to stay within limit
	counterVal := atomic.AddUint32(&counter[workerID], 1) & ((1 << counterBits) - 1)
	if counterVal == 0 {
		// Counter wrapped, so increment turn
		turnVal := atomic.AddUint32(&turn[workerID], 1) & ((1 << turnBits) - 1)
		cycleMap[workerID][turnVal]++
	}

	// Form the unique ID: workerID (4 bits) + turn (4 bits) + counter (24 bits)
	return (workerID << (turnBits + counterBits)) | (turn[workerID] << counterBits) | counterVal
}

// ExpandID extends the ID with the cycle value from the cycleMap.
func ExpandID(id uint32) uint64 {
	workerID := id >> (turnBits + counterBits)
	turnVal := (id >> counterBits) & ((1 << turnBits) - 1)

	// Add cycle information from cycleMap to make the ID globally unique over time
	cycle := uint64(cycleMap[workerID][turnVal])
	return (cycle << (workerBits + turnBits + counterBits)) | uint64(id)
}
