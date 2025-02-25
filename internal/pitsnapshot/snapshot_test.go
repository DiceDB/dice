package pitsnapshot

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testBufferSize = 1000000
)

func addDataToDummyStore(store *DummyStore, count int) {
	for i := 0; i < count; i++ {
		store.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}
	//log.Println("Total size of store", len(store.store))
}

func fetchOrWriteInStore(ctx context.Context, store *DummyStore, count int, rangeOfKeys int, duration time.Duration, shouldWrite bool) {
	randomKeys, _ := pickRandomNumbers(0, count-1, rangeOfKeys)
	readsCount := 0
	writesCount := 0

	for {
		select {
		case <-time.After(duration):
			log.Println("Duration over", "Reads count:", readsCount, "Writes count:", writesCount)
			return
		case <-ctx.Done():
			log.Println("Context done", "Reads count:", readsCount, "Writes count:", writesCount)
			return
		default:
			for _, key := range randomKeys {
				if shouldWrite {
					writesCount += 1
					store.Set(fmt.Sprintf("key%d", key), fmt.Sprintf("value%d", key))
				}
				readsCount += 1
				store.Get(fmt.Sprintf("key%d", key))
			}
		}
	}
}

func pickRandomNumbers(min, max, x int) ([]int, error) {
	if min > max {
		return nil, fmt.Errorf("min should be less than or equal to max")
	}
	if x < 0 {
		return nil, fmt.Errorf("number of random numbers to pick should be non-negative")
	}
	if x > (max - min + 1) {
		return nil, fmt.Errorf("number of random numbers to pick is greater than the range size")
	}

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Create a map to ensure unique random numbers
	randomNumbers := make(map[int]struct{})
	for len(randomNumbers) < x {
		num := rand.Intn(max-min+1) + min
		randomNumbers[num] = struct{}{}
	}

	// Convert the map keys to a slice
	result := make([]int, 0, x)
	for num := range randomNumbers {
		result = append(result, num)
	}

	return result, nil
}

// Total key size - 1000
// Range of keys to access - 100
// Duration - 10ms
// Allow writes - No
func TestSnapshotWithoutChangesWithNoRangeAccess(t *testing.T) {
	storeSize := testBufferSize
	// Create a new dummy store
	dummyStore := NewDummyStore()
	addDataToDummyStore(dummyStore, storeSize)

	snapshot, err := NewPointInTimeSnapshot(context.Background(), dummyStore)
	snapshot.Run()

	assert.Nil(t, err)
	<-snapshot.ctx.Done()
	time.Sleep(10 * time.Millisecond)
}

// Total key size - 1000
// Range of keys to access - 100
// Duration - 10ms
// Allow writes - No
func TestSnapshotWithoutChangesWithLowRangeAccess(t *testing.T) {
	storeSize := testBufferSize
	// Create a new dummy store
	dummyStore := NewDummyStore()
	addDataToDummyStore(dummyStore, storeSize)

	snapshot, err := NewPointInTimeSnapshot(context.Background(), dummyStore)
	go fetchOrWriteInStore(snapshot.ctx, dummyStore, storeSize/10000, storeSize/100000, 10*time.Millisecond, false)
	snapshot.Run()

	assert.Nil(t, err)
	<-snapshot.ctx.Done()
	time.Sleep(10 * time.Millisecond)
}

// Total key size - 1000
// Range of keys to access - 100
// Duration - 10ms
// Allow writes - No
func TestSnapshotWithChangesWithLowRangeAccess(t *testing.T) {
	storeSize := testBufferSize
	// Create a new dummy store
	dummyStore := NewDummyStore()
	addDataToDummyStore(dummyStore, storeSize)

	snapshot, err := NewPointInTimeSnapshot(context.Background(), dummyStore)
	go fetchOrWriteInStore(snapshot.ctx, dummyStore, storeSize/10000, storeSize/100000, 10*time.Millisecond, true)
	snapshot.Run()

	assert.Nil(t, err)
	<-snapshot.ctx.Done()
	time.Sleep(10 * time.Millisecond)
}

// Test write speed without any snapshots
// Total key size - 1000
// Range of keys to access - 100
// Duration - 10ms
// Allow writes - No
func TestNoSnapshotWithChangesWithLowRangeAccess(t *testing.T) {
	storeSize := testBufferSize
	// Create a new dummy store
	dummyStore := NewDummyStore()
	addDataToDummyStore(dummyStore, storeSize)

	ctx, cancel := context.WithCancel(context.Background())
	go fetchOrWriteInStore(ctx, dummyStore, storeSize/10000, storeSize/100000, 10*time.Millisecond, true)

	time.Sleep(750 * time.Millisecond)
	cancel()
	<-ctx.Done()
	time.Sleep(10 * time.Millisecond)
}

func TestNoSnapshotWithChangesWithWideRangeAccess(t *testing.T) {
	storeSize := testBufferSize
	// Create a new dummy store
	dummyStore := NewDummyStore()
	addDataToDummyStore(dummyStore, storeSize)

	snapshot, err := NewPointInTimeSnapshot(context.Background(), dummyStore)
	go fetchOrWriteInStore(snapshot.ctx, dummyStore, storeSize/10, storeSize/100, 10*time.Millisecond, true)
	snapshot.Run()

	assert.Nil(t, err)

	time.Sleep(750 * time.Millisecond)
	<-snapshot.ctx.Done()
	time.Sleep(10 * time.Millisecond)
}

func TestNoSnapshotWithChangesWithAllRangeAccess(t *testing.T) {
	storeSize := testBufferSize
	// Create a new dummy store
	dummyStore := NewDummyStore()
	addDataToDummyStore(dummyStore, storeSize)

	snapshot, err := NewPointInTimeSnapshot(context.Background(), dummyStore)
	go fetchOrWriteInStore(snapshot.ctx, dummyStore, storeSize, storeSize, 10*time.Millisecond, true)
	snapshot.Run()

	assert.Nil(t, err)

	time.Sleep(750 * time.Millisecond)
	<-snapshot.ctx.Done()
	time.Sleep(10 * time.Millisecond)
}
