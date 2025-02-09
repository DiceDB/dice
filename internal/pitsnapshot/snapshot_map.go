package pitsnapshot

import (
	"bytes"
	"context"
	"encoding/gob"
	"log"
)

type (
	SnapshotMap struct {
		tempRepr map[string]interface{}
		buffer   []StoreMapUpdate
		flusher  *PITFlusher
		closing  bool
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
	}
	return

}

func (sm *SnapshotMap) TempAdd(key string, val interface{}) (err error) {
	if _, ok := sm.tempRepr[key]; ok {
		return
	}
	sm.tempRepr[key] = val
	return nil
}

func (sm *SnapshotMap) TempGet(key string) (interface{}, error) {
	return sm.tempRepr[key], nil
}

func (sm *SnapshotMap) Store(key string, val interface{}) (err error) {
	log.Println("Storing data in snapshot", "key", key, "value", val)
	if sm.closing {
		log.Println("rejecting writes to the update channel since the snapshot map is closing")
		return
	}
	sm.buffer = append(sm.buffer, StoreMapUpdate{Key: key, Value: val})
	if len(sm.buffer) == BufferSize {
		sm.flusher.updatesCh <- sm.buffer
		sm.buffer = []StoreMapUpdate{}
	}
	return
}

func (sm *SnapshotMap) Close() (err error) {
	sm.closing = true
	// Send the remaining updates
	if err = sm.flusher.Flush(sm.buffer); err != nil {
		return
	}
	sm.flusher.Close()
	return
}
