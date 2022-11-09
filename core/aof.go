package core

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dicedb/dice/config"
)

// TODO: Support Expiration
// TODO: Support non-kv data structures
// TODO: Support sync write
func dumpKey(fp *os.File, key string, obj *Obj) {
	cmd := fmt.Sprintf("SET %s %s", key, obj.Value)
	tokens := strings.Split(cmd, " ")
	fp.Write(Encode(tokens, false))
}

// TODO: To to new and switch
func DumpAllAOF() {
	// If Not present create append only file with rw-r-r permission
	fp, err := os.OpenFile(config.AOFFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Print("error", err)
		return
	}
	defer fp.Close()
	log.Println("rewriting AOF file at", config.AOFFile)
	for k, obj := range store {
		dumpKey(fp, k, obj)
	}
	log.Println("AOF file rewrite complete")
}
