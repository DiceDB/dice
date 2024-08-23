package core

import (
	"bufio"
	"fmt"
	"github.com/dicedb/dice/config"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

var FlushingInProgress int32

func NewAOF(path string) (*AOF, error) {
	err := os.MkdirAll(filepath.Dir(path), 0750)
	if err != nil {
		return nil, err
	}

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

// TODO: Support Expiration
// TODO: Support non-kv data structures
// TODO: Support sync write
func dumpKey(aof *AOF, key string, obj *Obj) (err error) {
	cmd := fmt.Sprintf("SET %s %s", key, obj.Value)
	tokens := strings.Split(cmd, " ")
	return aof.Write(string(Encode(tokens, false)))
}

func DumpAllAOF(store *Store) error {
	var (
		aof *AOF
		err error
	)

	if aof, err = NewAOF(config.TempAOFFile); err != nil {
		return err
	}

	defer aof.Close()

	log.Println("rewriting AOF file at", config.TempAOFFile)

	withLocks(func() {
		for k, obj := range store.store {
			err = dumpKey(aof, *((*string)(k)), obj)
		}
	}, store, WithStoreLock())

	err = aof.file.Sync()
	if err != nil {
		return fmt.Errorf("failed flushing AOF to disk: %w", err)
	}

	err = os.Rename(config.TempAOFFile, config.AOFFile)
	if err != nil {
		// TODO: consider cleanup on boot/here of zombie tmp files
		return fmt.Errorf("failed renaming AOF file: %w", err)
	}

	log.Println("AOF file rewrite complete")

	return nil
}
