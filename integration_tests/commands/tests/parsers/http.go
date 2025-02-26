package parsers

import (
	"strings"

	"github.com/dicedb/dice/integration_tests/commands/tests/servers"
)

func HTTPCommandExecuter(exec *servers.HTTPCommandExecutor, cmd string) (interface{}, error) {
	// convert the command to a HTTPCommand
	// cmd starts with Command and Body is values after that
	tokens := strings.Split(cmd, " ")
	command := tokens[0]
	body := make(map[string]interface{})
	if len(tokens) > 1 {
		// convert the tokens []string to []interface{}
		values := make([]interface{}, len(tokens[1:]))
		for i, v := range tokens[1:] {
			values[i] = v
		}
		body["values"] = values
	} else {
		body["values"] = []interface{}{}
	}
	diceHTTPCmd := servers.HTTPCommand{
		Command: strings.ToLower(command),
		Body:    body,
	}
	res, err := exec.FireCommand(diceHTTPCmd)
	if err != nil {
		return nil, err
	}
	return ParseResponse(res), nil
}
