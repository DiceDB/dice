package core

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"sync"
	"syscall"

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

// TODO: Support Expiration
// TODO: Support non-kv data structures
// TODO: Support sync write
func dumpKey(aof *AOF, key string, obj *Obj) (err error) {
	cmd := fmt.Sprintf("SET %s %s", key, obj.Value)
	tokens := strings.Split(cmd, " ")
	return aof.Write(string(Encode(tokens, false)))
}

// TODO: To to new and switch
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

	log.Println("flush temp file's data to disk")
	if err := syscall.Fsync(int(aof.file.Fd())); err != nil {
		fmt.Println("fsync failed")
		return err
	}

	log.Println("rename tmp file to actual AOF file")
	if err := os.Rename(config.TempAOFFile, config.AOFFile); err != nil {
		fmt.Println("rename failed")
		return err
	}

	log.Println("AOF file rewrite complete")
	return err
}
