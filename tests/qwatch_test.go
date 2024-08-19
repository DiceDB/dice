package tests

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"net"
	"testing"

	"github.com/dicedb/dice/core"
	redis "github.com/dicedb/go-dice"
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
	// iterate over all the test cases and Delete the keys
	for _, tc := range qWatchTestCases {
		core.Del(fmt.Sprintf("match:100:user:%d", tc.userID))
	}
}

// TestQWATCH tests the QWATCH functionality using raw network connections.
func TestQWATCH(t *testing.T) {
	resetQWatchStore()
	publisher := getLocalConnection()
	subscribers := []net.Conn{getLocalConnection(), getLocalConnection(), getLocalConnection()}
	defer func() {
		publisher.Close()
		for _, sub := range subscribers {
			sub.Close()
		}
	}()

	respParsers := make([]*core.RESPParser, len(subscribers))

	// Subscribe to the QWATCH query
	for i, subscriber := range subscribers {
		rp := fireCommandAndGetRESPParser(subscriber, fmt.Sprintf("QWATCH \"%s\"", qWatchQuery))
		assert.Assert(t, rp != nil)
		respParsers[i] = rp

		v, err := rp.DecodeOne()
		assert.NilError(t, err)
		assert.Equal(t, 0, len(v.([]interface{})))
	}

	runQWatchScenarios(t, publisher, respParsers)
}

// TestQWATCHWithSDK tests the QWATCH functionality using the Redis SDK.
func TestQWATCHWithSDK(t *testing.T) {
	resetQWatchStore()
	ctx := context.Background()
	publisher := getLocalSdk()
	subscribers := []*redis.Client{getLocalSdk(), getLocalSdk(), getLocalSdk()}
	defer func() {
		publisher.Close()
		for _, sub := range subscribers {
			sub.Close()
		}
	}()

	channels := make([]<-chan *redis.QMessage, len(subscribers))

	// Subscribe to the QWATCH query
	for i, subscriber := range subscribers {
		qwatch := subscriber.QWatch(ctx)
		assert.Assert(t, qwatch != nil)
		err := qwatch.WatchQuery(ctx, qWatchQuery)
		assert.NilError(t, err)
		channels[i] = qwatch.Channel()
	}

	runQWatchScenarios(t, publisher, channels)
}

// runQWatchScenario executes the QWATCH test scenarios.
func runQWatchScenarios(t *testing.T, publisher interface{}, receivers interface{}) {
	for _, tc := range qWatchTestCases {
		// Publish updates based on the publisher type
		switch p := publisher.(type) {
		case net.Conn:
			fireCommand(p, fmt.Sprintf("SET match:100:user:%d %d", tc.userID, tc.score))
		case *redis.Client:
			err := p.Set(context.Background(), fmt.Sprintf("match:100:user:%d", tc.userID), tc.score, 0).Err()
			assert.NilError(t, err)
		}

		// For raw connections, parse RESP responses
		for _, expectedUpdate := range tc.expectedUpdates {
			switch r := receivers.(type) {
			case []*core.RESPParser:
				// For raw connections, parse RESP responses
				for _, rp := range r {
					v, err := rp.DecodeOne()
					assert.NilError(t, err)
					update := v.([]interface{})
					assert.DeepEqual(t, expectedUpdate, update)
				}
			case []<-chan *redis.QMessage:
				// For raw connections, parse RESP responses
				for _, ch := range r {
					v := <-ch
					assert.Equal(t, len(v.Updates), len(expectedUpdate))
					for i, update := range v.Updates {
						assert.DeepEqual(t, expectedUpdate[i], []interface{}{update.Key, update.Value})
					}
				}
			}
		}
	}
}

var JSONTestCases = []struct {
	key             string
	value           string
	qwatchQuery     string
	expectedUpdates [][]interface{}
}{
	{
		key:         "match:100:user:0",
		value:       `{"name":"Tom"}`,
		qwatchQuery: "SELECT $key, $value FROM `match:100:user:0` WHERE '$value.name' = 'Tom'",
		expectedUpdates: [][]interface{}{
			{[]interface{}{"match:100:user:0", map[string]interface{}{"name": "Tom"}}},
		},
	},
	{
		key:         "match:100:user:1",
		value:       `{"name":"Tom","age":24}`,
		qwatchQuery: "SELECT $key, $value FROM `match:100:user:1` WHERE '$value.age' > 20",
		expectedUpdates: [][]interface{}{
			{[]interface{}{"match:100:user:1", map[string]interface{}{"name": "Tom", "age": float64(24)}}},
		},
	},
	{
		key:         "match:100:user:2",
		value:       `{"score":10.36}`,
		qwatchQuery: "SELECT $key, $value FROM `match:100:user:2` WHERE '$value.score' = 10.36",
		expectedUpdates: [][]interface{}{
			{[]interface{}{"match:100:user:2", map[string]interface{}{"score": 10.36}}},
		},
	},
	{
		key:         "match:100:user:3",
		value:       `{"field1":{"field2":{"field3":{"score":10.36}}}}`,
		qwatchQuery: "SELECT $key, $value FROM `match:100:user:3` WHERE '$value.field1.field2.field3.score' > 10.1",
		expectedUpdates: [][]interface{}{
			{[]interface{}{"match:100:user:3", map[string]interface{}{
				"field1": map[string]interface{}{
					"field2": map[string]interface{}{
						"field3": map[string]interface{}{
							"score": 10.36,
						},
					},
				},
			}}},
		},
	},
}

func TestQwatchWithJSON(t *testing.T) {
	resetJSONTestCases()
	publisher := getLocalConnection()

	subscribers := make([]net.Conn, 0, len(JSONTestCases))

	for i := 0; i < len(JSONTestCases); i++ {
		subscribers = append(subscribers, getLocalConnection())
	}

	defer func() {
		publisher.Close()
		for _, sub := range subscribers {
			sub.Close()
		}
	}()

	respParsers := make([]*core.RESPParser, len(subscribers))

	for i, testCase := range JSONTestCases {
		rp := fireCommandAndGetRESPParser(subscribers[i], fmt.Sprintf("QWATCH \"%s\"", testCase.qwatchQuery))
		assert.Assert(t, rp != nil)
		respParsers[i] = rp

		v, err := rp.DecodeOne()
		assert.NilError(t, err)
		assert.Equal(t, 0, len(v.([]interface{})))
	}

	for i, tc := range JSONTestCases {
		fireCommand(publisher, fmt.Sprintf("JSON.SET %s $ %s", tc.key, tc.value))

		for _, expectedUpdate := range tc.expectedUpdates {
			rp := respParsers[i]

			v, err := rp.DecodeOne()
			assert.NilError(t, err)
			update := v.([]interface{})

			assert.Equal(t, len(expectedUpdate), len(update), fmt.Sprintf("Expected update: %v, got %v", expectedUpdate, update))
			assert.Equal(t, expectedUpdate[0].([]interface{})[0], update[0].([]interface{})[0], "Key mismatch")

			var expectedJSON, actualJSON interface{}
			assert.NilError(t, sonic.UnmarshalString(tc.value, &expectedJSON))
			assert.NilError(t, sonic.UnmarshalString(update[0].([]interface{})[1].(string), &actualJSON))
			assert.DeepEqual(t, expectedJSON, actualJSON)
		}
	}
}

func resetJSONTestCases() {
	for _, testCase := range JSONTestCases {
		core.Del(testCase.key)
	}
}
