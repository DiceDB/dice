package pitsnapshot

import (
	"log"

	"github.com/dicedb/dice/internal/store"
)

type (
	DummyStore struct {
		store            map[string]string
		ongoingSnapshots map[uint64]store.Snapshotter
	}
)

func NewDummyStore() *DummyStore {
	return &DummyStore{
		store:            make(map[string]string),
		ongoingSnapshots: make(map[uint64]store.Snapshotter),
	}
}

func (store *DummyStore) Set(key, value string) {

	// Take a snapshot of the existing data since it's going to be overridden
	for _, snapshot := range store.ongoingSnapshots {
		snapshot.TempAdd(key, value)
	}
	store.store[key] = value
}

func (store *DummyStore) Get(key string) string {
	return store.store[key]
}

func (store *DummyStore) StartSnapshot(snapshotID uint64, snapshot store.Snapshotter) (err error) {
	log.Println("Starting snapshot for snapshotID", snapshotID)
	store.ongoingSnapshots[snapshotID] = snapshot

	for k, v := range store.store {
		// Check if the data is overridden
		tempVal, _ := snapshot.TempGet(k)
		if tempVal != nil {
			v = tempVal.(string)
		}
		err = snapshot.Store(k, v)
		if err != nil {
			log.Println("Error storing data in snapshot", "error", err)
		}
	}
	store.StopSnapshot(snapshotID)
	return
}

func (store *DummyStore) StopSnapshot(snapshotID uint64) (err error) {
	if snapshot, isPresent := store.ongoingSnapshots[snapshotID]; isPresent {
		snapshot.Close()
		delete(store.ongoingSnapshots, snapshotID)
	}
	delete(store.ongoingSnapshots, snapshotID)
	return
}
