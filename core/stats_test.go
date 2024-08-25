package core_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/dicedb/dice/core"
	"gotest.tools/v3/assert"
)

func TestIncrDecrStat(t *testing.T) {
	var wg sync.WaitGroup
	keyspace := *core.CreateKeyspaceStat()
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			keyspace.IncrStat("keys")
		}()
	}
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			keyspace.DecrKeys("keys")
		}()
	}
	wg.Wait()
	value := keyspace.GetStat("keys")
	assert.Equal(t, value, 7, fmt.Sprintf("Incorrect stat , Expected : %d, Actutal : %d\n", 7, value))
}
