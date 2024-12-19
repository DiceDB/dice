package hash

import (
	ds "github.com/dicedb/dice/internal/datastructures"
)

const (
	EncodingInt = iota
	EncodingString
)

type Hash struct {
	ds.BaseDataStructure[ds.DSInterface]
	Value map[string]HashItemInterface
}

func NewHash() ds.DSInterface {
	return &Hash{
		Value: make(map[string]HashItemInterface),
	}
}

func GetIfTypeHash(ds ds.DSInterface) (*Hash, bool) {
	h, ok := ds.(*Hash)
	return h, ok
}

func (s *Hash) Serialize() []byte {
	// add length of the set
	b := make([]byte, 0)
	b = append(b, byte(len(s.Value)))
	for key := range s.Value {
		if intKey, ok := any(key).(int64); ok {
			b = append(b, byte(intKey))
		} else if strKey, ok := any(key).(string); ok {
			b = append(b, []byte(strKey)...)
		}
	}
	return b
}

func (s *Hash) Add(key, value string, expiry int64) bool {
	old, ok := s.Value[key]
	if ok {
		old.Set(value)
		old.SetExpiry(expiry)
		return false
	}
	s.Value[key] = NewHashItem(value, expiry)
	return true
}

func (s *Hash) Get(key string) (string, bool) {
	item, ok := s.Value[key]
	if !ok {
		return "", false
	}
	return item.Get()
}

func (s *Hash) Del(key string) bool {
	_, ok := s.Value[key]
	if !ok {
		return false
	}
	delete(s.Value, key)
	return true
}

func (s *Hash) Exists(key string) bool {
	_, ok := s.Value[key]
	return ok
}
