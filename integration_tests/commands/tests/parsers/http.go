package parsers

import (
	"strings"

	"github.com/dicedb/dice/integration_tests/commands/tests/servers"
)

func ParseResponse(response interface{}) interface{} {
	// convert the output to the int64 if it is float64
	switch response.(type) {
	case float64:
		return int64(response.(float64))
	case nil:
		return "(nil)"
	default:
		return response
	}
}

func HttpCommandExecuter(exec *servers.HTTPCommandExecutor, cmd string) (interface{}, error) {
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
	diceHttpCmd := servers.HTTPCommand{
		Command: strings.ToLower(command),
		Body:    body,
	}
	res, err := exec.FireCommand(diceHttpCmd)
	if err != nil {
		return nil, err
	}
	return ParseResponse(res), nil
}
