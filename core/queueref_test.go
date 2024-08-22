package core_test

import (
	"fmt"
	"testing"

	"time"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/server/utils"
	"github.com/dicedb/dice/testutils"
	"gotest.tools/v3/assert"
)

func TestQueueRef(t *testing.T) {
	qr, err := core.NewQueueRef()
	if err != nil {
		t.Fatal(err)
	}
	store := core.NewStore()

	if _, err := qr.Remove(store); err != core.ErrQueueEmpty {
		t.Error("removing from an empty queueref should return an empty queue error")
	}

	if qr.Insert("does not exist", store) != false {
		t.Error("inserting the reference of the key that does not exist should not work but it did")
	}

	store.Put("key that exists", store.NewObj(10, -1, core.ObjTypeString, core.ObjEncodingInt))
	if qr.Insert("key that exists", store) != true {
		t.Error("inserting the reference of the key that exists should work but it did not")
	}

	if qe, _ := qr.Remove(store); qe.Obj.Value != 10 {
		t.Error("removing from the queueref should return the obj")
	}

	if _, err := qr.Remove(store); err != core.ErrQueueEmpty {
		t.Error("removing again from an empty queueref should return an empty queue error")
	}

	store.Put("key1", store.NewObj(10, -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key1", store)
	store.Put("key2", store.NewObj(20, -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key2", store)
	store.Put("key3", store.NewObj(30, -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key3", store)
	store.Put("key4", store.NewObj(40, -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key4", store)
	store.Put("key5", store.NewObj(50, -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key5", store)
	store.Put("key6", store.NewObj(60, -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key6", store)
	store.Put("key7", store.NewObj(60, -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key7", store)

	observedVals := make([]int64, 0, 6)
	for _, qe := range qr.Iterate(6, store) {
		observedVals = append(observedVals, int64(qe.Obj.Value.(int)))
	}

	expectedVals := []int64{10, 20, 30, 40, 50, 60}
	if !testutils.EqualInt64Slice(observedVals, expectedVals) {
		t.Error("iterating through the queueref should return the objs in the order they were added. Expected: ", expectedVals, " Got: ", observedVals)
	}
}

// Test for removing from queue with single non-expired keys
func TestRemoveSingleNonExpiredKeys(t *testing.T) {
	qr, err := core.NewQueueRef()
	if err != nil {
		t.Fatal(err)
	}
	store := core.NewStore()
	val := 10
	store.Put("key1", store.NewObj(val, -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key1", store)
	qe, err := qr.Remove(store)
	assert.Check(t, err == nil || qe.Obj.Value == val, fmt.Sprintf("removal from queue with single non expired key failed, Expected : %d, Got : %d\n", val, qe.Obj.Value))
}

// Test for removing from queue with multiple non-expired keys
func TestRemoveMultipleNonExpiredKeys(t *testing.T) {
	qr, err := core.NewQueueRef()
	if err != nil {
		t.Fatal(err)
	}
	store := core.NewStore()
	val := [3]int{10, 20, 30}
	store.Put("key1", store.NewObj(val[0], -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key1", store)
	store.Put("key2", store.NewObj(val[1], -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key2", store)
	store.Put("key3", store.NewObj(val[2], -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key3", store)
	for i := 0; i < len(val); i++ {
		qe, err := qr.Remove(store)
		assert.Check(t, err == nil && qe.Obj.Value == val[i], fmt.Sprintf("removal from queue with multiple non expired keys failed, Expected : %d , Got : %d\n", val[i], qe.Obj.Value))
	}
}

// Test for removing from queue with expired keys before non-expired
func TestRemoveExpiredBeforeNonExpire(t *testing.T) {
	store := core.NewStore()
	mockTime := &utils.MockClock{CurrTime: time.Now()}
	utils.CurrentTime = mockTime
	qr, err := core.NewQueueRef()
	if err != nil {
		t.Fatal(err)
	}
	val := [3]int{10, 20}
	store.Put("key1", store.NewObj(val[0], 1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key1", store)
	store.Put("key2", store.NewObj(val[1], -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key2", store)
	mockTime.SetTime(time.Now().Add(2 * time.Millisecond))
	qe, err := qr.Remove(store)
	assert.Check(t, err == nil || qe.Obj.Value == val[1], fmt.Sprintf("test for removing expired key before non-expired failed , Expected : %d, Got : %d\n", val[1], qe.Obj.Value))
}

// Test for removing from queue with multiple expired keys before non-expired
func TestRemoveMultipleExpiredBeforeNonExpire(t *testing.T) {
	store := core.NewStore()
	mockTime := &utils.MockClock{CurrTime: time.Now()}
	utils.CurrentTime = mockTime
	qr, err := core.NewQueueRef()
	if err != nil {
		t.Fatal(err)
	}
	val := [3]int{10, 20, 30}
	store.Put("key1", store.NewObj(val[0], 1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key1", store)
	store.Put("key2", store.NewObj(val[1], 1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key2", store)
	store.Put("key3", store.NewObj(val[2], -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key3", store)
	mockTime.SetTime(time.Now().Add(2 * time.Millisecond))
	fmt.Printf("Queue size : %d\n", qr.Length(store))
	qe, err := qr.Remove(store)
	assert.Check(t, err == nil || qe.Obj.Value == val[2], fmt.Sprintf("test for removing mulitple expired key before non-expired failed , Expected : %d, Got %d\n", val[2], qe.Obj.Value))
}

// Test for removing from queue with all expired keys
func TestRemoveAllExpired(t *testing.T) {
	store := core.NewStore()
	mockTime := &utils.MockClock{CurrTime: time.Now()}
	utils.CurrentTime = mockTime
	qr, err := core.NewQueueRef()
	if err != nil {
		t.Fatal(err)
	}
	val := [3]int{10, 20, 30}
	store.Put("key1", store.NewObj(val[0], 1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key1", store)
	store.Put("key2", store.NewObj(val[1], 1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key2", store)
	store.Put("key3", store.NewObj(val[2], 1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key3", store)
	mockTime.SetTime(time.Now().Add(2 * time.Millisecond))
	// remove from empty queue
	_, err = qr.Remove(store)
	assert.Equal(t, err, core.ErrQueueEmpty, fmt.Sprintf("test for removing from empty queue failed Expected : %s, Got : %s\n", core.ErrQueueEmpty, err))
}

func TestQueueRefMaxConstraints(t *testing.T) {
	config.KeysLimit = 20000000
	core.WatchChannel = make(chan core.WatchEvent, config.KeysLimit)

	qr, err := core.NewQueueRef()
	if err != nil {
		t.Errorf("error creating QueueRef: %v", err)
	}
	store := core.NewStore()
	for i := 0; i < core.MaxQueueSize; i++ {
		key := fmt.Sprintf("key%d", i)
		store.Put(key, store.NewObj(i, -1, core.ObjTypeString, core.ObjEncodingInt))
		if !qr.Insert(key, store) {
			t.Errorf("insert failed on element %d, expected successful insert", i)
		}
	}

	keyOverflow := "key_overflow"
	store.Put(keyOverflow, store.NewObj(9999, -1, core.ObjTypeString, core.ObjEncodingInt))
	if qr.Insert(keyOverflow, store) {
		t.Errorf("insert succeeded on element %d, expected failure due to maxElements limit", core.MaxQueueSize)
	}

	for i := 0; i < core.MaxQueues-1; i++ {
		_, err := core.NewQueueRef()
		if err != nil {
			t.Errorf("error creating QueueRef: %v", err)
		}
	}

	_, err = core.NewQueueRef()
	if err == nil {
		t.Errorf("expected error creating QueueRef, got %v", err)
	}
}

func TestQueueRefLen(t *testing.T) {
	store := core.NewStore()

	qr, err := core.NewQueueRef()
	if err != nil {
		t.Fatalf("error creating queue: %v", err)
	}
	mockTime := &utils.MockClock{CurrTime: time.Now()}
	utils.CurrentTime = mockTime

	store.Put("key1", store.NewObj(10, -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key1", store)

	store.Put("key2", store.NewObj(20, 10, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key2", store)

	store.Put("key3", store.NewObj(30, -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key3", store)

	store.Put("key4", store.NewObj(40, -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key4", store)

	mockTime.SetTime(time.Now().Add(10 * time.Millisecond))
	store.Del("key3")

	store.Put("key5", store.NewObj(50, -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key5", store)

	length := qr.Length(store)

	expectedLength := int64(3)
	if length != expectedLength {
		t.Errorf("expected queue length %d, got %d", expectedLength, length)
	}
}

// Benchmark queueref by inserting expired, non-expired and expired keys in order and removing them
func BenchmarkQueueRef(b *testing.B) {
	b.Run("Insert and Remove", benchmarkQueueRefInsertAndRemove)
}

func benchmarkQueueRefInsertAndRemove(b *testing.B) {
	store := core.NewStore()
	mockTime := &utils.MockClock{CurrTime: time.Now()}
	utils.CurrentTime = mockTime
	benchmarkCases := []struct {
		name            string
		nonExpiredCount int
	}{
		{"small", 10},
		{"medium", 100},
		{"large", 1000},
	}

	for _, benchmark := range benchmarkCases {
		b.Run(fmt.Sprintf("Benchmark %s", benchmark.name), func(b *testing.B) {
			qr, err := core.NewQueueRef()
			if err != nil {
				b.Fatal(err)
			}
			expiredCount := benchmark.nonExpiredCount / 2
			nonExpiredCount := benchmark.nonExpiredCount
			config.KeysLimit = benchmark.nonExpiredCount * 10
			store.ResetStore()

			// Insert expired keys
			b.Run("InsertExpiredKeys", func(b *testing.B) {
				for i := 0; i < expiredCount; i++ {
					insertKey(b, qr, i, true, store)
				}
			})

			// Insert non-expired keys
			b.Run("InsertNonExpiredKeys", func(b *testing.B) {
				for i := 0; i < nonExpiredCount; i++ {
					insertKey(b, qr, expiredCount+i, false, store)
				}
			})
			b.StopTimer()
			// Allow expired keys to expire
			mockTime.SetTime(time.Now().Add(2 * time.Millisecond))
			b.StartTimer()

			// Remove only non-expired number of keys (expired keys will be auto-removed)
			b.Run("RemoveNonExpiredKeys", func(b *testing.B) {
				for i := 0; i < nonExpiredCount; i++ {
					_, err := qr.Remove(store)
					if err != nil {
						b.Fatalf("Unexpected error during removal: %v", err)
					}
				}
			})

			b.StopTimer()
			// Validate that the queue is empty
			validateQueueEmpty(b, qr, store)
		})
	}
}

func insertKey(b *testing.B, qr *core.QueueRef, i int, expired bool, store *core.Store) {
	key := fmt.Sprintf("k%d", i)
	value := fmt.Sprintf("v%d", i)
	var expiration int64
	if expired {
		expiration = 1 // 1ms expiration
	} else {
		expiration = -1 // No expiration
	}

	store.Put(key, store.NewObj(value, expiration, core.ObjTypeString, core.ObjEncodingInt))
	if !qr.Insert(key, store) {
		b.Fatalf("Failed to insert key: %s", key)
	}
}

func validateQueueEmpty(b *testing.B, qr *core.QueueRef, store *core.Store) {
	_, err := qr.Remove(store)
	if err != core.ErrQueueEmpty {
		b.Fatalf("Queue not empty after benchmark. Got error: %v, Queue size: %d", err, qr.Length(store))
	}
}
