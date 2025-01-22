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

func TestHttpCommands(t *testing.T) {
	exec := servers.NewHTTPCommandExecutor()
	allTests := GetAllTests()

	for _, test := range allTests {
		t.Run(test.Name, func(t *testing.T) {
			if !Validate(&test) {
				t.Fatal("Test progression failed...")
			}

			// Setup commands
			if len(test.Setup) > 0 {
				for _, setup := range test.Setup {
					for idx, cmd := range setup.Input {
						output, _ := parsers.HttpCommandExecuter(exec, cmd)
						assert.Equal(t, setup.Output[idx], output)
					}
				}
			}

			for idx, cmd := range test.Input {
				if len(test.Delays) > 0 {
					time.Sleep(test.Delays[idx])
				}
				output, _ := parsers.HttpCommandExecuter(exec, cmd)
				if len(test.Assert) > 0 {
					SwitchAsserts(t, test.Assert[idx], test.Output[idx], output)
				} else {
					assert.Equal(t, test.Output[idx], output)
				}
			}
			if len(test.Cleanup) > 0 {
				// join all the keys to be cleaned up
				keys := ""
				for _, key := range test.Cleanup {
					keys += key + " "
				}
				parsers.HttpCommandExecuter(exec, `DEL `+keys)
			}
		})

	}
}
