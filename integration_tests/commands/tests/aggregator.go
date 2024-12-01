package tests

import "time"

var allTests []Meta

type Meta struct {
	Name    string
	Cmd     string
	Input   []string
	Output  []interface{}
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
