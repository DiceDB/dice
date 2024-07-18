package tests

import (
	"context"
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

type qWatchTestCase struct {
	userID          int
	score           int
	expectedUpdates [][]interface{}
}

var qWatchQuery = "SELECT $key, $value FROM `match:100:*` ORDER BY $value DESC LIMIT 3"

var qWatchTestCases = []qWatchTestCase{
	{0, 11, [][]interface{}{
		{[]interface{}{"match:100:user:0", "11"}},
	}},
	{1, 33, [][]interface{}{
		{[]interface{}{"match:100:user:1", "33"}, []interface{}{"match:100:user:0", "11"}},
	}},
	{2, 22, [][]interface{}{
		{[]interface{}{"match:100:user:1", "33"}, []interface{}{"match:100:user:2", "22"}, []interface{}{"match:100:user:0", "11"}},
	}},
	{3, 0, [][]interface{}{
		{[]interface{}{"match:100:user:1", "33"}, []interface{}{"match:100:user:2", "22"}, []interface{}{"match:100:user:0", "11"}},
	}},
	{4, 44, [][]interface{}{
		{[]interface{}{"match:100:user:4", "44"}, []interface{}{"match:100:user:1", "33"}, []interface{}{"match:100:user:2", "22"}},
	}},
	{5, 50, [][]interface{}{
		{[]interface{}{"match:100:user:5", "50"}, []interface{}{"match:100:user:4", "44"}, []interface{}{"match:100:user:1", "33"}},
	}},
	{2, 40, [][]interface{}{
		{[]interface{}{"match:100:user:5", "50"}, []interface{}{"match:100:user:4", "44"}, []interface{}{"match:100:user:2", "40"}},
	}},
	{6, 55, [][]interface{}{
		{[]interface{}{"match:100:user:6", "55"}, []interface{}{"match:100:user:5", "50"}, []interface{}{"match:100:user:4", "44"}},
	}},
	{0, 60, [][]interface{}{
		{[]interface{}{"match:100:user:0", "60"}, []interface{}{"match:100:user:6", "55"}, []interface{}{"match:100:user:5", "50"}},
	}},
	{5, 70, [][]interface{}{
		{[]interface{}{"match:100:user:5", "70"}, []interface{}{"match:100:user:0", "60"}, []interface{}{"match:100:user:6", "55"}},
	}},
}

// Before each test, we need to reset the database.
func resetQWatchStore() {
	conn := getLocalConnection()
	// iterate over all the test cases and Delete the keys
	for _, tc := range qWatchTestCases {
		fireCommand(conn, fmt.Sprintf("DEL match:100:user:%d", tc.userID))
	}
}

func TestQWATCH(t *testing.T) {
	resetQWatchStore()

	publisher := getLocalConnection()
	subscriber := getLocalConnection()

	rp := fireCommandAndGetRESPParser(subscriber, fmt.Sprintf("QWATCH \"%s\"", qWatchQuery))
	if rp == nil {
		t.Fail()
	}

	// Read first message (OK)
	v, err := rp.DecodeOne()
	assert.NilError(t, err)
	assert.Equal(t, "OK", v.(string))

	for _, tc := range qWatchTestCases {
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

func TestQWATCHWithSDK(t *testing.T) {
	resetQWatchStore()
	ctx := context.Background()

	publisher := getLocalSdk()
	subscriber := getLocalSdk()

	qwatch := subscriber.QWatch(ctx)
	if qwatch == nil {
		t.Fail()
	}

	err := qwatch.WatchQuery(ctx, qWatchQuery)
	assert.NilError(t, err)

	ch := qwatch.Channel()

	for _, tc := range qWatchTestCases {
		// Set the value for the userID
		err := publisher.Set(ctx, fmt.Sprintf("match:100:user:%d", tc.userID), tc.score, 0).Err()
		assert.NilError(t, err)

		for _, expectedUpdate := range tc.expectedUpdates {
			// Check if the update is received by the subscriber.
			v := <-ch

			assert.Equal(t, len(v.Updates), len(expectedUpdate))

			for i, update := range v.Updates {
				assert.DeepEqual(t, expectedUpdate[i], []interface{}{update.Key, update.Value})
			}
		}
	}
}
