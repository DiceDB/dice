package pitsnapshot

import (
	"bytes"
	"context"
	"encoding/gob"
	"log"
	"sync"
)

type (
	SnapshotMap struct {
		tempRepr  map[string]interface{}
		buffer    []StoreMapUpdate
		flusher   *PITFlusher
		closing   bool
		mLock     *sync.RWMutex
		totalKeys uint64
	}
	StoreMapUpdate struct {
		Key   string
		Value interface{}
	}
)

func (s StoreMapUpdate) Serialize(updates []StoreMapUpdate) (updateBytes []byte, err error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(updates)
	if err != nil {
		return nil, err
	}
	updateBytes = buf.Bytes()
	return
}

func (s StoreMapUpdate) Deserialize(data []byte) (result []StoreMapUpdate, err error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err = dec.Decode(&result); err != nil {
		return
	}
	return
}

func NewSnapshotMap(ctx context.Context, flusher *PITFlusher) (sm *SnapshotMap, err error) {
	sm = &SnapshotMap{
		tempRepr: make(map[string]interface{}, 1000),
		flusher:  flusher,
		mLock:    &sync.RWMutex{},
	}
	return

}

func (sm *SnapshotMap) TempAdd(key string, val interface{}) (err error) {
	sm.mLock.Lock()
	defer sm.mLock.Unlock()

	if _, ok := sm.tempRepr[key]; ok {
		return
	}
	sm.tempRepr[key] = val
	return nil
}

func (sm *SnapshotMap) TempGet(key string) (interface{}, error) {
	sm.mLock.RLock()
	defer sm.mLock.RUnlock()

	return sm.tempRepr[key], nil
}

func (sm *SnapshotMap) Store(key string, val interface{}) (err error) {
	//log.Println("Storing data in snapshot", "key", key, "value", val)
	if sm.closing {
		log.Println("rejecting writes to the update channel since the snapshot map is closing")
		return
	}
	sm.buffer = append(sm.buffer, StoreMapUpdate{Key: key, Value: val})
	if len(sm.buffer) >= BufferSize {
		bufferCopy := make([]StoreMapUpdate, len(sm.buffer))
		copy(bufferCopy, sm.buffer)
		sm.totalKeys += uint64(len(bufferCopy))
		sm.flusher.updatesCh <- bufferCopy
		sm.buffer = []StoreMapUpdate{}
	}
	return
}

func (sm *SnapshotMap) Close() (err error) {
	sm.closing = true
	// Send the remaining updates
	//log.Println("Closing the snapshot map, sending the remaining updates to the flusher. Total keys processed", sm.totalKeys)
	if err = sm.flusher.Flush(sm.buffer); err != nil {
		return
	}
	sm.flusher.Close()
	if sm.totalKeys != sm.flusher.totalKeys {
		log.Println("[error] Total keys processed in the snapshot map and the flusher don't match", sm.totalKeys, sm.flusher.totalKeys)
	}
	return
}
