package core_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/server/utils"
	"github.com/dicedb/dice/testutils"
)

func CreateAndPushKeyToStack(sr *core.StackRef, key string, value int, expDurationMs int64, store *core.Store) {
	obj := store.NewObj(value, expDurationMs, core.ObjTypeString, core.ObjEncodingInt)
	store.Put(key, obj)
	sr.Push(key, store)
}

func TestStackRef(t *testing.T) {
	store := core.NewStore()
	sr, _ := core.NewStackRef()

	if _, err := sr.Pop(store); err != core.ErrStackEmpty {
		t.Error("popping from an empty stackref should return an empty stack error")
	}

	if sr.Push("does not exist", store) != false {
		t.Error("inserting the reference of the key that does not exist should not work but it did")
	}

	store.Put("key that exists", store.NewObj(10, -1, core.ObjTypeString, core.ObjEncodingInt))
	if sr.Push("key that exists", store) != true {
		t.Error("inserting the reference of the key that exists should work but it did not")
	}

	if qe, _ := sr.Pop(store); qe.Obj.Value != 10 {
		t.Error("removing from the stackref should return the obj")
	}

	if _, err := sr.Pop(store); err != core.ErrStackEmpty {
		t.Error("removing again from an empty stackref should return an empty stack error")
	}

	store.Put("key1", store.NewObj(10, -1, core.ObjTypeString, core.ObjEncodingInt))
	sr.Push("key1", store)
	store.Put("key2", store.NewObj(20, -1, core.ObjTypeString, core.ObjEncodingInt))
	sr.Push("key2", store)
	store.Put("key3", store.NewObj(30, -1, core.ObjTypeString, core.ObjEncodingInt))
	sr.Push("key3", store)
	store.Put("key4", store.NewObj(40, -1, core.ObjTypeString, core.ObjEncodingInt))
	sr.Push("key4", store)
	store.Put("key5", store.NewObj(50, -1, core.ObjTypeString, core.ObjEncodingInt))
	sr.Push("key5", store)
	store.Put("key6", store.NewObj(60, -1, core.ObjTypeString, core.ObjEncodingInt))
	sr.Push("key6", store)
	store.Put("key7", store.NewObj(60, -1, core.ObjTypeString, core.ObjEncodingInt))
	sr.Push("key7", store)

	observedVals := make([]int64, 0, 6)
	for _, se := range sr.Iterate(6, store) {
		observedVals = append(observedVals, int64(se.Obj.Value.(int)))
	}

	expectedVals := []int64{60, 60, 50, 40, 30, 20}
	if !testutils.EqualInt64Slice(observedVals, expectedVals) {
		t.Error("iterating through the stackref should return the elements in the order they were pushed. Expected:", expectedVals, ", got: ", observedVals)
	}

	sr2, _ := core.NewStackRef()
	obj1 := store.NewObj(10, -1, core.ObjTypeString, core.ObjEncodingInt)
	store.Put("key1", obj1)
	sr2.Push("key1", store)

	if key, err := sr2.Pop(store); err != nil || key.Obj != store.Get("key1") {
		t.Error("popping non-expired key from stackref should pop element")

	}

	keys := []struct {
		key           string
		value         int
		expDurationMs int64
	}{
		{"key0", 0, 0},     // expired key
		{"key1", 10, 0},    // expired key
		{"key2", 20, -1},   // non-expired key
		{"key3", 30, 0},    // expired key
		{"key4", 40, 0},    // expired key
		{"key5", 50, 0},    // expired key
		{"key6", 60, -1},   // non-expired key
		{"key7", 70, 0},    // expired key
		{"key8", 80, -1},   // non-expired key
		{"key9", 90, -1},   // non-expired key
		{"key10", 100, -1}, // non-expired key
	}

	for _, k := range keys {
		CreateAndPushKeyToStack(sr2, k.key, k.value, k.expDurationMs, store)
	}

	expectedVals = []int64{100, 90, 80}
	observedVals = make([]int64, 0, 3)

	for i := 0; i < len(expectedVals); i++ {
		key, err := sr2.Pop(store)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if key == nil {
			t.Errorf("expected key, got nil")
		} else {
			observedVals = append(observedVals, int64(key.Obj.Value.(int)))
		}
	}

	if !testutils.EqualInt64Slice(observedVals, expectedVals) {
		t.Errorf("expected %v, got %v", expectedVals, observedVals)
	}

	if key, err := sr2.Pop(store); err != nil || int64(key.Obj.Value.(int)) != 60 {
		t.Errorf("popping an expired key followed by a non-expired key should work")
	}

	if key, err := sr2.Pop(store); err != nil || int64(key.Obj.Value.(int)) != 20 {
		t.Errorf("popping multiple expired keys followed by a non-expired key should work")
	}

	if _, err := sr2.Pop(store); err != core.ErrStackEmpty {
		t.Errorf("popping all expired keys till stack becomes empty should work")
	}

}

func TestStackRefLen(t *testing.T) {
	store := core.NewStore()

	sr, err := core.NewStackRef()
	if err != nil {
		t.Fatalf("error creating stack: %v", err)
	}
	mockTime := &utils.MockClock{CurrTime: time.Now()}
	utils.CurrentTime = mockTime

	store.Put("key1", store.NewObj(10, -1, core.ObjTypeString, core.ObjEncodingInt))
	sr.Push("key1", store)

	store.Put("key2", store.NewObj(20, 10, core.ObjTypeString, core.ObjEncodingInt))
	sr.Push("key2", store)

	store.Put("key3", store.NewObj(30, -1, core.ObjTypeString, core.ObjEncodingInt))
	sr.Push("key3", store)

	store.Put("key4", store.NewObj(40, -1, core.ObjTypeString, core.ObjEncodingInt))
	sr.Push("key4", store)

	mockTime.SetTime(time.Now().Add(10 * time.Millisecond))
	store.Del("key3")

	store.Put("key5", store.NewObj(50, -1, core.ObjTypeString, core.ObjEncodingInt))
	sr.Push("key5", store)

	length := sr.Length(store)

	expectedLength := int64(3)
	if length != expectedLength {
		t.Errorf("expected stack length %d, got %d", expectedLength, length)
	}

}
func TestStackRefMaxConstraints(t *testing.T) {
	config.KeysLimit = 20000000
	core.WatchChannel = make(chan core.WatchEvent, config.KeysLimit)
	store := core.NewStore()
	core.StackCount = 0
	sr, err := core.NewStackRef()
	if err != nil {
		t.Errorf("error creating StackRef: %v", err)
	}

	for i := 0; i < config.MaxStackSize; i++ {
		key := fmt.Sprintf("key%d", i)
		store.Put(key, store.NewObj(i, -1, core.ObjTypeString, core.ObjEncodingInt))
		if !sr.Push(key, store) {
			t.Errorf("push failed on element %d, expected successful push", i)
		}
	}
	for i := 0; i < config.MaxStacks-1; i++ {
		_, err := core.NewStackRef()
		if err != nil {
			t.Errorf("error creating StackRef: %v", err)
		}
	}

	_, err = core.NewStackRef()
	if err == nil {
		t.Errorf("expected error creating StackRef, got %v", err)
	}
}

func BenchmarkRemoveLargeNumberOfExpiredKeys(b *testing.B) {
	store := core.NewStore()
	for _, v := range benchmarkDataSizesStackQueue {
		config.KeysLimit = 20000000 // Set a high limit for benchmarking
		core.WatchChannel = make(chan core.WatchEvent, config.KeysLimit)
		core.StackCount = 0
		config.MaxStacks = 2000000 // Set a high limit for benchmarking
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				var sr *core.StackRef
				var err error
				if sr, err = core.NewStackRef(); err != nil {
					b.Fatal(err)
				}
				CreateAndPushKeyToStack(sr, "initialKey", 0, -1, store)
				for j := 1; j < v; j++ {
					CreateAndPushKeyToStack(sr, strconv.Itoa(j), j, 0, store)
				}
				b.StartTimer()
				if _, err := sr.Pop(store); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
