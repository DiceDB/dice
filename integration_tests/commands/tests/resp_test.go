package tests

import (
	"log"
	"testing"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/integration_tests/commands/tests/parsers"
	"github.com/dicedb/dice/integration_tests/commands/tests/servers"
	"github.com/stretchr/testify/assert"
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
			if !Validate(&test) {
				t.Fatal("Test progression failed...")
			}

			if len(test.Setup) > 0 {
				for _, setup := range test.Setup {
					for idx, cmd := range setup.Input {
						output := parsers.RespCommandExecuter(conn, cmd)
						assert.Equal(t, setup.Output[idx], output)
					}
				}
			}

			for idx, cmd := range test.Input {
				if len(test.Delays) > 0 {
					time.Sleep(test.Delays[idx])
				}

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
