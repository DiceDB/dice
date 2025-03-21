package ironhawk

import (
	"github.com/dicedb/dicedb-go/wire"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestConcurrentCommands(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	const numCommands = 10

	var wg sync.WaitGroup

	wg.Add(numCommands)

	for i := 0; i < numCommands; i++ {
		go func() {
			defer wg.Done()

			id := strconv.Itoa(i)

			// small delay to ensure commands overlap
			time.Sleep(time.Millisecond * 10)

			resp := client.Fire(&wire.Command{Cmd: "ECHO", Args: []string{id}})
			if resp.GetErr() != "" {
				t.Errorf("Expected no error, got %v", resp.GetErr())
			}
			if resp.GetVStr() != id {
				t.Errorf("Expected %s, got %v", id, resp.GetVStr())
			}
		}()
	}

	wg.Wait()
}
