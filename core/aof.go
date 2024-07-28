package core

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/dicedb/dice/config"
)

type AOF struct {
	file   *os.File
	writer *bufio.Writer
	mutex  sync.Mutex
	path   string
}

func NewAOF(path string) (*AOF, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
	if err := a.file.Sync(); err != nil {
		return err
	}
	return nil
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

// TODO: Support Expiration
// TODO: Support non-kv data structures
// TODO: Support sync write
func dumpKey(aof *AOF, key string, obj *Obj) (err error) {
	cmd := fmt.Sprintf("SET %s %s", key, obj.Value)
	tokens := strings.Split(cmd, " ")
	if err = aof.Write(string(Encode(tokens, false))); err != nil {
		return err
	}
	return nil
}

// TODO: To to new and switch
func DumpAllAOF() error {
	var (
		aof *AOF
		err error
	)
	if aof, err = NewAOF(config.AOFFile); err != nil {
		return err
	}

	log.Println("rewriting AOF file at", config.AOFFile)

	storeMutex.RLock()
	defer storeMutex.RUnlock()

	for k, obj := range store {
		if err = dumpKey(aof, *((*string)(k)), obj); err != nil {
			return err
		}
	}
	log.Println("AOF file rewrite complete")
	return nil
}
