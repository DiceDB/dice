// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package store

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestAOF(t *testing.T) {
	testFile := "test.aof"

	// Ensure cleanup after tests
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatalf("Failed to remove test file: %v", err)
		}
	}(testFile)

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

func TestAOFWithExat(t *testing.T) {
	testFile := "test.aof"

	defer os.Remove(testFile)

	t.Run("Create and Write with EXAT", func(t *testing.T) {
		aof, err := NewAOF(testFile)
		if err != nil {
			t.Fatalf("Failed to create AOF: %v", err)
		}
		defer aof.Close()

		futureTime := time.Now().Unix() + 60
		operations := []string{
			"SET key1 value1 EXAT " + strconv.FormatInt(futureTime, 10),
		}

		for _, op := range operations {
			if err := aof.Write(op); err != nil {
				t.Errorf("Failed to write operation: %v", err)
			}
		}
	})

	t.Run("Load and Verify EXAT", func(t *testing.T) {
		aof, err := NewAOF(testFile)
		if err != nil {
			t.Fatalf("Failed to open existing AOF: %v", err)
		}
		defer aof.Close()

		loadedOps, err := aof.Load()
		if err != nil {
			t.Fatalf("Failed to load operations: %v", err)
		}

		futureTime := time.Now().Unix() + 60
		expectedOps := []string{
			"SET key1 value1 EXAT " + strconv.FormatInt(futureTime, 10),
		}

		if !reflect.DeepEqual(loadedOps, expectedOps) {
			t.Errorf("Loaded operations do not match expected. Got %v, want %v", loadedOps, expectedOps)
		}

	})

	t.Run("Append to Existing with EXAT", func(t *testing.T) {
		aof, err := NewAOF(testFile)
		if err != nil {
			t.Fatalf("Failed to open existing AOF: %v", err)
		}

		newFutureTime := time.Now().Unix() + 120
		newOp := "SET key2 value2 EXAT " + strconv.FormatInt(newFutureTime, 10)
		if err := aof.Write(newOp); err != nil {
			t.Errorf("Failed to append operation: %v", err)
		}

		aof.Close()

		aof, err = NewAOF(testFile)
		if err != nil {
			t.Fatalf("Failed to open existing AOF: %v", err)
		}
		defer aof.Close()

		loadedOps, err := aof.Load()
		if err != nil {
			t.Fatalf("Failed to load operations after append: %v", err)
		}

		expectedOps := []string{
			"SET key1 value1 EXAT " + strconv.FormatInt(time.Now().Unix()+60, 10),
			newOp,
		}

		if !reflect.DeepEqual(loadedOps, expectedOps) {
			t.Errorf("Loaded operations after append do not match expected. Got %v, want %v", loadedOps, expectedOps)
		}
	})

	t.Run("Concurrent Writes with EXAT", func(t *testing.T) {
		os.Remove(testFile)
		aof, err := NewAOF(testFile)
		if err != nil {
			t.Fatalf("Failed to create AOF: %v", err)
		}
		defer aof.Close()

		concurrentOps := 100
		done := make(chan bool)

		for i := 0; i < concurrentOps; i++ {
			go func(i int) {
				op := fmt.Sprintf("SET key%d value%d EXAT %d", i, i, time.Now().Unix()+int64(60+i))
				if err := aof.Write(op); err != nil {
					t.Errorf("Failed concurrent write: %v", err)
				}
				done <- true
			}(i)
		}

		for i := 0; i < concurrentOps; i++ {
			<-done
		}

		aof, err = NewAOF(testFile)
		if err != nil {
			t.Fatalf("Failed to open existing AOF: %v", err)
		}
		defer aof.Close()

		loadedOps, err := aof.Load()
		if err != nil {
			t.Fatalf("Failed to load after concurrent writes: %v", err)
		}

		expectedLen := concurrentOps
		if len(loadedOps) != expectedLen {
			t.Errorf("Unexpected number of operations. Got %d, want %d", len(loadedOps), expectedLen)
		}
	})
}

func BenchmarkAOFWithExat(b *testing.B) {
	testFile := "test.aof"

	// Ensure cleanup after tests
	defer os.Remove(testFile)

	// Create the AOF instance
	aof, err := NewAOF(testFile)
	if err != nil {
		b.Fatalf("Failed to create AOF: %v", err)
	}
	defer aof.Close()

	futureTime := time.Now().Unix() + 60
	operation := "SET key1 value1 EXAT " + strconv.FormatInt(futureTime, 10)

	b.ResetTimer() // Reset the timer to exclude setup time

	for i := 0; i < b.N; i++ {
		if err := aof.Write(operation); err != nil {
			b.Errorf("Failed to write operation: %v", err)
		}
	}
}

func TestAOFConcurrentReadWrite(t *testing.T) {
	testFile := "test_concurrent.aof"
	defer os.Remove(testFile)

	aof, err := NewAOF(testFile)
	if err != nil {
		t.Fatalf("Failed to create AOF: %v", err)
	}
	defer aof.Close()

	for i := 0; i < 5; i++ {
		if err := aof.Write(fmt.Sprintf("INIT key%d value%d", i, i)); err != nil {
			t.Fatalf("Failed to write initial data: %v", err)
		}
	}

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				if err := aof.Write(fmt.Sprintf("WRITE-%d-%d key%d value%d", id, j, id*10+j, id*10+j)); err != nil {
					errors <- fmt.Errorf("write error: %v", err)
					return
				}
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				ops, err := aof.Load()
				if err != nil {
					errors <- fmt.Errorf("load error: %v", err)
					return
				}
				if len(ops) < 5 {
					errors <- fmt.Errorf("unexpected number of operations: got %d, want at least 5", len(ops))
					return
				}
				time.Sleep(time.Millisecond * 2)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Race condition detected: %v", err)
	}

	finalOps, err := aof.Load()
	if err != nil {
		t.Fatalf("Failed final load: %v", err)
	}

	expectedCount := 5 + (10 * 5)
	if len(finalOps) != expectedCount {
		t.Errorf("Final operation count mismatch: got %d, want %d", len(finalOps), expectedCount)
	}

	t.Logf("Successfully completed concurrent read/write test with %d operations", len(finalOps))
}
