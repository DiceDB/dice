package server

import (
	"io"
	"strings"

	"github.com/dicedb/dice/core"
)

func toArrayString(ai []interface{}) ([]string, error) {
	as := make([]string, len(ai))
	for i := range ai {
		as[i] = ai[i].(string)
	}
	return as, nil
}

func readCommands(c io.ReadWriter) (core.RedisCmds, bool, error) {
	var hasABORT bool = false
	rp := core.NewRESPParser(c)
	values, err := rp.DecodeMultiple()
	if err != nil {
		return nil, false, err
	}

	var cmds []*core.RedisCmd = make([]*core.RedisCmd, 0)
	for _, value := range values {
		tokens, err := toArrayString(value.([]interface{}))
		if err != nil {
			return nil, false, err
		}
		cmd := strings.ToUpper(tokens[0])
		cmds = append(cmds, &core.RedisCmd{
			Cmd:  cmd,
			Args: tokens[1:],
		})

		if cmd == "ABORT" {
			hasABORT = true
		}
	}
	return cmds, hasABORT, err
}

func respond(cmds core.RedisCmds, c *core.Client) {
	core.EvalAndRespond(cmds, c)
}
