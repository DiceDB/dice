package store

import ds "github.com/dicedb/dice/internal/datastructures"

type TestSDS struct {
	value int
	ds.BaseDataStructure[ds.DSInterface]
}

var _ ds.DSInterface = &TestSDS{}

func (t *TestSDS) Size() int {
	return 0
}

func (t *TestSDS) Serialize() []byte {
	return []byte{}
}
