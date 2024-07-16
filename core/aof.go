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
	if _, err := fp.Write(Encode(tokens, false)); err != nil {
		log.Panic(err)
	}
}

// TODO: To to new and switch
func DumpAllAOF() {
	fp, err := os.OpenFile(config.AOFFile, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Print("error", err)
		return
	}
	log.Println("rewriting AOF file at", config.AOFFile)
	for k, obj := range store {
		dumpKey(fp, *((*string)(k)), obj)
	}
	log.Println("AOF file rewrite complete")
}
