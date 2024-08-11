package core_test

import (
	"fmt"
	"testing"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/testutils"
	"gotest.tools/v3/assert"
	"math/rand"
	"time"
)

func TestQueueRef(t *testing.T) {
	qr := core.NewQueueRef()

	if _, err := qr.Remove(); err != core.ErrQueueEmpty {
		t.Error("removing from an empty queueref should return an empty queue error")
	}

	if qr.Insert("does not exist") != false {
		t.Error("inserting the reference of the key that does not exist should not work but it did")
	}

	core.Put("key that exists", core.NewObj(10, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	if qr.Insert("key that exists") != true {
		t.Error("inserting the reference of the key that exists should work but it did not")
	}

	if qe, _ := qr.Remove(); qe.Obj.Value != 10 {
		t.Error("removing from the queueref should return the obj")
	}

	if _, err := qr.Remove(); err != core.ErrQueueEmpty {
		t.Error("removing again from an empty queueref should return an empty queue error")
	}

	core.Put("key1", core.NewObj(10, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	qr.Insert("key1")
	core.Put("key2", core.NewObj(20, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	qr.Insert("key2")
	core.Put("key3", core.NewObj(30, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	qr.Insert("key3")
	core.Put("key4", core.NewObj(40, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	qr.Insert("key4")
	core.Put("key5", core.NewObj(50, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	qr.Insert("key5")
	core.Put("key6", core.NewObj(60, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	qr.Insert("key6")
	core.Put("key7", core.NewObj(60, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
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
	core.Put("key1", core.NewObj(val, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	qr.Insert("key1")
	qe, err := qr.Remove()
	assert.Check(t, err == nil || qe.Obj.Value == val, fmt.Sprintf("removal from queue with single non expired key failed, Expected : %d, Got : %d\n", val, qe.Obj.Value))
}

// Test for removing from queue with multiple non-expired keys
func TestRemoveMultipleNonExpiredKeys(t *testing.T) {
	qr := core.NewQueueRef()
	val := [3]int{10, 20, 30}
	core.Put("key1", core.NewObj(val[0], -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	qr.Insert("key1")
	core.Put("key2", core.NewObj(val[1], -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	qr.Insert("key2")
	core.Put("key3", core.NewObj(val[2], -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
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
	core.Put("key1", core.NewObj(val[0], 1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	qr.Insert("key1")
	core.Put("key2", core.NewObj(val[1], -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	qr.Insert("key2")
	time.Sleep(2 * time.Millisecond)
	qe, err := qr.Remove()
	assert.Check(t, err == nil || qe.Obj.Value == val[1], fmt.Sprintf("test for removing expired key before non-expired failed , Expected : %d, Got : %d\n", val[1], qe.Obj.Value))
}

// Test for removing from queue with multiple expired keys before non-expired
func TestRemoveMultipleExpiredBeforeNonExpire(t *testing.T) {
	qr := core.NewQueueRef()
	val := [3]int{10, 20, 30}
	core.Put("key1", core.NewObj(val[0], 1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	qr.Insert("key1")
	core.Put("key2", core.NewObj(val[1], 1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	qr.Insert("key2")
	core.Put("key3", core.NewObj(val[2], -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
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
	core.Put("key1", core.NewObj(val[0], 1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	qr.Insert("key1")
	core.Put("key2", core.NewObj(val[1], 1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	qr.Insert("key2")
	core.Put("key3", core.NewObj(val[2], 1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	qr.Insert("key3")
	time.Sleep(2 * time.Millisecond)
	// remove from empty queue
	_, err := qr.Remove()
	assert.Equal(t, err, core.ErrQueueEmpty, fmt.Sprintf("test for removing from empty queue failed Expected : %s, Got : %s\n", core.ErrQueueEmpty, err))
}

// Benchmark queueref by inserting expired, non-expired and expired keys in order and removing them
func BenchmarkQueueRef(b *testing.B) {
	// Setup code...
	qr := core.NewQueueRef()
	config.KeysLimit = 100000000 // override keys limit high for benchmarking

	// need to init again for each round with the overriden buffer size
	// otherwise the watchchannel buffer size will stay as it is with the global keylimits size
	core.WatchChannel = make(chan core.WatchEvent, config.KeysLimit)

	expiredCount := 0
	nonExpiredCount := 0
	b.ResetTimer()
	b.Run("Insertion of expired and non-expired keys", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("k%d", i)
			value := fmt.Sprintf("v%d", i)
			random := rand.Intn(2)
			// Insert keys without expiration
			expiration := -1
			if random == 1 {
				// Insert keys with expiration of 1ms
				expiration = 1
				expiredCount++
			} else {
				nonExpiredCount++
			}
			core.Put(key, core.NewObj(value, int64(expiration), core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
			qr.Insert(key)
		}
	})
	b.StopTimer()
	// Allow keys to expire
	time.Sleep(2 * time.Millisecond)
	b.StartTimer()

	b.Run("Removal of keys", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Removal benchmark...
			_, err := qr.Remove()
			if err == core.ErrQueueEmpty {
				break
			}
		}
	})
	b.StopTimer()
	_, err := qr.Remove()
	len := qr.Length()
	if err != core.ErrQueueEmpty {
		b.Fatalf("Expired : %v, Got : %v, Size : %v", core.ErrQueueEmpty, err, len)
	}
}
