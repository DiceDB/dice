//go:build linux && arm64

package eval

import (
	"syscall"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/clientio"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/server/utils"
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
	childThreadID, _, _ := syscall.Syscall(syscall.SYS_GETTID, 0, 0, 0)
	newChild, _, _ := syscall.Syscall(syscall.SYS_CLONE, syscall.CLONE_PARENT_SETTID|syscall.CLONE_CHILD_CLEARTID|uintptr(syscall.SIGCHLD), 0, childThreadID)
	if newChild == 0 {
		// We are inside child process now, so we'll start flushing to disk.
		if err := dstore.DumpAllAOF(store); err != nil {
			return diceerrors.NewErrWithMessage("AOF failed")
		}
		return []byte(utils.EmptyStr)
	}
	// Back to main threadg
	return clientio.RespOK
}
