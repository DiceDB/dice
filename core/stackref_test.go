package core_test

import (
	"testing"

	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/testutils"
)

func TestStackRef(t *testing.T) {
	sr := core.NewStackRef()
	store := core.NewStore()

	if _, err := sr.Pop(); err != core.ErrStackEmpty {
		t.Error("popping from an empty stackref should return an empty stack error")
	}

	if sr.Push("does not exist") != false {
		t.Error("inserting the reference of the key that does not exist should not work but it did")
	}

	store.Put("key that exists", core.NewObj(10, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT), -1)
	if sr.Push("key that exists") != true {
		t.Error("inserting the reference of the key that exists should work but it did not")
	}

	if qe, _ := sr.Pop(); qe.Obj.Value != 10 {
		t.Error("removing from the stackref should return the obj")
	}

	if _, err := sr.Pop(); err != core.ErrStackEmpty {
		t.Error("removing again from an empty stackref should return an empty stack error")
	}

	store.Put("key1", core.NewObj(10, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT), -1)
	sr.Push("key1")
	store.Put("key2", core.NewObj(20, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT), -1)
	sr.Push("key2")
	store.Put("key3", core.NewObj(30, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT), -1)
	sr.Push("key3")
	store.Put("key4", core.NewObj(40, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT), -1)
	sr.Push("key4")
	store.Put("key5", core.NewObj(50, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT), -1)
	sr.Push("key5")
	store.Put("key6", core.NewObj(60, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT), -1)
	sr.Push("key6")
	store.Put("key7", core.NewObj(60, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT), -1)
	sr.Push("key7")

	observedVals := make([]int64, 0, 6)
	for _, se := range sr.Iterate(6) {
		observedVals = append(observedVals, int64(se.Obj.Value.(int)))
	}

	expectedVals := []int64{60, 60, 50, 40, 30, 20}
	if !testutils.EqualInt64Slice(observedVals, expectedVals) {
		t.Error("iterating through the stackref should return the elements in the order they were pushed. Expected:", expectedVals, ", got: ", observedVals)
	}
}
