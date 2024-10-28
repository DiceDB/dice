package id

import (
	"sync"
	"sync/atomic"
	"testing"
)

func BenchmarkNextID(b *testing.B) {
	var workerCounter uint32

	// Run the benchmark in parallel for each workerID to simulate high-concurrency environment
	b.RunParallel(func(pb *testing.PB) {
		workerID := atomic.AddUint32(&workerCounter, 1) - 1

		for pb.Next() {
			NextID(workerID)
		}
	})
}

func TestNextID(t *testing.T) {
	const (
		numWorkers   = 16   // Number of worker threads
		idsPerWorker = 1000 // Number of IDs each worker should generate
	)

	uniqueIDs := make(map[uint32]bool)
	var mu sync.Mutex

	var wg sync.WaitGroup
	for workerID := uint32(0); workerID < numWorkers; workerID++ {
		wg.Add(1)
		go func(wID uint32) {
			defer wg.Done()
			for i := 0; i < idsPerWorker; i++ {
				id := NextID(wID)

				// Lock used only for updating the test map (not in ID generation)
				mu.Lock()
				if uniqueIDs[id] {
					t.Errorf("Duplicate ID found: %d", id)
				}
				uniqueIDs[id] = true
				mu.Unlock()
			}
		}(workerID)
	}

	wg.Wait()
	expectedCount := numWorkers * idsPerWorker
	if len(uniqueIDs) != expectedCount {
		t.Errorf("Expected %d unique IDs, got %d", expectedCount, len(uniqueIDs))
	}
}
