package tests

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/dicedb/dice/internal/constants"

	"github.com/bytedance/sonic"

	"github.com/dicedb/dice/core"
	redis "github.com/dicedb/go-dice"
	"gotest.tools/v3/assert"
)

type qWatchTestCase struct {
	userID          int
	score           int
	expectedUpdates [][]interface{}
}

var qWatchQuery = "SELECT $key, $value FROM `match:100:*` ORDER BY $value desc LIMIT 3"

var qWatchTestCases = []qWatchTestCase{
	{0, 11, [][]interface{}{
		{[]interface{}{"match:100:user:0", int64(11)}},
	}},
	{1, 33, [][]interface{}{
		{[]interface{}{"match:100:user:1", int64(33)}, []interface{}{"match:100:user:0", int64(11)}},
	}},
	{2, 22, [][]interface{}{
		{[]interface{}{"match:100:user:1", int64(33)}, []interface{}{"match:100:user:2", int64(22)}, []interface{}{"match:100:user:0", int64(11)}},
	}},
	{3, 0, [][]interface{}{
		{[]interface{}{"match:100:user:1", int64(33)}, []interface{}{"match:100:user:2", int64(22)}, []interface{}{"match:100:user:0", int64(11)}},
	}},
	{4, 44, [][]interface{}{
		{[]interface{}{"match:100:user:4", int64(44)}, []interface{}{"match:100:user:1", int64(33)}, []interface{}{"match:100:user:2", int64(22)}},
	}},
	{5, 50, [][]interface{}{
		{[]interface{}{"match:100:user:5", int64(50)}, []interface{}{"match:100:user:4", int64(44)}, []interface{}{"match:100:user:1", int64(33)}},
	}},
	{2, 40, [][]interface{}{
		{[]interface{}{"match:100:user:5", int64(50)}, []interface{}{"match:100:user:4", int64(44)}, []interface{}{"match:100:user:2", int64(40)}},
	}},
	{6, 55, [][]interface{}{
		{[]interface{}{"match:100:user:6", int64(55)}, []interface{}{"match:100:user:5", int64(50)}, []interface{}{"match:100:user:4", int64(44)}},
	}},
	{0, 60, [][]interface{}{
		{[]interface{}{"match:100:user:0", int64(60)}, []interface{}{"match:100:user:6", int64(55)}, []interface{}{"match:100:user:5", int64(50)}},
	}},
	{5, 70, [][]interface{}{
		{[]interface{}{"match:100:user:5", int64(70)}, []interface{}{"match:100:user:0", int64(60)}, []interface{}{"match:100:user:6", int64(55)}},
	}},
}

// TestQWATCH tests the QWATCH functionality using raw network connections.
func TestQWATCH(t *testing.T) {
	publisher := getLocalConnection()

	subscribers := []net.Conn{getLocalConnection(), getLocalConnection(), getLocalConnection()}

	// Cleanup Store for next tests
	defer func() {
		for _, tc := range qWatchTestCases {
			fireCommand(publisher, fmt.Sprintf("DEL match:100:user:%d", tc.userID))
		}

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
		assert.Equal(t, 3, len(v.([]interface{})))
	}

	runQWatchScenarios(t, publisher, respParsers)
}

// TestQWATCHWithSDK tests the QWATCH functionality using the Redis SDK.
func TestQWATCHWithSDK(t *testing.T) {
	ctx := context.Background()
	publisher := getLocalSdk()

	subscribers := []*redis.Client{getLocalSdk(), getLocalSdk(), getLocalSdk()}

	// Cleanup Store for next tests
	for _, tc := range qWatchTestCases {
		publisher.Del(context.Background(), fmt.Sprintf("match:100:user:%d", tc.userID))
	}

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
		//	Get the first message
		<-channels[i]
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
					assert.DeepEqual(t, []interface{}{constants.Qwatch, qWatchQuery, expectedUpdate}, update)
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
	publisher := getLocalConnection()

	// Cleanup Store for next tests
	for _, tc := range JSONTestCases {
		fireCommand(publisher, fmt.Sprintf("DEL %s", tc.key))
	}

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
		assert.Equal(t, 3, len(v.([]interface{})))
	}

	for i, tc := range JSONTestCases {
		fireCommand(publisher, fmt.Sprintf("JSON.SET %s $ %s", tc.key, tc.value))

		for _, expectedUpdate := range tc.expectedUpdates {
			rp := respParsers[i]

			v, err := rp.DecodeOne()
			assert.NilError(t, err)
			response := v.([]interface{})
			assert.Equal(t, 3, len(response))
			assert.Equal(t, constants.Qwatch, response[0])

			update := response[2].([]interface{})

			assert.Equal(t, len(expectedUpdate), len(update), fmt.Sprintf("Expected update: %v, got %v", expectedUpdate, update))
			assert.Equal(t, expectedUpdate[0].([]interface{})[0], update[0].([]interface{})[0], "Key mismatch")

			var expectedJSON, actualJSON interface{}
			assert.NilError(t, sonic.UnmarshalString(tc.value, &expectedJSON))
			assert.NilError(t, sonic.UnmarshalString(update[0].([]interface{})[1].(string), &actualJSON))
			assert.DeepEqual(t, expectedJSON, actualJSON)
		}
	}
}
