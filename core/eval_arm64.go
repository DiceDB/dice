package core

import "syscall"

/* Description - Spawn a background thread to persist the data via AOF technique. Current implementation is
based on CoW optimization */
// TODO: Implement Acknowledgement so that main process could know whether child has finished writing to its AOF file or not.
// TODO: Make it safe from failure, an stable policy would be to write the new flushes to a temporary files and then rename them to the main process's AOF file
// TODO: Add fsync() and fdatasync() to persist to AOF for above cases.
func evalBGREWRITEAOF(args []string) []byte {
	// Fork a child process, this child process would inherit all the uncommitted pages from main process.
	// This technique utilizes the CoW or copy-on-write, so while the main process is free to modify them
	// the child would save all the pages to disk.
	//On arm64 platform the fork() is not available hence we use clone() system call to implement a copy-on-write
	childThreadID, _, _ := syscall.Syscall(syscall.SYS_GETTID, 0, 0, 0)
	pid, _, _ := syscall.Syscall(syscall.SYS_CLONE, syscall.CLONE_PARENT_SETTID|syscall.CLONE_CHILD_CLEARTID|uintptr(syscall.SIGCHLD), 0, childThreadID)
	if pid == 0 {
		//We are inside child process now, so we'll start flushing to disk.
		DumpAllAOF()
		//Replace
		syscall.Exit(0) // Clean up the child process
		return []byte("")
	} else {
		//Back to main thread
		return RESP_OK
	}
}
