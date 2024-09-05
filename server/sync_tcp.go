package server

import (
	"fmt"
	"io"
	"strings"

	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/core/cmd"
)

func toArrayString(ai []interface{}) ([]string, error) {
	as := make([]string, len(ai))
	for i, v := range ai {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("element at index %d is not a string", i)
		}
		as[i] = s
	}
	return as, nil
}

func readCommands(c io.ReadWriter) (cmd.RedisCmds, bool, error) {
	var hasABORT = false
	rp := core.NewRESPParser(c)
	values, err := rp.DecodeMultiple()
	if err != nil {
		return nil, false, err
	}

	var cmds = make([]*cmd.RedisCmd, 0)
	for _, value := range values {
		arrayValue, ok := value.([]interface{})
		if !ok {
			return nil, false, fmt.Errorf("expected array, got %T", value)
		}

		tokens, err := toArrayString(arrayValue)
		if err != nil {
			return nil, false, err
		}

		if len(tokens) == 0 {
			return nil, false, fmt.Errorf("empty command")
		}

		command := strings.ToUpper(tokens[0])
		cmds = append(cmds, &cmd.RedisCmd{
			Cmd:  command,
			Args: tokens[1:],
		})

		if command == "ABORT" {
			hasABORT = true
		}
	}
	return cmds, hasABORT, nil
}
