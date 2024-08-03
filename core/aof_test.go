package core

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestAOF(t *testing.T) {
	testFile := "test.aof"

	// Ensure cleanup after tests
	defer os.Remove(testFile)

	t.Run("Create and Write", func(t *testing.T) {
		aof, err := NewAOF(testFile)
		if err != nil {
			t.Fatalf("Failed to create AOF: %v", err)
		}
		defer aof.Close()

		operations := []string{
			"SET key1 value1",
			"SET key2 value2",
			"DEL key1",
		}

		for _, op := range operations {
			if err := aof.Write(op); err != nil {
				t.Errorf("Failed to write operation: %v", err)
			}
		}
	})

	t.Run("Load and Verify", func(t *testing.T) {
		aof, err := NewAOF(testFile)
		if err != nil {
			t.Fatalf("Failed to open existing AOF: %v", err)
		}
		defer aof.Close()

		loadedOps, err := aof.Load()
		if err != nil {
			t.Fatalf("Failed to load operations: %v", err)
		}

		expectedOps := []string{
			"SET key1 value1",
			"SET key2 value2",
			"DEL key1",
		}

		if !reflect.DeepEqual(loadedOps, expectedOps) {
			t.Errorf("Loaded operations do not match expected. Got %v, want %v", loadedOps, expectedOps)
		}
	})

	t.Run("Append to Existing", func(t *testing.T) {
		aof, err := NewAOF(testFile)
		if err != nil {
			t.Fatalf("Failed to open existing AOF: %v", err)
		}

		newOp := "SET key3 value3"
		if err := aof.Write(newOp); err != nil {
			t.Errorf("Failed to append operation: %v", err)
		}

		aof.Close()

		// Reload and verify
		aof, _ = NewAOF(testFile)
		loadedOps, err := aof.Load()
		if err != nil {
			t.Fatalf("Failed to load operations after append: %v", err)
		}

		expectedOps := []string{
			"SET key1 value1",
			"SET key2 value2",
			"DEL key1",
			"SET key3 value3",
		}

		if !reflect.DeepEqual(loadedOps, expectedOps) {
			t.Errorf("Loaded operations after append do not match expected. Got %v, want %v", loadedOps, expectedOps)
		}
	})

	t.Run("Concurrent Writes", func(t *testing.T) {
		aof, err := NewAOF(testFile)
		if err != nil {
			t.Fatalf("Failed to create AOF: %v", err)
		}
		defer aof.Close()

		concurrentOps := 100
		done := make(chan bool)

		for i := 0; i < concurrentOps; i++ {
			go func(i int) {
				op := fmt.Sprintf("SET key%d value%d", i, i)
				if err := aof.Write(op); err != nil {
					t.Errorf("Failed concurrent write: %v", err)
				}
				done <- true
			}(i)
		}

		for i := 0; i < concurrentOps; i++ {
			<-done
		}

		loadedOps, err := aof.Load()
		if err != nil {
			t.Fatalf("Failed to load after concurrent writes: %v", err)
		}

		if len(loadedOps) != concurrentOps+4 { // +4 for previous operations
			t.Errorf("Unexpected number of operations. Got %d, want %d", len(loadedOps), concurrentOps+4)
		}
	})
}
