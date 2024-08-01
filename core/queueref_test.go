package core_test

import (
	"testing"

	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/testutils"
)

// type queuereftcase struct {
// 	op            byte
// 	key           string
// 	value         interface{}
// 	expectedValue bool
// 	expectedError error
// 	list          []interface{}
// }

func TestQueueRef(t *testing.T) {
	store := core.NewStore()
	qr := core.NewQueueRef(store)

	if _, err := qr.Remove(); err != core.ErrQueueEmpty {
		t.Error("removing from an empty queueref should return an empty queue error")
	}

	if qr.Insert("does not exist") != false {
		t.Error("inserting the reference of the key that does not exist should not work but it did")
	}

	store.Put("key that exists", core.NewObj(10, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT), -1)
	if qr.Insert("key that exists") != true {
		t.Error("inserting the reference of the key that exists should work but it did not")
	}

	if qe, _ := qr.Remove(); qe.Obj.Value != 10 {
		t.Error("removing from the queueref should return the obj")
	}

	if _, err := qr.Remove(); err != core.ErrQueueEmpty {
		t.Error("removing again from an empty queueref should return an empty queue error")
	}

	store.Put("key1", core.NewObj(10, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT), -1)
	qr.Insert("key1")
	store.Put("key2", core.NewObj(20, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT), -1)
	qr.Insert("key2")
	store.Put("key3", core.NewObj(30, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT), -1)
	qr.Insert("key3")
	store.Put("key4", core.NewObj(40, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT), -1)
	qr.Insert("key4")
	store.Put("key5", core.NewObj(50, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT), -1)
	qr.Insert("key5")
	store.Put("key6", core.NewObj(60, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT), -1)
	qr.Insert("key6")
	store.Put("key7", core.NewObj(60, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT), -1)
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
