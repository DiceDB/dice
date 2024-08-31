package utils

import (
	"errors"
	"syscall"
)

func Fork() (uintptr, bool, error) {
	childPID, isChild, errno := syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)
	if errno != 0 {
		return 0, false, errors.New(errno.Error())
	}

	var child bool
	if isChild == 1 {
		child = true
	}

	return childPID, child, nil
}

func PID() (uintptr, uintptr) {
	x, y, err := syscall.Syscall(syscall.SYS_GETPID, 0, 0, 0)
	if err != 0 {
		panic(err)
	}

	return x, y
}
