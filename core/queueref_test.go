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
	core.Put("key2", core.NewObj(20, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	core.Put("key3", core.NewObj(30, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	core.Put("key4", core.NewObj(40, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	core.Put("key5", core.NewObj(50, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	core.Put("key6", core.NewObj(60, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))
	core.Put("key7", core.NewObj(60, -1, core.OBJ_TYPE_STRING, core.OBJ_ENCODING_INT))

	observedVals := make([]int64, 0, 6)
	for _, qe := range qr.Iterate(6) {
		observedVals = append(observedVals, qe.Obj.Value.(int64))
	}

	testutils.EqualInt64Slice(observedVals, []int64{10, 20, 30, 40, 50, 60})
}
