//go:build darwin && arm64

package eval

import (
	"github.com/dicedb/dice/internal/clientio"
	diceerrors "github.com/dicedb/dice/internal/errors"
	dstore "github.com/dicedb/dice/internal/store"
	"syscall"
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
	pid, _, _ := syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)

	if pid == 0 {
		// We are inside child process now, so we'll start flushing to disk.
		if err := dstore.DumpAllAOF(store); err != nil {
			return diceerrors.NewErrWithMessage("AOF failed")
		}
		syscall.Exit(0)
	} else if pid < 0 {
		// Fork failed
		return diceerrors.NewErrWithMessage("Fork failed")
	}

	// Back to main thread
	return clientio.RespOK
}
