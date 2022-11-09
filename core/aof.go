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
	fp, err := os.OpenFile(config.AOFFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Print("error", err)
		return
	}
	/* Note: A close function also returns an error, a plain defer is harmful. 
	A successful close does not guarantee that the data has been successfully saved 
	to disk,as the kernel uses the buffer cache to defer writes or write calls delays the writing 
	to disk to mitigate cost of frequent writes to disk. A more reliable method is is to use 
	fsync() or f.Sync() and Direct I/O
	*/
	defer func() {
	if err := fp.Close(); err != nil {
		fmt.Print("error", err)
		return
	}
	}()
	log.Println("rewriting AOF file at", config.AOFFile)
	for k, obj := range store {
		dumpKey(fp, k, obj)
	}
	log.Println("AOF file rewrite complete")
}
