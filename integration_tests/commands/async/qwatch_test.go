package async

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/sql"
	dicedb "github.com/dicedb/dicedb-go"
	"github.com/stretchr/testify/assert"
)

type qWatchTestCase struct {
	key             string
	userID          int
	score           int
	expectedUpdates [][]interface{}
}

type qWatchSDKSubscriber struct {
	client *dicedb.Client
	qwatch *dicedb.QWatch
}

var qWatchQuery = "SELECT $key, $value WHERE $key like 'match:10?:*' ORDER BY $value desc LIMIT 3"

var qWatchTestCases = []qWatchTestCase{
	{"match:100:user", 0, 11, [][]interface{}{
		{[]interface{}{"match:100:user:0", int64(11)}},
	}},
	{"match:100:user", 1, 33, [][]interface{}{
		{[]interface{}{"match:100:user:1", int64(33)}, []interface{}{"match:100:user:0", int64(11)}},
	}},
	{"match:101:user", 2, 22, [][]interface{}{
		{[]interface{}{"match:100:user:1", int64(33)}, []interface{}{"match:101:user:2", int64(22)}, []interface{}{"match:100:user:0", int64(11)}},
	}},
	{"match:102:user", 3, 0, [][]interface{}{
		{[]interface{}{"match:100:user:1", int64(33)}, []interface{}{"match:101:user:2", int64(22)}, []interface{}{"match:100:user:0", int64(11)}},
	}},
	{"match:100:user", 4, 44, [][]interface{}{
		{[]interface{}{"match:100:user:4", int64(44)}, []interface{}{"match:100:user:1", int64(33)}, []interface{}{"match:101:user:2", int64(22)}},
	}},
	{"match:100:user", 5, 50, [][]interface{}{
		{[]interface{}{"match:100:user:5", int64(50)}, []interface{}{"match:100:user:4", int64(44)}, []interface{}{"match:100:user:1", int64(33)}},
	}},
	{"match:101:user", 2, 40, [][]interface{}{
		{[]interface{}{"match:100:user:5", int64(50)}, []interface{}{"match:100:user:4", int64(44)}, []interface{}{"match:101:user:2", int64(40)}},
	}},
	{"match:100:user", 6, 55, [][]interface{}{
		{[]interface{}{"match:100:user:6", int64(55)}, []interface{}{"match:100:user:5", int64(50)}, []interface{}{"match:100:user:4", int64(44)}},
	}},
	{"match:100:user", 0, 60, [][]interface{}{
		{[]interface{}{"match:100:user:0", int64(60)}, []interface{}{"match:100:user:6", int64(55)}, []interface{}{"match:100:user:5", int64(50)}},
	}},
	{"match:100:user", 5, 70, [][]interface{}{
		{[]interface{}{"match:100:user:5", int64(70)}, []interface{}{"match:100:user:0", int64(60)}, []interface{}{"match:100:user:6", int64(55)}},
	}},
}

// TestQWATCH tests the QWATCH functionality using raw network connections.
func TestQWATCH(t *testing.T) {
	publisher, subscribers, cleanup := setupQWATCHTest(t, qWatchQuery)
	defer cleanup()

	respParsers := subscribeToQWATCH(t, subscribers, qWatchQuery)
	runQWatchScenarios(t, publisher, respParsers, qWatchQuery, qWatchTestCases)
}

func TestQWATCHWithSDK(t *testing.T) {
	publisher, subscribers, cleanup := setupQWATCHTestWithSDK(t)
	defer cleanup()

	channels := subscribeToQWATCHWithSDK(t, subscribers)
	runQWatchScenarios(t, publisher, channels, qWatchQuery, qWatchTestCases)
}

func setupQWATCHTest(t *testing.T, query string) (net.Conn, []net.Conn, func()) {
	t.Helper()
	publisher := getLocalConnection()
	subscribers := []net.Conn{getLocalConnection(), getLocalConnection(), getLocalConnection()}

	cleanup := func() {
		cleanupQWATCHKeys(publisher)
		if err := publisher.Close(); err != nil {
			t.Errorf("Error closing publisher connection: %v", err)
		}
		for _, sub := range subscribers {
			FireCommand(sub, fmt.Sprintf("Q.UNWATCH \"%s\"", query))
			time.Sleep(100 * time.Millisecond)
			if err := sub.Close(); err != nil {
				t.Errorf("Error closing subscriber connection: %v", err)
			}
		}
	}

	return publisher, subscribers, cleanup
}

func cleanupQWATCHKeys(publisher net.Conn) {
	for _, tc := range qWatchTestCases {
		FireCommand(publisher, fmt.Sprintf("DEL %s:%d", tc.key, tc.userID))
	}
	time.Sleep(100 * time.Millisecond)
}

func setupQWATCHTestWithSDK(t *testing.T) (*dicedb.Client, []qWatchSDKSubscriber, func()) {
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

func cleanupKeysWithSDK(publisher *dicedb.Client) {
	for _, tc := range qWatchTestCases {
		publisher.Del(context.Background(), fmt.Sprintf("%s:%d", tc.key, tc.userID))
	}
	time.Sleep(100 * time.Millisecond)
}

func subscribeToQWATCH(t *testing.T, subscribers []net.Conn, query string) []*clientio.RESPParser {
	t.Helper()
	respParsers := make([]*clientio.RESPParser, len(subscribers))
	for i, subscriber := range subscribers {
		rp := fireCommandAndGetRESPParser(subscriber, fmt.Sprintf("Q.WATCH \"%s\"", query))
		assert.True(t, rp != nil)
		respParsers[i] = rp

		v, err := rp.DecodeOne()
		assert.Nil(t, err)
		castedValue, ok := v.([]interface{})
		if !ok {
			t.Errorf("Type assertion to []interface{} failed for value: %v", v)
			return nil
		}
		assert.Equal(t, 3, len(castedValue))
	}
	return respParsers
}

func subscribeToQWATCHWithSDK(t *testing.T, subscribers []qWatchSDKSubscriber) []<-chan *dicedb.QMessage {
	t.Helper()
	ctx := context.Background()
	channels := make([]<-chan *dicedb.QMessage, len(subscribers))
	for i, subscriber := range subscribers {
		qwatch := subscriber.client.QWatch(ctx)
		subscribers[i].qwatch = qwatch
		assert.True(t, qwatch != nil)
		err := qwatch.WatchQuery(ctx, qWatchQuery)
		assert.Nil(t, err)
		channels[i] = qwatch.Channel()
		<-channels[i] // Get the first message
	}
	return channels
}

func runQWatchScenarios(t *testing.T, publisher interface{}, receivers interface{}, query string, tests []qWatchTestCase) {
	t.Helper()
	for _, tc := range tests {
		publishUpdate(t, publisher, tc)
		verifyUpdates(t, receivers, tc.expectedUpdates, query)
	}
}

func publishUpdate(t *testing.T, publisher interface{}, tc qWatchTestCase) {
	key := fmt.Sprintf("%s:%d", tc.key, tc.userID)
	switch p := publisher.(type) {
	case net.Conn:
		FireCommand(p, fmt.Sprintf("SET %s %d", key, tc.score))
	case *dicedb.Client:
		err := p.Set(context.Background(), key, tc.score, 0).Err()
		assert.Nil(t, err)
	}
}

func verifyUpdates(t *testing.T, receivers interface{}, expectedUpdates [][]interface{}, query string) {
	for _, expectedUpdate := range expectedUpdates {
		switch r := receivers.(type) {
		case []*clientio.RESPParser:
			verifyRESPUpdates(t, r, expectedUpdate, query)
		case []<-chan *dicedb.QMessage:
			verifySDKUpdates(t, r, expectedUpdate)
		}
	}
}

func verifyRESPUpdates(t *testing.T, respParsers []*clientio.RESPParser, expectedUpdate []interface{}, query string) {
	for _, rp := range respParsers {
		v, err := rp.DecodeOne()
		assert.Nil(t, err)
		update, ok := v.([]interface{})
		if !ok {
			t.Errorf("Type assertion to []interface{} failed for value: %v", v)
			return
		}
		assert.Equal(t, []interface{}{sql.Qwatch, query, expectedUpdate}, update)
	}
}

func verifySDKUpdates(t *testing.T, channels []<-chan *dicedb.QMessage, expectedUpdate []interface{}) {
	for _, ch := range channels {
		v := <-ch
		assert.Equal(t, len(v.Updates), len(expectedUpdate), v.Updates)
		for i, update := range v.Updates {
			assert.Equal(t, expectedUpdate[i], []interface{}{update.Key, update.Value})
		}
	}
}

// Test cases for WHERE clause for Regular keys

var qWatchWhereQuery = "SELECT $key, $value WHERE $value > 50 and $key like 'match:10?:*' ORDER BY $value desc"

var qWatchWhereTestCases = []qWatchTestCase{
	{"match:100:user", 0, 55, [][]interface{}{
		{[]interface{}{"match:100:user:0", int64(55)}},
	}},
	{"match:100:user", 1, 60, [][]interface{}{
		{[]interface{}{"match:100:user:1", int64(60)}, []interface{}{"match:100:user:0", int64(55)}},
	}},
	{"match:100:user", 2, 80, [][]interface{}{
		{[]interface{}{"match:100:user:2", int64(80)}, []interface{}{"match:100:user:1", int64(60)}, []interface{}{"match:100:user:0", int64(55)}},
	}},
	{"match:100:user", 0, 90, [][]interface{}{
		{[]interface{}{"match:100:user:0", int64(90)}, []interface{}{"match:100:user:2", int64(80)}, []interface{}{"match:100:user:1", int64(60)}},
	}},
}

func TestQWatchWhere(t *testing.T) {
	publisher, subscribers, cleanup := setupQWATCHTest(t, qWatchWhereQuery)
	defer cleanup()

	respParsers := subscribeToQWATCH(t, subscribers, qWatchWhereQuery)
	runQWatchScenarios(t, publisher, respParsers, qWatchWhereQuery, qWatchWhereTestCases)
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
		qwatchQuery: "SELECT $key, $value WHERE $key like 'match:200:user:0' AND '$value.name' = 'Tom'",
		expectedUpdates: [][]interface{}{
			{[]interface{}{"match:200:user:0", map[string]interface{}{"name": "Tom"}}},
		},
	},
	{
		key:         "match:200:user:1",
		value:       `{"name":"Tom","age":24}`,
		qwatchQuery: "SELECT $key, $value WHERE $key like 'match:200:user:1' AND '$value.age' > 20",
		expectedUpdates: [][]interface{}{
			{[]interface{}{"match:200:user:1", map[string]interface{}{"name": "Tom", "age": float64(24)}}},
		},
	},
	{
		key:         "match:200:user:2",
		value:       `{"score":10.36}`,
		qwatchQuery: "SELECT $key, $value WHERE $key like 'match:200:user:2' AND '$value.score' = 10.36",
		expectedUpdates: [][]interface{}{
			{[]interface{}{"match:200:user:2", map[string]interface{}{"score": 10.36}}},
		},
	},
	{
		key:         "match:200:user:3",
		value:       `{"field1":{"field2":{"field3":{"score":10.36}}}}`,
		qwatchQuery: "SELECT $key, $value WHERE $key like 'match:200:user:3' AND '$value.field1.field2.field3.score' > 10.1",
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
	publisher, subscribers, cleanup := setupJSONTest(t, JSONTestCases)
	defer cleanup()

	respParsers := subscribeToJSONQueries(t, subscribers, JSONTestCases)
	runJSONScenarios(t, publisher, respParsers, JSONTestCases)
}

func setupJSONTest(t *testing.T, tests []JSONTestCase) (net.Conn, []net.Conn, func()) {
	publisher := getLocalConnection()
	subscribers := make([]net.Conn, len(tests))
	for i := range subscribers {
		subscribers[i] = getLocalConnection()
	}

	cleanup := func() {
		cleanupJSONKeys(publisher, tests)
		if err := publisher.Close(); err != nil {
			t.Errorf("Error closing publisher connection: %v", err)
		}
		for i, sub := range subscribers {
			FireCommand(sub, fmt.Sprintf("Q.UNWATCH \"%s\"", tests[i].qwatchQuery))
			if err := sub.Close(); err != nil {
				t.Errorf("Error closing subscriber connection: %v", err)
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	return publisher, subscribers, cleanup
}

func subscribeToJSONQueries(t *testing.T, subscribers []net.Conn, tests []JSONTestCase) []*clientio.RESPParser {
	respParsers := make([]*clientio.RESPParser, len(subscribers))
	for i, testCase := range tests {
		rp := fireCommandAndGetRESPParser(subscribers[i], fmt.Sprintf("Q.WATCH \"%s\"", testCase.qwatchQuery))
		assert.True(t, rp != nil)
		respParsers[i] = rp

		v, err := rp.DecodeOne()
		assert.Nil(t, err)
		assert.Equal(t, 3, len(v.([]interface{})), fmt.Sprintf("Expected 3 elements, got %v", v))
	}
	return respParsers
}

func runJSONScenarios(t *testing.T, publisher net.Conn, respParsers []*clientio.RESPParser, tests []JSONTestCase) {
	for i, tc := range tests {
		FireCommand(publisher, fmt.Sprintf("JSON.SET %s $ %s", tc.key, tc.value))
		verifyJSONUpdates(t, respParsers[i], tc)
	}
}

func verifyJSONUpdates(t *testing.T, rp *clientio.RESPParser, tc JSONTestCase) {
	for _, expectedUpdate := range tc.expectedUpdates {
		v, err := rp.DecodeOne()
		assert.Nil(t, err)
		response, ok := v.([]interface{})
		if !ok {
			t.Errorf("Type assertion to []interface{} failed for value: %v", v)
			return
		}
		assert.Equal(t, 3, len(response))
		assert.Equal(t, sql.Qwatch, response[0])

		update, ok := response[2].([]interface{})
		if !ok {
			t.Errorf("Type assertion to []interface{} failed for value: %v", response[2])
			return
		}
		assert.Equal(t, len(expectedUpdate), len(update), fmt.Sprintf("Expected update: %v, got %v", expectedUpdate, update))
		assert.Equal(t, expectedUpdate[0].([]interface{})[0], update[0].([]interface{})[0], "Key mismatch")

		var expectedJSON, actualJSON interface{}
		assert.Nil(t, sonic.UnmarshalString(tc.value, &expectedJSON))
		assert.Nil(t, sonic.UnmarshalString(update[0].([]interface{})[1].(string), &actualJSON))
		assert.Equal(t, expectedJSON, actualJSON)
	}
}

func cleanupJSONKeys(publisher net.Conn, tests []JSONTestCase) {
	for _, tc := range tests {
		FireCommand(publisher, fmt.Sprintf("DEL %s", tc.key))
	}
}
func TestQwatchWithJSONOrderBy(t *testing.T) {
	publisher, subscriber, cleanup, watchquery := setupJSONOrderByTest(t)
	defer cleanup()

	respParser := subscribeToJSONOrderByQuery(t, subscriber, watchquery)
	runJSONOrderByScenarios(t, publisher, respParser)
}

func setupJSONOrderByTest(t *testing.T) (net.Conn, net.Conn, func(), string) {
	watchquery := "SELECT $key, $value WHERE $key like 'player:*' ORDER BY $value.score DESC LIMIT 3"
	publisher := getLocalConnection()
	subscriber := getLocalConnection()

	cleanup := func() {
		cleanupJSONOrderByKeys(publisher)
		if err := publisher.Close(); err != nil {
			t.Errorf("Error closing publisher connection: %v", err)
		}
		FireCommand(subscriber, fmt.Sprintf("Q.UNWATCH \"%s\"", watchquery))
		time.Sleep(100 * time.Millisecond)
		if err := subscriber.Close(); err != nil {
			t.Errorf("Error closing subscriber connection: %v", err)
		}
	}

	return publisher, subscriber, cleanup, watchquery
}

func subscribeToJSONOrderByQuery(t *testing.T, subscriber net.Conn, watchquery string) *clientio.RESPParser {
	rp := fireCommandAndGetRESPParser(subscriber, fmt.Sprintf("Q.WATCH \"%s\"", watchquery))
	assert.True(t, rp != nil)

	v, err := rp.DecodeOne()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(v.([]interface{})), fmt.Sprintf("Expected 3 elements, got %v", v))

	return rp
}

func runJSONOrderByScenarios(t *testing.T, publisher net.Conn, respParser *clientio.RESPParser) {
	scenarios := []struct {
		key             string
		value           string
		expectedUpdates [][]interface{}
	}{
		{
			key:   "player:1",
			value: `{"name":"Alice","score":100}`,
			expectedUpdates: [][]interface{}{
				{[]interface{}{"player:1", map[string]interface{}{"name": "Alice", "score": float64(100)}}},
			},
		},
		{
			key:   "player:2",
			value: `{"name":"Bob","score":80}`,
			expectedUpdates: [][]interface{}{
				{[]interface{}{"player:1", map[string]interface{}{"name": "Alice", "score": float64(100)}},
					[]interface{}{"player:2", map[string]interface{}{"name": "Bob", "score": float64(80)}}},
			},
		},
		{
			key:   "player:3",
			value: `{"name":"Charlie","score":90}`,
			expectedUpdates: [][]interface{}{
				{[]interface{}{"player:1", map[string]interface{}{"name": "Alice", "score": float64(100)}},
					[]interface{}{"player:3", map[string]interface{}{"name": "Charlie", "score": float64(90)}},
					[]interface{}{"player:2", map[string]interface{}{"name": "Bob", "score": float64(80)}}},
			},
		},
		{
			key:   "player:4",
			value: `{"name":"David","score":95}`,
			expectedUpdates: [][]interface{}{
				{[]interface{}{"player:1", map[string]interface{}{"name": "Alice", "score": float64(100)}},
					[]interface{}{"player:4", map[string]interface{}{"name": "David", "score": float64(95)}},
					[]interface{}{"player:3", map[string]interface{}{"name": "Charlie", "score": float64(90)}}},
			},
		},
	}

	for _, sc := range scenarios {
		FireCommand(publisher, fmt.Sprintf("JSON.SET %s $ %s", sc.key, sc.value))
		verifyJSONOrderByUpdates(t, respParser, sc)
	}
}

func verifyJSONOrderByUpdates(t *testing.T, rp *clientio.RESPParser, tc struct {
	key             string
	value           string
	expectedUpdates [][]interface{}
}) {
	expectedUpdates := tc.expectedUpdates[0]

	// Decode the response
	v, err := rp.DecodeOne()
	assert.Nil(t, err, "Failed to decode response")

	// Cast the response to []interface{}
	response, ok := v.([]interface{})
	assert.True(t, ok, "Response is not of type []interface{}: %v", v)

	// Verify response structure
	assert.Equal(t, 3, len(response), "Expected response to have 3 elements")
	assert.Equal(t, sql.Qwatch, response[0], "First element should be Qwatch constant")

	// Extract updates from the response
	updates, ok := response[2].([]interface{})
	assert.True(t, ok, "Updates are not of type []interface{}: %v", response[2])

	// Verify number of updates
	assert.Equal(t, len(expectedUpdates), len(updates),
		"Number of updates mismatch. Expected: %d, Got: %d", len(expectedUpdates), len(updates))

	// Verify each update
	for i, expectedRow := range expectedUpdates {
		actualRow, ok := updates[i].([]interface{})
		assert.True(t, ok, "Update row is not of type []interface{}: %v", updates[i])

		// Verify key
		assert.Equal(t, expectedRow.([]interface{})[0], actualRow[0],
			"Key mismatch at index %d", i)

		// Verify JSON value
		var actualJSON interface{}
		err := sonic.UnmarshalString(actualRow[1].(string), &actualJSON)
		assert.Nil(t, err, "Failed to unmarshal JSON at index %d", i)

		assert.Equal(t, expectedRow.([]interface{})[1], actualJSON)
	}
}

func cleanupJSONOrderByKeys(publisher net.Conn) {
	for i := 1; i <= 4; i++ {
		FireCommand(publisher, fmt.Sprintf("DEL player:%d", i))
	}
}

// Test cases for WHERE clause for JSON keys

var whereJSONTestCases = []JSONTestCase{
	{
		key:         "match:200:user:0",
		value:       `{"name":"Tom"}`,
		qwatchQuery: "SELECT $key, $value WHERE '$value.name' = 'Tom' AND $key like 'match:200:user:0'",
		expectedUpdates: [][]interface{}{
			{[]interface{}{"match:200:user:0", map[string]interface{}{"name": "Tom"}}},
		},
	},
	{
		key:         "match:200:user:1",
		value:       `{"name":"Tom","age":24}`,
		qwatchQuery: "SELECT $key, $value WHERE '$value.age' > 20 AND $key like 'match:200:user:1'",
		expectedUpdates: [][]interface{}{
			{[]interface{}{"match:200:user:1", map[string]interface{}{"name": "Tom", "age": float64(24)}}},
		},
	},
	{
		key:         "match:200:user:2",
		value:       `{"score":10.36}`,
		qwatchQuery: "SELECT $key, $value WHERE '$value.score' = 10.36 AND $key like 'match:200:user:2'",
		expectedUpdates: [][]interface{}{
			{[]interface{}{"match:200:user:2", map[string]interface{}{"score": 10.36}}},
		},
	},
	{
		key:         "match:200:user:3",
		value:       `{"field1":{"field2":{"field3":{"score":10.36}}}}`,
		qwatchQuery: "SELECT $key, $value WHERE '$value.field1.field2.field3.score' > 10.1 AND $key like 'match:200:user:3'",
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

func TestQwatchWhereWithJSON(t *testing.T) {
	tests := whereJSONTestCases
	publisher, subscribers, cleanup := setupJSONTest(t, tests)
	defer cleanup()

	respParsers := subscribeToJSONQueries(t, subscribers, tests)
	runJSONScenarios(t, publisher, respParsers, tests)
}
