package utils

import (
	"errors"
	"syscall"
)

func Fork() (uintptr, error) {
	_, childPID, errno := syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)
	if errno != 0 {
		return 0, errors.New(errno.Error())
	}

	return childPID, nil
}
