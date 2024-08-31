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

var (
	FileMode           int = 0644
	flushingInProgress int32
)

func NewAOF(path string) (*AOF, error) {
	err := os.Mkdir(filepath.Dir(path), os.FileMode(FileMode))
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

func NewTmpAOF() (*AOF, error) {
	err := os.Mkdir(filepath.Dir(config.TempAOFFile), os.FileMode(FileMode))
	if err != nil {
		return nil, err
	}

	// this differs from NewAOF function since we want to
	// truncate an existing tmp file in case it was not properly cleaned
	f, err := os.Create(config.TempAOFFile)
	if err != nil {
		return nil, err
	}

	return &AOF{
		file:   f,
		writer: bufio.NewWriter(f),
		path:   config.TempAOFFile,
	}, nil
}

func (a *AOF) Write(operation string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if err := a.writeWithoutPersist(operation); err != nil {
		return err
	}

	if err := a.writer.Flush(); err != nil {
		return err
	}

	return a.file.Sync()
}

func (a *AOF) writeWithoutPersist(operation string) error {
	if _, err := a.writer.WriteString(operation + "\n"); err != nil {
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
	return aof.writeWithoutPersist(string(Encode(tokens, false)))
}

// DumpAllAOF rewrites all store mutations to a tmp AOF file
// NOTE: this function is called from the spawned child process
func DumpAllAOF(store *Store) error {
	var (
		aof *AOF
		err error
	)

	if aof, err = NewTmpAOF(); err != nil {
		return err
	}

	defer func() {
		_ = aof.file.Close()

		if err != nil {
			// failure occurred during the flushing process
			// try deleting the tmp file
			_ = os.RemoveAll(config.TempAOFFile)
		}
	}()

	log.Println("rewriting AOF file at", config.TempAOFFile)

	for k, obj := range store.store {
		err = dumpKey(aof, k, obj)
		if err != nil {
			return err
		}
	}

	err = aof.writer.Flush()
	if err != nil {
		return fmt.Errorf("failed flushing AOF bufio to file: %w", err)
	}

	err = aof.file.Sync()
	if err != nil {
		return fmt.Errorf("failed flushing AOF to disk: %w", err)
	}

	log.Println("AOF file rewrite complete")

	return nil
}
