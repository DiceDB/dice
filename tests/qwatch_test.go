package tests

import (
	"context"
	"fmt"
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
