package core

import (
	"path/filepath"
	"time"

	"github.com/dicedb/dice/config"
)

var store map[string]*Obj
var expires map[*Obj]uint64

type filterValTuple struct {
	Rkey   string
	Rvalue interface{}
}

func init() {
	store = make(map[string]*Obj)
	expires = make(map[*Obj]uint64)
}

func setExpiry(obj *Obj, expDurationMs int64) {
	expires[obj] = uint64(time.Now().UnixMilli()) + uint64(expDurationMs)
}

func NewObj(value interface{}, expDurationMs int64, oType uint8, oEnc uint8) *Obj {
	obj := &Obj{
		Value:          value,
		TypeEncoding:   oType | oEnc,
		LastAccessedAt: getCurrentClock(),
	}
	if expDurationMs > 0 {
		setExpiry(obj, expDurationMs)
	}
	return obj
}

func Put(k string, obj *Obj) {
	if len(store) >= config.KeysLimit {
		evict()
	}
	obj.LastAccessedAt = getCurrentClock()
	store[k] = obj
	if KeyspaceStat[0] == nil {
		KeyspaceStat[0] = make(map[string]int)
	}
	KeyspaceStat[0]["keys"]++
}

func Get(k string) *Obj {
	v := store[k]
	if v != nil {
		if hasExpired(v) {
			Del(k)
			return nil
		}
		v.LastAccessedAt = getCurrentClock()
	}
	return v
}

func Del(k string) bool {
	if obj, ok := store[k]; ok {
		delete(store, k)
		delete(expires, obj)
		KeyspaceStat[0]["keys"]--
		return true
	}
	return false
}

func FilterVals(pattern string) ([]filterValTuple, error) {
	var filteredVals []filterValTuple
	for key, value := range store {
		val := value.Value.(string)
		// look for a pattern match
		// TODO: use regex or something light weight
		ok, err := filepath.Match(pattern, val)
		if err != nil {
			return []filterValTuple{}, err
		}
		// match found
		if ok {
			keyValuePair := filterValTuple{Rkey: key, Rvalue: val}
			filteredVals = append(filteredVals, keyValuePair)
		}
	}

	return filteredVals, nil
}
