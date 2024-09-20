package server

import (
	"bytes"
	"fmt"

	"github.com/dicedb/dice/internal/clientio"
	diceerrors "github.com/dicedb/dice/internal/errors"
	dstore "github.com/dicedb/dice/internal/store"
)

// evalINFO creates a buffer with the info of total keys per db
// Returns the encoded buffer as response
func respINFO(args []string) []byte {
	var info []byte
	buf := bytes.NewBuffer(info)
	buf.WriteString("# Keyspace\r\n")
	for i := range dstore.KeyspaceStat {
		fmt.Fprintf(buf, "db%d:keys=%d,expires=0,avg_ttl=0\r\n", i, dstore.KeyspaceStat[i]["keys"])
	}
	return clientio.Encode(buf.String(), false)
}

// evaluate PING command
func respPING(args []string) []byte {
	if len(args) >= 2 {
		return diceerrors.NewErrArity("PING")
	}

	if len(args) == 0 {
		return clientio.Encode("PONG", true)
	}
	return clientio.Encode(args[0], false)
}
