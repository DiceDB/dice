package pitsnapshot

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func addDataToDummyStore(store *DummyStore) {
	store.Set("key1", "value1")
	store.Set("key2", "value2")
	store.Set("key3", "value3")
}

func TestSnapshot(t *testing.T) {
	// Create a new dummy store
	dummyStore := NewDummyStore()
	addDataToDummyStore(dummyStore)

	snapshot, err := NewPointInTimeSnapshot(context.Background(), dummyStore)

	assert.Nil(t, err)
	<-snapshot.ctx.Done()
}
