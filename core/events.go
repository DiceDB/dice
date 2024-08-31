package core

import (
	"github.com/charmbracelet/log"
)

func Shutdown(store *Store) {
	waitForChild, err := BGREWRITEAOF(store)
	if err != nil {
		log.Error(err)
		return
	}

	// wait for child process to finish rewriting in a blocking manner in order to avoid corruption
	waitForChild()
}
