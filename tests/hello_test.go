package tests

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestHello(t *testing.T) {
	conn := getLocalConnection()

	expected := []interface{}{
		"proto", int64(2),
		"id", "0.0.0.0:7379",
		"mode", "standalone",
		"role", "master",
		"modules", []interface{}{},
	}

	actual := fireCommand(conn, "HELLO")

	assert.DeepEqual(t, expected, actual)
}
