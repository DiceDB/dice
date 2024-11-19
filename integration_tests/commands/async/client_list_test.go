package async

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var clientFields = []string{
	"id=",
	"addr=",
	"laddr=",
	"fd=",
	"name=",
	"age=",
	"idle=",
	"flags=",
	"multi=",
	"argv-mem=",
	"cmd=",
	"user=",
}

func TestClientList(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	t.Run("Single Connection Active", func(t *testing.T) {
		resp := FireCommand(conn, "CLIENT LIST")
		results := resp.(string)

		// Expecting 2 clients: 1 established during TestMain + this connection
		clientLines := strings.Split(strings.TrimSpace(results), "\n")
		assert.Equal(t, 2, len(clientLines))

		// validate the fields
		validateFields(t, clientLines)
	})

	t.Run("Multiple Connections Active", func(t *testing.T) {
		// Open an additional connection
		secondConn := getLocalConnection()
		defer secondConn.Close()

		resp := FireCommand(conn, "CLIENT LIST")
		results := resp.(string)

		// Expecting 3 clients: 1 from TestMain + 1 initial + 1 new connection
		clientLines := strings.Split(strings.TrimSpace(results), "\n")
		assert.Equal(t, 3, len(clientLines))

		validateFields(t, clientLines)
	})

	t.Run("Client List After Closing Connection", func(t *testing.T) {
		// Open and close a new connection
		tempConn := getLocalConnection()
		tempConn.Close()

		result := FireCommand(conn, "CLIENT LIST")
		results := result.(string)

		// Still expecting 2 clients: TestMain + the initial test connection
		clientLines := strings.Split(strings.TrimSpace(results), "\n")
		assert.Equal(t, 2, len(clientLines))

		validateFields(t, clientLines)
	})
}

func validateFields(t *testing.T, clientLines []string) {
	for _, client := range clientLines {
		for _, field := range clientFields {
			// Check if output contains 'addr=' etc
			assert.Contains(t, client, field)
		}
	}
}
