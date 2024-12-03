package tests

import (
	"fmt"
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

func Validate(test *Meta) bool {
	// Validate test structure
	if len(test.Input) != len(test.Output) {
		fmt.Printf("Test %s: mismatch between number of inputs (%d) and outputs (%d)", test.Name, len(test.Input), len(test.Output))
		return false
	}
	if len(test.Delays) > 0 && len(test.Delays) != len(test.Input) {
		fmt.Printf("Test %s: mismatch between number of inputs (%d) and delays (%d)", test.Name, len(test.Input), len(test.Delays))
		return false
	}
	if len(test.Setup) > 0 {
		for _, setup := range test.Setup {
			if len(setup.Input) != len(setup.Output) {
				fmt.Printf("Test %s (Setup): mismatch between number of setup inputs (%d) and outputs (%d)", test.Name, len(setup.Input), len(setup.Output))
				return false
			}
		}
	}

	return true
}
