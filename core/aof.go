package core

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/server/utils"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
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
	err := os.MkdirAll(filepath.Dir(path), os.FileMode(FileMode))
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
	err := os.MkdirAll(filepath.Dir(config.TempAOFFile), os.FileMode(FileMode))
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

func BGREWRITEAOF(store *Store) (func(), error) {
	// Fork a child process, this child process would inherit all the uncommitted pages from main process.
	// This technique utilizes the CoW or copy-on-write, so while the main process is free to modify them
	// the child would save all the pages to disk.
	// Check details here -https://www.sobyte.net/post/2022-10/fork-cow/
	active := !atomic.CompareAndSwapInt32(&flushingInProgress, 0, 1)
	if active {
		return nil, fmt.Errorf("BGREWRITEAOF is already running, please try again later...")
	}

	var (
		aofSize  int64
		err      error
		childPID uintptr
		isChild  bool
	)

	withLocks(func() {
		// get the size of the current AOF file, it will be used
		// in order to concat the new rewritten AOF file with any additions
		// made to the existing AOF file during the rewrite process.
		info, err2 := os.Stat(config.AOFFile)
		if err2 != nil {
			if !errors.Is(err, os.ErrNotExist) {
				err = fmt.Errorf("failed getting aof file size: %v", err2)
				return
			}

			return
		}

		aofSize = info.Size()

		// Why forking inside withLocks()?
		//
		// if we would have fork outside this function the aof file size could grow more
		// by other store writes, leading to wrong concat of the additions.
		childPID, isChild, err2 = utils.Fork()
		if err2 != nil {
			err = fmt.Errorf("failed forking AOF child process: %v", err2)
			return
		}
	}, store, WithStoreLock())
	if err != nil {
		atomic.CompareAndSwapInt32(&flushingInProgress, 1, 0)
		return nil, err
	}

	//x, y := utils.PID()
	//fmt.Printf("PARENT: my pid: %d x: %d, y: %d\n", os.Getpid(), x, y)

	if isChild {
		time.Sleep(10 * time.Second)
		if err = DumpAllAOF(store); err != nil {
			log.Errorf("BGREWRITEAOF Process: %v", err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	return func() {
		defer atomic.CompareAndSwapInt32(&flushingInProgress, 1, 0)

		var ws syscall.WaitStatus
		_, err := syscall.Wait4(int(childPID), &ws, 0, nil)
		if err != nil {
			log.Errorf("failed waiting on BGREWRITEAOF process to complete: %v", err)
			return
		}

		if !ws.Exited() {
			log.Errorf("BGREWRITEAOF process didnt exited gracefully")
			return
		}

		withLocks(func() {
			currentAOF, err2 := os.Open(config.AOFFile)
			if err2 != nil {
				err = fmt.Errorf("failed opening aof file: %v", err2)
				return
			}

			defer currentAOF.Close()

			_, err2 = currentAOF.Seek(aofSize, 0)
			if err2 != nil {
				err = fmt.Errorf("failed seeking aof file: %v", err2)
				return
			}

			tmpAOF, err2 := os.OpenFile(config.AOFFile, os.O_APPEND|os.O_WRONLY, 0755)
			if err2 != nil {
				err = fmt.Errorf("failed opening tmp aof file: %v", err2)
				return
			}

			defer tmpAOF.Close()

			// concat any new additions to the current aof file to the new rewritten aof file
			_, err2 = io.Copy(tmpAOF, currentAOF)
			if err2 != nil {
				err = fmt.Errorf("failed concating current aof file to tmp aof: %v", err2)
				return
			}

			// Why do we make the server concat and rename the tmp aof file and
			// not the child process?
			//
			// because we want to make sure there is no ongoing write operation to the existing AOF file during a store mutation.
			// this is why we guard this operation with the store lock.
			err = os.Rename(config.TempAOFFile, config.AOFFile)
			if err != nil {
				log.Errorf("failed renaming AOF file: %w", err)
				return
			}

			log.Infof("AOF renamed successfully from \"%s\" to \"%s\"", config.TempAOFFile, config.AOFFile)
		}, store, WithStoreLock())
		if err != nil {

		}
	}, nil
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

	log.Infof("rewriting AOF file at %s", config.TempAOFFile)

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

	log.Info("AOF file rewrite complete")

	return nil
}

// TODO: Support Expiration
// TODO: Support non-kv data structures
// TODO: Support sync write
func dumpKey(aof *AOF, key string, obj *Obj) (err error) {
	cmd := fmt.Sprintf("SET %s %s", key, obj.Value)
	tokens := strings.Split(cmd, " ")
	return aof.writeWithoutPersist(string(Encode(tokens, false)))
}
