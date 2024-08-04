package core_test

import (
	"fmt"
	"testing"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/testutils"
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
	if qe, err := qr.Remove(); err != nil || qe.Obj.Value != val {
		t.Error("removal from queue with single non expired key failed, Expected : ", val, " Got : ", qe.Obj.Value)
	}
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
		if qe, err := qr.Remove(); err != nil || qe.Obj.Value != val[i] {
			t.Error("removal from queue with multiple non expired keys failed, Expected : ", val[i], " Got : ", qe.Obj.Value)
			break
		}
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
	if qe, err := qr.Remove(); err != nil || qe.Obj.Value != val[1] {
		t.Error("test for removing expired key before non-expired failed , Expected : ", val[1], " Got : ", qe.Obj.Value)
	}
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
	if qe, err := qr.Remove(); err != nil || qe.Obj.Value != val[2] {
		t.Error("test for removing mulitple expired key before non-expired failed , Expected : ", val[2], " Got : ", qe.Obj.Value)
	}
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
	if _, err := qr.Remove(); err == nil {
		t.Error("test for removing all expired key failed, Expected : ", core.ErrQueueEmpty, " Got : ", err)
	}
	// remove from empty queue
	if _, err := qr.Remove(); err == nil {
		fmt.Println(err)
		t.Error("test for removing from empty queue failed Expected : ", core.ErrQueueEmpty, " Got : ", err)
	}
}

// Benchmark queueref by inserting expired, non-expired and expired keys in order and removing them
func BenchmarkQueueRef(b *testing.B) {
	benchmarkCount := [][]int{{10, 20, 30}, {100, 200, 300}, {100, 2000, 3000}, {10000, 20000, 30000}}

	for _, count := range benchmarkCount {
		qr := core.NewQueueRef()
		config.KeysLimit = 100000000 // override keys limit high for benchmarking

		// need to init again for each round with the overriden buffer size
		// otherwise the watchchannel buffer size will stay as it is with the global keylimits size
		core.WatchChannel = make(chan core.WatchEvent, config.KeysLimit)
		b.Run(fmt.Sprintf("Benchmark queue with expired : %d, non-expired : %d, expired : %d keys", count[0],
			count[1], count[2]), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// total number of keys , expired + non-expired + expired
				total := count[0] + count[1] + count[2]
				dataset := []tupple{}

				for i := 0; i < total; i++ {
					dataset = append(dataset, tupple{fmt.Sprintf("k%d", i), fmt.Sprintf("v%d", i)})
				}

				// Delete all keys
				for _, data := range dataset {
					obj := core.Get(data.key)
					if obj != nil {
						core.Del(data.key)
					}
				}

				// Insert all keys expired keys first
				for i := 0; i < count[0]; i++ {
					data := dataset[i]
					core.Put(data.key, core.NewObj(data.value, 1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
					qr.Insert(data.key)
				}

				// Insert all keys non-expired keys in the middle
				for i := count[0]; i < count[0]+count[1]; i++ {
					data := dataset[i]
					core.Put(data.key, core.NewObj(data.value, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
					qr.Insert(data.key)
				}

				// Insert all keys expired keys last
				for i := count[0] + count[1]; i < count[1]+count[2]; i++ {
					data := dataset[i]
					core.Put(data.key, core.NewObj(data.value, 1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
					qr.Insert(data.key)
				}
				time.Sleep(2 * time.Millisecond)

				for i := 0; i < count[1]; i++ {
					expectedValue := fmt.Sprintf("v%d", count[0]+i)
					data, err := qr.Remove()
					//fmt.Printf("Data %s\n", data.Obj.Value)
					if data.Obj.Value != expectedValue {
						b.Fatal(err)
					}
				}
				// check if queue is empty
				if _, err := qr.Remove(); err != core.ErrQueueEmpty {
					b.Fatal("Expected : ", core.ErrQueueEmpty, " Got : ", err)
				}
			}
		})
	}
}
