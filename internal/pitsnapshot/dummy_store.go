package pitsnapshot

import (
	"log"
	"sync"

	"github.com/dicedb/dice/internal/store"
)

type (
	DummyStore struct {
		store            map[string]string
		ongoingSnapshots map[uint64]store.Snapshotter
		snapLock         *sync.RWMutex
		stLock           *sync.RWMutex
	}
)

func NewDummyStore() *DummyStore {
	return &DummyStore{
		store:            make(map[string]string),
		ongoingSnapshots: make(map[uint64]store.Snapshotter),
		snapLock:         &sync.RWMutex{},
		stLock:           &sync.RWMutex{},
	}
}

func (store *DummyStore) Set(key, value string) {

	// Take a snapshot of the existing data since it's going to be overridden
	store.snapLock.RLock()
	for _, snapshot := range store.ongoingSnapshots {
		snapshot.TempAdd(key, value)
	}
	store.snapLock.RUnlock()

	store.stLock.Lock()
	defer store.stLock.Unlock()

	store.store[key] = value
}

func (store *DummyStore) Get(key string) string {
	store.stLock.RLock()
	defer store.stLock.RUnlock()

	return store.store[key]
}

func (store *DummyStore) StartSnapshot(snapshotID uint64, snapshot store.Snapshotter) (err error) {
	log.Println("Starting snapshot for snapshotID", snapshotID)

	store.snapLock.Lock()
	store.ongoingSnapshots[snapshotID] = snapshot
	store.snapLock.Unlock()

	store.stLock.RLock()
	keys := make([]string, 0, len(store.store))
	for k := range store.store {
		keys = append(keys, k)
	}
	store.stLock.RUnlock()

	for _, k := range keys {

		store.stLock.RLock()
		v := store.store[k]
		store.stLock.RUnlock()

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
	store.snapLock.Lock()
	if snapshot, isPresent := store.ongoingSnapshots[snapshotID]; isPresent {
		if err = snapshot.Close(); err != nil {
			log.Println("Error closing snapshot", "error", err)
		}
		delete(store.ongoingSnapshots, snapshotID)
	}
	store.snapLock.Unlock()
	return
}
