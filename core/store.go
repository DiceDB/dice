package core

import (
	"time"

	"github.com/dicedb/dice/config"
)

var store map[string]*Obj

func init() {
	store = make(map[string]*Obj)
}

func NewObj(value interface{}, durationMs int64, oType uint8, oEnc uint8) *Obj {
	var expiresAt int64 = -1
	if durationMs > 0 {
		expiresAt = time.Now().UnixMilli() + durationMs
	}

	return &Obj{
		Value:        value,
		TypeEncoding: oType | oEnc,
		ExpiresAt:    expiresAt,
	}
}

func Put(k string, obj *Obj) {
	if len(store) >= config.KeysLimit {
		evict()
	}
	store[k] = obj
	if KeyspaceStat[0] == nil {
		KeyspaceStat[0] = make(map[string]int)
	}
	KeyspaceStat[0]["keys"]++
}

func Get(k string) *Obj {
	v := store[k]
	if v != nil {
		if v.ExpiresAt != -1 && v.ExpiresAt <= time.Now().UnixMilli() {
			Del(k)
			return nil
		}
	}
	return v
}

func Del(k string) bool {
	if _, ok := store[k]; ok {
		delete(store, k)
		KeyspaceStat[0]["keys"]--
		return true
	}
	return false
}
