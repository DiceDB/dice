package tests

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

func TestQWATCH(t *testing.T) {
	publisher := getLocalConnection()
	subscriber := getLocalConnection()

	rp := fireCommandAndGetRESPParser(subscriber, "QWATCH \"SELECT $key, $value FROM `match:100:*` ORDER BY $value DESC LIMIT 3\"")
	if rp == nil {
		t.Fail()
	}

	type testCase struct {
		userID          int
		score           int
		expectedUpdates [][]interface{}
	}

	testCases := []testCase{
		// Initial insertion.
		{ /*key=*/ 0, 11, [][]interface{}{
			{[]interface{}{"match:100:user:0", "11"}},
		}},
		// Insertion of a second key.
		{ /*key=*/ 1 /*value=*/, 33, [][]interface{}{
			{[]interface{}{"match:100:user:1", "33"}, []interface{}{"match:100:user:0", "11"}},
		}},
		// Insertion of a third key.
		{ /*key=*/ 2 /*value=*/, 22, [][]interface{}{
			{[]interface{}{"match:100:user:1", "33"}, []interface{}{"match:100:user:2", "22"}, []interface{}{"match:100:user:0", "11"}},
		}},
		// Insertion of a fourth key with value 0, should not affect the order.
		{ /*key=*/ 3 /*value=*/, 0, [][]interface{}{
			{[]interface{}{"match:100:user:1", "33"}, []interface{}{"match:100:user:2", "22"}, []interface{}{"match:100:user:0", "11"}},
		}},
		// Insertion of a fifth key with value 44, should reorder the list.
		{ /*key=*/ 4 /*value=*/, 44, [][]interface{}{
			{[]interface{}{"match:100:user:4", "44"}, []interface{}{"match:100:user:1", "33"}, []interface{}{"match:100:user:2", "22"}},
		}},
		// Insertion of a sixth key with value 50, should reorder the list.
		{ /*key=*/ 5 /*value=*/, 50, [][]interface{}{
			{[]interface{}{"match:100:user:5", "50"}, []interface{}{"match:100:user:4", "44"}, []interface{}{"match:100:user:1", "33"}},
		}},
		// Update an existing key, should reorder the list.
		{ /*key=*/ 2 /*value=*/, 40, [][]interface{}{
			{[]interface{}{"match:100:user:5", "50"}, []interface{}{"match:100:user:4", "44"}, []interface{}{"match:100:user:2", "40"}},
		}},
		// Add a new key with value 55, should reorder the list.
		{ /*key=*/ 6 /*value=*/, 55, [][]interface{}{
			{[]interface{}{"match:100:user:6", "55"}, []interface{}{"match:100:user:5", "50"}, []interface{}{"match:100:user:4", "44"}},
		}},
		// Update the value of an existing key, should reorder the list.
		{ /*key=*/ 0 /*value=*/, 60, [][]interface{}{
			{[]interface{}{"match:100:user:0", "60"}, []interface{}{"match:100:user:6", "55"}, []interface{}{"match:100:user:5", "50"}},
		}},
		// Reuse another user, should reorder the list
		{ /*key=*/ 5 /*value=*/, 70, [][]interface{}{
			{[]interface{}{"match:100:user:5", "70"}, []interface{}{"match:100:user:0", "60"}, []interface{}{"match:100:user:6", "55"}},
		}},
	}

	// Read first message (OK)
	v, err := rp.DecodeOne()
	assert.NilError(t, err)
	assert.Equal(t, "OK", v.(string))

	for _, tc := range testCases {
		// Set the value for the userID
		fireCommand(publisher, fmt.Sprintf("SET match:100:user:%d %d", tc.userID, tc.score))

		for _, expectedUpdate := range tc.expectedUpdates {
			// Check if the update is received by the subscriber.
			v, err := rp.DecodeOne()
			assert.NilError(t, err)

			// Message format: [key, op, message]
			// Ensure the update matches the expected value.
			update := v.([]interface{})
			assert.DeepEqual(t, expectedUpdate, update)
		}
	}
}
