//go:build darwin && arm64

package eval

import (
	"syscall"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/clientio"
	diceerrors "github.com/dicedb/dice/internal/errors"
	dstore "github.com/dicedb/dice/internal/store"
)

/* Description - Spawn a background thread to persist the data via AOF technique. Current implementation is
based on CoW optimization and Fork */
// TODO: Implement Acknowledgement so that main process could know whether child has finished writing to its AOF file or not.
// TODO: Make it safe from failure, an stable policy would be to write the new flushes to a temporary files and then rename them to the main process's AOF file
// TODO: Add fsync() and fdatasync() to persist to AOF for above cases.
func EvalBGREWRITEAOF(args []string, store *dstore.Store) []byte {
	// Fork a child process, this child process would inherit all the uncommitted pages from main process.
	// This technique utilizes the CoW or copy-on-write, so while the main process is free to modify them
	// the child would save all the pages to disk.
	// Check details here -https://www.sobyte.net/post/2022-10/fork-cow/
	// TODO: Fix this to work with the threading
	// TODO: Problem at hand: In multi-threaded environment, each shard instance would fork a child process.
	// TODO: Each child process would now have a copy of the network file descriptor thus resulting in resource leaks.
	// TODO: We need to find an alternative approach for the multi-threaded environment.
	if config.EnableMultiThreading {
		return nil
	}

	// Get the original PID (Process ID) - This is needed to check if we are in child process or not.
	// The document approach is to check the return value of fork (it's 0 for child and non-zero for parent) but this is more reliable.
	// For more details check - https://github.com/DiceDB/dice/issues/683
	originalPID := syscall.Getpid()
	_, _, err := syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)

	if err != 0 {
		return diceerrors.NewErrWithMessage("Fork failed")
	}

	isChild := syscall.Getppid() == originalPID

	if isChild {
		// We are inside child process now, so we'll start flushing to disk.
		if err := dstore.DumpAllAOF(store); err != nil {
			return diceerrors.NewErrWithMessage("AOF failed")
		}
		syscall.Exit(0)
	}

	// Back to main thread
	return clientio.RespOK
}
