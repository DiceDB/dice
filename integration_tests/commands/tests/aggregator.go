package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var allTests []Meta

// SetupCmd struct is used to define a setup command
// Input: Input commands to be executed
// Output: Expected output of the setup command
// Keep the setup tests simple which return "OK" or "integer" or "(nil)"
// For complex setup tests, use the test cases
type SetupCmd struct {
	Input  []string
	Output []interface{}
}

const (
	EQUAL = "EQUAL"
	JSON  = "JSON"
	ARRAY = "ARRAY"
)

// Meta struct is used to define a test case
// Name: Name of the test case
// Cmd: Command to be executed
// Setup: Setup commands to be executed
// Input: Input commands to be executed
// Output: Expected output of the test case
// Delays: Delays to be introduced between commands
// CleanupKeys: list of keys to be cleaned up after the test case
type Meta struct {
	Name    string
	Cmd     string
	Setup   []SetupCmd
	Input   []string
	Output  []interface{}
	Assert  []string
	Delays  []time.Duration
	Cleanup []string
}

// RegisterTests appends a Meta slice to the global test list
func RegisterTests(tests []Meta) {
	allTests = append(allTests, tests...)
}

// GetAllTests returns all registered test cases
func GetAllTests() []Meta {
	return allTests
}

func SwitchAsserts(t *testing.T, kind string, expected interface{}, actual interface{}) {
	switch kind {
	case EQUAL:
		assert.Equal(t, expected, actual)
	case JSON:
		assert.JSONEq(t, expected.(string), actual.(string))
	case ARRAY:
		assert.ElementsMatch(t, expected, actual)
	}
}
