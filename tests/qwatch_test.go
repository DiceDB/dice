package tests

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/internal/constants"
	redis "github.com/dicedb/go-dice"
	"gotest.tools/v3/assert"
	"net"
	"testing"
	"time"
)

type qWatchTestCase struct {
	userID          int
	score           int
	expectedUpdates [][]interface{}
}

type qWatchSDKSubscriber struct {
	client *redis.Client
	qwatch *redis.QWatch
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
	publisher, subscribers, cleanup := setupQWATCHTest(t)
	defer cleanup()

	respParsers := subscribeToQWATCH(t, subscribers)
	runQWatchScenarios(t, publisher, respParsers)
}

func TestQWATCHWithSDK(t *testing.T) {
	publisher, subscribers, cleanup := setupQWATCHTestWithSDK(t)
	defer cleanup()

	channels := subscribeToQWATCHWithSDK(t, subscribers)
	runQWatchScenarios(t, publisher, channels)
}

func setupQWATCHTest(t *testing.T) (net.Conn, []net.Conn, func()) {
	t.Helper()
	publisher := getLocalConnection()
	subscribers := []net.Conn{getLocalConnection(), getLocalConnection(), getLocalConnection()}

	cleanup := func() {
		cleanupKeys(publisher)
		if err := publisher.Close(); err != nil {
			t.Errorf("Error closing publisher connection: %v", err)
		}
		for _, sub := range subscribers {
			fireCommand(sub, fmt.Sprintf("QUNWATCH \"%s\"", qWatchQuery))
			time.Sleep(100 * time.Millisecond)
			if err := sub.Close(); err != nil {
				t.Errorf("Error closing subscriber connection: %v", err)
			}
		}
	}

	return publisher, subscribers, cleanup
}

func setupQWATCHTestWithSDK(t *testing.T) (*redis.Client, []qWatchSDKSubscriber, func()) {
	t.Helper()
	publisher := getLocalSdk()
	subscribers := []qWatchSDKSubscriber{{client: getLocalSdk()}, {client: getLocalSdk()}, {client: getLocalSdk()}}

	cleanup := func() {
		cleanupKeysWithSDK(publisher)
		if err := publisher.Close(); err != nil {
			t.Errorf("Error closing publisher connection: %v", err)
		}
		for _, sub := range subscribers {
			if err := sub.qwatch.UnwatchQuery(context.Background(), qWatchQuery); err != nil {
				t.Errorf("Error unwatching query: %v", err)
			}
			if err := sub.client.Close(); err != nil {
				t.Errorf("Error closing subscriber connection: %v", err)
			}
		}
	}

	return publisher, subscribers, cleanup
}

func subscribeToQWATCH(t *testing.T, subscribers []net.Conn) []*core.RESPParser {
	t.Helper()
	respParsers := make([]*core.RESPParser, len(subscribers))
	for i, subscriber := range subscribers {
		rp := fireCommandAndGetRESPParser(subscriber, fmt.Sprintf("QWATCH \"%s\"", qWatchQuery))
		assert.Assert(t, rp != nil)
		respParsers[i] = rp

		v, err := rp.DecodeOne()
		assert.NilError(t, err)
		castedValue, ok := v.([]interface{})
		if !ok {
			t.Errorf("Type assertion to []interface{} failed for value: %v", v)
			return nil
		}
		assert.Equal(t, 3, len(castedValue))
	}
	return respParsers
}

func subscribeToQWATCHWithSDK(t *testing.T, subscribers []qWatchSDKSubscriber) []<-chan *redis.QMessage {
	t.Helper()
	ctx := context.Background()
	channels := make([]<-chan *redis.QMessage, len(subscribers))
	for i, subscriber := range subscribers {
		qwatch := subscriber.client.QWatch(ctx)
		subscribers[i].qwatch = qwatch
		assert.Assert(t, qwatch != nil)
		err := qwatch.WatchQuery(ctx, qWatchQuery)
		assert.NilError(t, err)
		channels[i] = qwatch.Channel()
		<-channels[i] // Get the first message
	}
	return channels
}

func runQWatchScenarios(t *testing.T, publisher interface{}, receivers interface{}) {
	t.Helper()
	for _, tc := range qWatchTestCases {
		publishUpdate(t, publisher, tc)
		verifyUpdates(t, receivers, tc.expectedUpdates)
	}
}

func publishUpdate(t *testing.T, publisher interface{}, tc qWatchTestCase) {
	key := fmt.Sprintf("match:100:user:%d", tc.userID)
	switch p := publisher.(type) {
	case net.Conn:
		fireCommand(p, fmt.Sprintf("SET %s %d", key, tc.score))
	case *redis.Client:
		err := p.Set(context.Background(), key, tc.score, 0).Err()
		assert.NilError(t, err)
	}
}

func verifyUpdates(t *testing.T, receivers interface{}, expectedUpdates [][]interface{}) {
	for _, expectedUpdate := range expectedUpdates {
		switch r := receivers.(type) {
		case []*core.RESPParser:
			verifyRESPUpdates(t, r, expectedUpdate)
		case []<-chan *redis.QMessage:
			verifySDKUpdates(t, r, expectedUpdate)
		}
	}
}

func verifyRESPUpdates(t *testing.T, respParsers []*core.RESPParser, expectedUpdate []interface{}) {
	for _, rp := range respParsers {
		v, err := rp.DecodeOne()
		assert.NilError(t, err)
		update, ok := v.([]interface{})
		if !ok {
			t.Errorf("Type assertion to []interface{} failed for value: %v", v)
			return
		}
		assert.DeepEqual(t, []interface{}{constants.Qwatch, qWatchQuery, expectedUpdate}, update)
	}
}

func verifySDKUpdates(t *testing.T, channels []<-chan *redis.QMessage, expectedUpdate []interface{}) {
	for _, ch := range channels {
		v := <-ch
		assert.Equal(t, len(v.Updates), len(expectedUpdate), v.Updates)
		for i, update := range v.Updates {
			assert.DeepEqual(t, expectedUpdate[i], []interface{}{update.Key, update.Value})
		}
	}
}

type JSONTestCase struct {
	key             string
	value           string
	qwatchQuery     string
	expectedUpdates [][]interface{}
}

var JSONTestCases = []JSONTestCase{
	{
		key:         "match:200:user:0",
		value:       `{"name":"Tom"}`,
		qwatchQuery: "SELECT $key, $value FROM `match:200:user:0` WHERE '$value.name' = 'Tom'",
		expectedUpdates: [][]interface{}{
			{[]interface{}{"match:200:user:0", map[string]interface{}{"name": "Tom"}}},
		},
	},
	{
		key:         "match:200:user:1",
		value:       `{"name":"Tom","age":24}`,
		qwatchQuery: "SELECT $key, $value FROM `match:200:user:1` WHERE '$value.age' > 20",
		expectedUpdates: [][]interface{}{
			{[]interface{}{"match:200:user:1", map[string]interface{}{"name": "Tom", "age": float64(24)}}},
		},
	},
	{
		key:         "match:200:user:2",
		value:       `{"score":10.36}`,
		qwatchQuery: "SELECT $key, $value FROM `match:200:user:2` WHERE '$value.score' = 10.36",
		expectedUpdates: [][]interface{}{
			{[]interface{}{"match:200:user:2", map[string]interface{}{"score": 10.36}}},
		},
	},
	{
		key:         "match:200:user:3",
		value:       `{"field1":{"field2":{"field3":{"score":10.36}}}}`,
		qwatchQuery: "SELECT $key, $value FROM `match:200:user:3` WHERE '$value.field1.field2.field3.score' > 10.1",
		expectedUpdates: [][]interface{}{
			{[]interface{}{"match:200:user:3", map[string]interface{}{
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
	publisher, subscribers, cleanup := setupJSONTest(t)
	defer cleanup()

	respParsers := subscribeToJSONQueries(t, subscribers)
	runJSONScenarios(t, publisher, respParsers)
}

func setupJSONTest(t *testing.T) (net.Conn, []net.Conn, func()) {
	publisher := getLocalConnection()
	subscribers := make([]net.Conn, len(JSONTestCases))
	for i := range subscribers {
		subscribers[i] = getLocalConnection()
	}

	cleanup := func() {
		cleanupJSONKeys(publisher)
		if err := publisher.Close(); err != nil {
			t.Errorf("Error closing publisher connection: %v", err)
		}
		for _, sub := range subscribers {
			if err := sub.Close(); err != nil {
				t.Errorf("Error closing subscriber connection: %v", err)
			}
		}
	}

	return publisher, subscribers, cleanup
}

func subscribeToJSONQueries(t *testing.T, subscribers []net.Conn) []*core.RESPParser {
	respParsers := make([]*core.RESPParser, len(subscribers))
	for i, testCase := range JSONTestCases {
		rp := fireCommandAndGetRESPParser(subscribers[i], fmt.Sprintf("QWATCH \"%s\"", testCase.qwatchQuery))
		assert.Assert(t, rp != nil)
		respParsers[i] = rp

		v, err := rp.DecodeOne()
		assert.NilError(t, err)
		assert.Equal(t, 3, len(v.([]interface{})), fmt.Sprintf("Expected 3 elements, got %v", v))
	}
	return respParsers
}

func runJSONScenarios(t *testing.T, publisher net.Conn, respParsers []*core.RESPParser) {
	for i, tc := range JSONTestCases {
		fireCommand(publisher, fmt.Sprintf("JSON.SET %s $ %s", tc.key, tc.value))
		verifyJSONUpdates(t, respParsers[i], tc)
	}
}

func verifyJSONUpdates(t *testing.T, rp *core.RESPParser, tc JSONTestCase) {
	for _, expectedUpdate := range tc.expectedUpdates {
		v, err := rp.DecodeOne()
		assert.NilError(t, err)
		response, ok := v.([]interface{})
		if !ok {
			t.Errorf("Type assertion to []interface{} failed for value: %v", v)
			return
		}
		assert.Equal(t, 3, len(response))
		assert.Equal(t, constants.Qwatch, response[0])

		update, ok := response[2].([]interface{})
		if !ok {
			t.Errorf("Type assertion to []interface{} failed for value: %v", response[2])
			return
		}
		assert.Equal(t, len(expectedUpdate), len(update), fmt.Sprintf("Expected update: %v, got %v", expectedUpdate, update))
		assert.Equal(t, expectedUpdate[0].([]interface{})[0], update[0].([]interface{})[0], "Key mismatch")

		var expectedJSON, actualJSON interface{}
		assert.NilError(t, sonic.UnmarshalString(tc.value, &expectedJSON))
		assert.NilError(t, sonic.UnmarshalString(update[0].([]interface{})[1].(string), &actualJSON))
		assert.DeepEqual(t, expectedJSON, actualJSON)
	}
}

func cleanupKeys(publisher net.Conn) {
	for _, tc := range qWatchTestCases {
		fireCommand(publisher, fmt.Sprintf("DEL match:100:user:%d", tc.userID))
	}
	time.Sleep(100 * time.Millisecond)
}

func cleanupKeysWithSDK(publisher *redis.Client) {
	for _, tc := range qWatchTestCases {
		publisher.Del(context.Background(), fmt.Sprintf("match:100:user:%d", tc.userID))
	}
	time.Sleep(100 * time.Millisecond)
}

func cleanupJSONKeys(publisher net.Conn) {
	for _, tc := range JSONTestCases {
		fireCommand(publisher, fmt.Sprintf("DEL %s", tc.key))
	}
}
