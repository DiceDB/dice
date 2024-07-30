package server

import (
	"testing"
	"time"
)

// TestIOThreadPoolInitialization tests the initialization of IOThreadPool
func TestIOThreadPoolInitialization(t *testing.T) {
	ipool := NewIOThreadPool(4)
	if len(ipool.pool) != 4 {
		t.Fatalf("expected thread pool size of 4, got %d", len(ipool.pool))
	}

	// Verify if the pool contains the correct number of IOThreads
	for i := 0; i < 4; i++ {
		thread := ipool.Get()
		if thread == nil {
			t.Fatal("expected to retrieve an IOThread, got nil")
		}
		ipool.Put(thread)
	}
}

// TestShardPoolInitialization tests the initialization of ShardPool
func TestShardPoolInitialization(t *testing.T) {
	spool := NewShardPool(4)
	if len(spool.shardThreads) != 4 {
		t.Fatalf("expected shard pool size of 4, got %d", len(spool.shardThreads))
	}

	// Verify if each shard is properly initialized
	for i, shard := range spool.shardThreads {
		if shard == nil {
			t.Fatalf("expected shard at index %d to be initialized, got nil", i)
		}
	}
}

// TestSharedNothingModel tests the shared-nothing property of threads and shards
func TestSharedNothingModel(t *testing.T) {
	spool := NewShardPool(4)
	ipool := NewIOThreadPool(4)

	// Create a request
	req := &Request{}
	resch := make(chan *Result)
	op := &Operation{
		Key:      "key1",
		Value:    "value1",
		Op:       "op1",
		ResultCH: resch,
	}

	// Get a thread from the pool and send a request
	thread := ipool.Get()
	thread.reqch <- req

	// Submit an operation to the shard pool
	spool.Submit(op)

	// Ensure that the operation was processed
	select {
	case <-resch:
		// Operation was processed
	case <-time.After(time.Second):
		t.Fatal("operation was not processed in time")
	}

	// Check shared-nothing property: operations should be isolated
	shard := spool.shardThreads[0]
	// Add a dummy operation to the shard
	shard.reqch <- &Operation{
		Key:      "key2",
		Value:    "value2",
		Op:       "op2",
		ResultCH: make(chan *Result),
	}

	// Ensure the thread is not processing the shard's operations
	select {
	case res := <-resch:
		t.Fatalf("unexpected result received from a different shard: %v", res)
	case <-time.After(time.Millisecond * 100):
		// Expected behavior
	}
}
