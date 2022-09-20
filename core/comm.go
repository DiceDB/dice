package core

import "syscall"

type FDComm struct {
	Fd int
}

func (f FDComm) Write(b []byte) (int, error) {
	return syscall.Write(f.Fd, b)
}

func (f FDComm) Read(b []byte) (int, error) {
	return syscall.Read(f.Fd, b)
}
