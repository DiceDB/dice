package async

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestAPPEND(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	FireCommand(conn, "DEL key listKey bitKey hashKey setKey")
	defer FireCommand(conn, "DEL key listKey bitKey hashKey setKey")

	setErrorMsg := "WRONGTYPE Operation against a key holding the wrong kind of value"
	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name: "APPEND After SET and DEL",
			commands: []string{
				"SET key value",
				"APPEND key value",
				"GET key",
				"APPEND key 100",
				"GET key",
				"DEL key",
				"APPEND key value",
				"GET key",
			},
			expected: []interface{}{"OK", int64(10), "valuevalue", int64(13), "valuevalue100", int64(1), int64(5), "value"},
		},
		{
			name: "APPEND to Integer Values",
			commands: []string{
				"DEL key",
				"APPEND key 1",
				"APPEND key 2",
				"GET key",
				"SET key 1",
				"APPEND key 2",
				"GET key",
			},
			expected: []interface{}{int64(0), int64(1), int64(2), "12", "OK", int64(2), "12"},
		},
		{
			name: "APPEND with leading zeros",
			commands: []string{
				"DEL key",
				"APPEND key 0043",
				"GET key",
				"APPEND key 0034",
				"GET key",
			},
			expected: []interface{}{int64(0), int64(4), "0043", int64(8), "00430034"},
		},
		{
			name: "APPEND with Various Data Types",
			commands: []string{
				"LPUSH listKey lValue",     // Add element to a list
				"SETBIT bitKey 0 1",        // Set a bit in a bitmap
				"HSET hashKey hKey hValue", // Set a field in a hash
				"SADD setKey sValue",       // Add element to a set
				"APPEND listKey value",     // Attempt to append to a list
				"APPEND bitKey value",      // Attempt to append to a bitmap
				"APPEND hashKey value",     // Attempt to append to a hash
				"APPEND setKey value",      // Attempt to append to a set
			},
			expected: []interface{}{int64(1), int64(0), int64(1), int64(1), setErrorMsg, setErrorMsg, setErrorMsg, setErrorMsg},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "DEL key")

			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}

	ttlTolerance := int64(2) // Set tolerance in seconds for the TTL checks
	ttlTestCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name: "APPEND with TTL Set",
			commands: []string{
				"SET key Hello EX 10", // Set a key with a 10-second TTL
				"TTL key",             // Check initial TTL
				"APPEND key World",    // Append a value
				"GET key",             // Get the final value
				"SLEEP 2",             // Sleep for 2 seconds
				"TTL key",             // Check TTL after append.
			},
			expected: []interface{}{"OK", int64(10), int64(10), "HelloWorld", "OK", int64(8), setErrorMsg, setErrorMsg, setErrorMsg, setErrorMsg},
		},
		{
			name: "APPEND before near TTL Expiry",
			commands: []string{
				"SET key Hello EX 3", // Set a key with a 10-second TTL
				"TTL key",            // Check initial TTL
				"APPEND key World",   // Append a value
				"SLEEP 3",            // Sleep for 5 seconds
				"GET key",            // Get the final value which should be (nil)
			},
			expected: []interface{}{"OK", int64(3), int64(10), "OK", "(nil)", int64(-2), setErrorMsg, setErrorMsg, setErrorMsg, setErrorMsg},
		},
	}

	for _, tc := range ttlTestCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "DEL key")

			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				// If checking TTL, apply a tolerance to account for system performance variability
				if cmd == "TTL key" { // Check if TTL command is executed
					expectedTTL := tc.expected[i].(int64)
					actualTTL := result.(int64)

					if actualTTL == -2 { // Key does not exist or is expired
						assert.DeepEqual(t, tc.expected[i], result)
					} else {
						assert.Assert(t, actualTTL >= expectedTTL-ttlTolerance && actualTTL <= expectedTTL+ttlTolerance,
							"TTL %d not within expected range [%d, %d]", actualTTL, expectedTTL-ttlTolerance, expectedTTL+ttlTolerance)
					}
				} else {
					assert.DeepEqual(t, tc.expected[i], result)
				}
			}
		})
	}
}
