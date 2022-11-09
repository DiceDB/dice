package core

import (
	"fmt"
	"github.com/dicedb/dice/config"
	"log"
	"os"
)

func FileSync(f *os.File) error {
	_, _, err := syscall.Syscall(syscall.SYS_FCNTL, f.Fd(), syscall.F_FULLFSYNC, 0)
	if err == 0 {
		return nil
	}
	return err
}

// TODO: To to new and switch
func DumpAllAOF() error {
	fp, err := os.OpenFile(config.AOFFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	/* Note: A close function also returns an error, a plain defer is harmful.
	A successful close does not guarantee that the data has been successfully saved
	to disk,as the kernel uses the buffer cache to defer writes or write calls delays the writing
	to disk to mitigate cost of frequent writes to disk. A more reliable method is to use
	fsync() or f.Sync() and Close()
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
	return FileSync(fp)
}
