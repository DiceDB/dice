package store

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/dicedb/dice/internal/object"

	"github.com/dicedb/dice/config"
)

type AOF struct {
	file   *os.File
	writer *bufio.Writer
	mutex  sync.Mutex
	path   string
}

const (
	FileMode int = 0644
)

func NewAOF(path string) (*AOF, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fs.FileMode(FileMode))
	if err != nil {
		return nil, err
	}

	return &AOF{
		file:   f,
		writer: bufio.NewWriter(f),
		path:   path,
	}, nil
}

func (a *AOF) Write(operation string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if _, err := a.writer.WriteString(operation + "\n"); err != nil {
		return err
	}
	if err := a.writer.Flush(); err != nil {
		return err
	}
	return a.file.Sync()
}

func (a *AOF) Close() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if err := a.writer.Flush(); err != nil {
		return err
	}
	return a.file.Close()
}

func (a *AOF) Load() ([]string, error) {
	f, err := os.Open(a.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var operations []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		operations = append(operations, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return operations, nil
}

func encodeString(v string) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
}

func encode(strs []string) []byte {
	var b []byte
	buf := bytes.NewBuffer(b)
	for _, b := range strs {
		buf.Write(encodeString(b))
	}
	return []byte(fmt.Sprintf("*%d\r\n%s", len(strs), buf.Bytes()))
}

// TODO: Support Expiration
// TODO: Support non-kv data structures
// TODO: Support sync write
func dumpKey(aof *AOF, key string, obj *object.Obj) (err error) {
	cmd := fmt.Sprintf("SET %s %s", key, obj.Value)
	tokens := strings.Split(cmd, " ")
	return aof.Write(string(encode(tokens)))
}

// DumpAllAOF dumps all keys in the store to the AOF file
func DumpAllAOF(store *Store) error {
	var (
		aof *AOF
		err error
	)
	if aof, err = NewAOF(config.DiceConfig.Persistence.AOFFile); err != nil {
		return err
	}
	defer aof.Close()

	log.Println("rewriting AOF file at", config.DiceConfig.Persistence.AOFFile)

	store.store.All(func(k string, obj *object.Obj) bool {
		err = dumpKey(aof, k, obj)
		// continue if no error
		return err == nil
	})

	log.Println("AOF file rewrite complete")
	return err
}
