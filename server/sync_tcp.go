package server

import (
	"io"
	"log"
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
	// TODO: Max read in one shot is 512 bytes
	// To allow input > 512 bytes, then repeated read until
	// we get EOF or designated delimiter
	var buf []byte = make([]byte, 512)
	n, err := c.Read(buf[:])
	if err != nil {
		return nil, false, err
	}

	values, maliciousFlag, err := core.Decode(buf[:n])
	if err != nil {
		return nil, maliciousFlag, err
	}
	if maliciousFlag {
		log.Println("Possible SECURITY ATTACK detected. This is likely due to an attacker attempting to use Cross Protocol Scripting to compromise your instance. Connection aborted.")
		return nil, true, nil
	}

	var cmds []*core.RedisCmd = make([]*core.RedisCmd, 0)
	for _, value := range values {
		tokens, err := toArrayString(value.([]interface{}))
		if err != nil {
			return nil, false, err
		}
		cmds = append(cmds, &core.RedisCmd{
			Cmd:  strings.ToUpper(tokens[0]),
			Args: tokens[1:],
		})
	}
	return cmds, false, err
}

func respond(cmds core.RedisCmds, c *core.Client) {
	core.EvalAndRespond(cmds, c)
}
