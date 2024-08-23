package utils

func Fork() (uintptr, error) {
	childPID, _, errno := syscall.Syscall(syscall.SYS_CLONE, 0, 0, 0)
	if errno != 0 {
		return 0, errors.New(errno.Error())
	}

	return childPID, nil
}
