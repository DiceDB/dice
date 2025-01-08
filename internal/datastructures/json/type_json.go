package json

import (
	ds "github.com/dicedb/dice/internal/datastructures"
)

type JSONString struct {
	ds.BaseDataStructure[ds.DSInterface]
	Value string
}

func NewJSONString() ds.DSInterface {
	return &JSONString{
		Value: "",
	}
}

func GetIfTypeJSONString(ds ds.DSInterface) (*JSONString, bool) {
	h, ok := ds.(*JSONString)
	return h, ok
}

func (s *JSONString) Serialize() []byte {
	return []byte(s.Value)
}
