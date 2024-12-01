package tests

import (
	"log"
	"testing"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/integration_tests/commands/tests/parsers"
	"github.com/dicedb/dice/integration_tests/commands/tests/servers"
	"gotest.tools/v3/assert"
)

func init() {
	parser := config.NewConfigParser()
	if err := parser.ParseDefaults(config.DiceConfig); err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}
}

func TestRespCommands(t *testing.T) {
	conn := servers.GetRespConn()
	allTests := GetAllTests()

	for _, test := range allTests {
		t.Run(test.Name, func(t *testing.T) {
			for idx, cmd := range test.Input {
				output := parsers.RespCommandExecuter(conn, cmd)
				assert.Equal(t, test.Output[idx], output)
			}
		})

		for _, key := range test.Cleanup {
			cmd := "DEL " + key
			_ = parsers.RespCommandExecuter(conn, cmd)
		}
	}
}
