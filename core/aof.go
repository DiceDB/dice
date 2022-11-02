package core

import (
	"fmt"
	"log"
	"os"
	"strings"

	dbEngine "github.com/dicedb/dice/IStorageEngines"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/object"
)

// TODO: Support Expiration
// TODO: Support non-kv data structures
// TODO: Support sync write
func dumpKey(fp *os.File, key string, obj *object.Obj) {
	cmd := fmt.Sprintf("SET %s %s", key, obj.Value)
	tokens := strings.Split(cmd, " ")
	fp.Write(Encode(tokens, false))
}

// TODO: To to new and switch
func DumpAllAOF(dh dbEngine.IKVStorage) {
	fp, err := os.OpenFile(config.AOFFile, os.O_CREATE|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Print("error", err)
		return
	}
	log.Println("rewriting AOF file at", config.AOFFile)
	dh.GetStorage().Range(func(k, v interface{}) bool {
		dumpKey(fp, k.(string), v.(*object.Obj))
		return true
	})
	log.Println("AOF file rewrite complete")
}
