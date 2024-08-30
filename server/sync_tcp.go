package server

import (
	"io"
	"strings"

	"github.com/dicedb/dice/core"
)

func toArrayString(ai []interface{}) []string {
	as := make([]string, len(ai))
	for i := range ai {
		as[i] = ai[i].(string)
	}
	return as
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
		tokens := toArrayString(value.([]interface{}))
		cmd := strings.ToUpper(tokens[0])
		cmds = append(cmds, &core.RedisCmd{
			ID:   core.NextID(),
			Cmd:  cmd,
			Args: tokens[1:],
		})

		if cmd == "ABORT" {
			hasABORT = true
		}
	}
	return cmds, hasABORT, err
}

func respond(cmds core.RedisCmds, c *core.Client, store *core.Store) {
	core.EvalAndRespond(cmds, c, store)
}
