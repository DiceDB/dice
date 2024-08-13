package core_test

import (
	"fmt"
	"testing"

	"time"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/testutils"
	"gotest.tools/v3/assert"
)

func TestQueueRef(t *testing.T) {
	qr := core.NewQueueRef()

	if _, err := qr.Remove(); err != core.ErrQueueEmpty {
		t.Error("removing from an empty queueref should return an empty queue error")
	}

	if qr.Insert("does not exist") != false {
		t.Error("inserting the reference of the key that does not exist should not work but it did")
	}

	core.Put("key that exists", core.NewObj(10, -1, core.ObjTypeString, core.ObjEncodingInt))
	if qr.Insert("key that exists") != true {
		t.Error("inserting the reference of the key that exists should work but it did not")
	}

	if qe, _ := qr.Remove(); qe.Obj.Value != 10 {
		t.Error("removing from the queueref should return the obj")
	}

	if _, err := qr.Remove(); err != core.ErrQueueEmpty {
		t.Error("removing again from an empty queueref should return an empty queue error")
	}

	core.Put("key1", core.NewObj(10, -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key1")
	core.Put("key2", core.NewObj(20, -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key2")
	core.Put("key3", core.NewObj(30, -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key3")
	core.Put("key4", core.NewObj(40, -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key4")
	core.Put("key5", core.NewObj(50, -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key5")
	core.Put("key6", core.NewObj(60, -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key6")
	core.Put("key7", core.NewObj(60, -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key7")

	observedVals := make([]int64, 0, 6)
	for _, qe := range qr.Iterate(6) {
		observedVals = append(observedVals, int64(qe.Obj.Value.(int)))
	}

	expectedVals := []int64{10, 20, 30, 40, 50, 60}
	if !testutils.EqualInt64Slice(observedVals, expectedVals) {
		t.Error("iterating through the queueref should return the objs in the order they were added. Expected: ", expectedVals, " Got: ", observedVals)
	}
}

// Test for removing from queue with single non-expired keys
func TestRemoveSingleNonExpiredKeys(t *testing.T) {
	qr := core.NewQueueRef()
	val := 10
	core.Put("key1", core.NewObj(val, -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key1")
	qe, err := qr.Remove()
	assert.Check(t, err == nil || qe.Obj.Value == val, fmt.Sprintf("removal from queue with single non expired key failed, Expected : %d, Got : %d\n", val, qe.Obj.Value))
}

// Test for removing from queue with multiple non-expired keys
func TestRemoveMultipleNonExpiredKeys(t *testing.T) {
	qr := core.NewQueueRef()
	val := [3]int{10, 20, 30}
	core.Put("key1", core.NewObj(val[0], -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key1")
	core.Put("key2", core.NewObj(val[1], -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key2")
	core.Put("key3", core.NewObj(val[2], -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key3")
	for i := 0; i < len(val); i++ {
		qe, err := qr.Remove()
		assert.Check(t, err == nil && qe.Obj.Value == val[i], fmt.Sprintf("removal from queue with multiple non expired keys failed, Expected : %d , Got : %d\n", val[i], qe.Obj.Value))
	}
}

// Test for removing from queue with expired keys before non-expired
func TestRemoveExpiredBeforeNonExpire(t *testing.T) {
	qr := core.NewQueueRef()
	val := [3]int{10, 20}
	core.Put("key1", core.NewObj(val[0], 1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key1")
	core.Put("key2", core.NewObj(val[1], -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key2")
	time.Sleep(2 * time.Millisecond)
	qe, err := qr.Remove()
	assert.Check(t, err == nil || qe.Obj.Value == val[1], fmt.Sprintf("test for removing expired key before non-expired failed , Expected : %d, Got : %d\n", val[1], qe.Obj.Value))
}

// Test for removing from queue with multiple expired keys before non-expired
func TestRemoveMultipleExpiredBeforeNonExpire(t *testing.T) {
	qr := core.NewQueueRef()
	val := [3]int{10, 20, 30}
	core.Put("key1", core.NewObj(val[0], 1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key1")
	core.Put("key2", core.NewObj(val[1], 1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key2")
	core.Put("key3", core.NewObj(val[2], -1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key3")
	time.Sleep(2 * time.Millisecond)
	fmt.Printf("Queue size : %d\n", qr.Length())
	qe, err := qr.Remove()
	assert.Check(t, err == nil || qe.Obj.Value == val[2], fmt.Sprintf("test for removing mulitple expired key before non-expired failed , Expected : %d, Got %d\n", val[2], qe.Obj.Value))
}

// Test for removing from queue with all expired keys
func TestRemoveAllExpired(t *testing.T) {
	qr := core.NewQueueRef()
	val := [3]int{10, 20, 30}
	core.Put("key1", core.NewObj(val[0], 1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key1")
	core.Put("key2", core.NewObj(val[1], 1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key2")
	core.Put("key3", core.NewObj(val[2], 1, core.ObjTypeString, core.ObjEncodingInt))
	qr.Insert("key3")
	time.Sleep(2 * time.Millisecond)
	// remove from empty queue
	_, err := qr.Remove()
	assert.Equal(t, err, core.ErrQueueEmpty, fmt.Sprintf("test for removing from empty queue failed Expected : %s, Got : %s\n", core.ErrQueueEmpty, err))
}

// Benchmark queueref by inserting expired, non-expired and expired keys in order and removing them
func BenchmarkQueueRef(b *testing.B) {
	b.Run("Insert and Remove", benchmarkQueueRefInsertAndRemove)
}

func benchmarkQueueRefInsertAndRemove(b *testing.B) {
	qr := core.NewQueueRef()
	expiredCount := b.N / 2
	nonExpiredCount := b.N

	config.KeysLimit = b.N * 10
	core.ResetStore()

	// Insert expired keys
	b.Run("InsertExpiredKeys", func(b *testing.B) {
		for i := 0; i < expiredCount; i++ {
			insertKey(b, qr, i, true)
		}
	})

	// Insert non-expired keys
	b.Run("InsertNonExpiredKeys", func(b *testing.B) {
		for i := 0; i < nonExpiredCount; i++ {
			insertKey(b, qr, expiredCount+i, false)
		}
	})

	b.StopTimer()
	// Allow expired keys to expire
	time.Sleep(2 * time.Millisecond)
	b.StartTimer()

	// Remove only non-expired number of keys (expired keys will be auto-removed)
	b.Run("RemoveNonExpiredKeys", func(b *testing.B) {
		for i := 0; i < nonExpiredCount; i++ {
			_, err := qr.Remove()
			if err != nil {
				b.Fatalf("Unexpected error during removal: %v", err)
			}
		}
	})

	b.StopTimer()

	// Validate that the queue is empty
	validateQueueEmpty(b, qr)
}

func insertKey(b *testing.B, qr *core.QueueRef, i int, expired bool) {
	key := fmt.Sprintf("k%d", i)
	value := fmt.Sprintf("v%d", i)
	var expiration int64
	if expired {
		expiration = 1 // 1ms expiration
	} else {
		expiration = -1 // No expiration
	}

	core.Put(key, core.NewObj(value, expiration, core.ObjTypeString, core.ObjEncodingInt))
	if !qr.Insert(key) {
		b.Fatalf("Failed to insert key: %s", key)
	}
}

func validateQueueEmpty(b *testing.B, qr *core.QueueRef) {
	_, err := qr.Remove()
	if err != core.ErrQueueEmpty {
		b.Fatalf("Queue not empty after benchmark. Got error: %v, Queue size: %d", err, qr.Length())
	}
}
