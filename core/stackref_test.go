package core_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/testutils"
)

func createAndPushKey(sr *core.StackRef, key string, value int, expDurationMs int64) {
	obj := core.NewObj(value, expDurationMs, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT)
	core.Put(key, obj)
	sr.Push(key)
}

func TestStackRef(t *testing.T) {
	sr := core.NewStackRef()

	if _, err := sr.Pop(); err != core.ErrStackEmpty {
		t.Error("popping from an empty stackref should return an empty stack error")
	}

	if sr.Push("does not exist") != false {
		t.Error("inserting the reference of the key that does not exist should not work but it did")
	}

	core.Put("key that exists", core.NewObj(10, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	if sr.Push("key that exists") != true {
		t.Error("inserting the reference of the key that exists should work but it did not")
	}

	if qe, _ := sr.Pop(); qe.Obj.Value != 10 {
		t.Error("removing from the stackref should return the obj")
	}

	if _, err := sr.Pop(); err != core.ErrStackEmpty {
		t.Error("removing again from an empty stackref should return an empty stack error")
	}

	core.Put("key1", core.NewObj(10, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	sr.Push("key1")
	core.Put("key2", core.NewObj(20, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	sr.Push("key2")
	core.Put("key3", core.NewObj(30, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	sr.Push("key3")
	core.Put("key4", core.NewObj(40, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	sr.Push("key4")
	core.Put("key5", core.NewObj(50, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	sr.Push("key5")
	core.Put("key6", core.NewObj(60, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	sr.Push("key6")
	core.Put("key7", core.NewObj(60, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	sr.Push("key7")

	observedVals := make([]int64, 0, 6)
	for _, se := range sr.Iterate(6) {
		observedVals = append(observedVals, int64(se.Obj.Value.(int)))
	}

	expectedVals := []int64{60, 60, 50, 40, 30, 20}
	if !testutils.EqualInt64Slice(observedVals, expectedVals) {
		t.Error("iterating through the stackref should return the elements in the order they were pushed. Expected:", expectedVals, ", got: ", observedVals)
	}

	sr2 := core.NewStackRef()
	obj1 := core.NewObj(10, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT)
	core.Put("key1", obj1)
	sr2.Push("key1")

	if key, err := sr2.Pop(); err != nil || key.Obj != core.Get("key1") {
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
		createAndPushKey(sr2, k.key, k.value, k.expDurationMs)
	}

	expectedVals = []int64{100, 90, 80}
	observedVals = make([]int64, 0, 3)

	for i := 0; i < len(expectedVals); i++ {
		key, err := sr2.Pop()
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

	if key, err := sr2.Pop(); err != nil || int64(key.Obj.Value.(int)) != 60 {
		t.Errorf("popping an expired key followed by a non-expired key should work")
	}

	if key, err := sr2.Pop(); err != nil || int64(key.Obj.Value.(int)) != 20 {
		t.Errorf("popping multiple expired keys followed by a non-expired key should work")
	}

	if _, err := sr2.Pop(); err != core.ErrStackEmpty {
		t.Errorf("popping all expired keys till stack becomes empty should work")
	}

}

func BenchmarkRemoveLargeNumberOfExpiredKeys(b *testing.B) {

	for _, v := range keys {
		populateData(v.count)

		b.Run(fmt.Sprintf("keys_%d", v.count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				sr := core.NewStackRef()
				createAndPushKey(sr, "initialKey", 0, -1)
				for j := 1; j < v.count; j++ {
					createAndPushKey(sr, strconv.Itoa(j), j, 0)
				}
				b.StartTimer()
				if _, err := sr.Pop(); err == core.ErrStackEmpty {
					continue
				}
			}
		})
	}
}
