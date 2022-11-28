package core_test

import (
	"testing"

	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/testutils"
)

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
}
