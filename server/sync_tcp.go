package server

import (
	"io"
	"strings"

	"github.com/dicedb/dice/core/cmd"

	"github.com/dicedb/dice/core"
)

func toArrayString(ai []interface{}) []string {
	as := make([]string, len(ai))
	for i := range ai {
		as[i] = ai[i].(string)
	}
	return as
}

func readCommands(c io.ReadWriter) (cmd.RedisCmds, bool, error) {
	var hasABORT bool = false
	rp := core.NewRESPParser(c)
	values, err := rp.DecodeMultiple()
	if err != nil {
		return nil, false, err
	}

	var cmds []*cmd.RedisCmd = make([]*cmd.RedisCmd, 0)
	for _, value := range values {
		tokens := toArrayString(value.([]interface{}))
		command := strings.ToUpper(tokens[0])
		cmds = append(cmds, &cmd.RedisCmd{
			ID:   core.NextID(),
			Cmd:  command,
			Args: tokens[1:],
		})

		if command == "ABORT" {
			hasABORT = true
		}
	}
	return cmds, hasABORT, err
}
