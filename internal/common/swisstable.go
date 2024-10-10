package common

import "github.com/cockroachdb/swiss"

type SwissTable[K comparable, V any] struct {
	M *swiss.Map[K, V]
}

func (t *SwissTable[K, V]) Put(key K, value V) {
	t.M.Put(key, value)
}

func (t *SwissTable[K, V]) Get(key K) (V, bool) {
	return t.M.Get(key)
}

func (t *SwissTable[K, V]) Delete(key K) {
	t.M.Delete(key)
}

func (t *SwissTable[K, V]) Len() int {
	return t.M.Len()
}

func (t *SwissTable[K, V]) All(f func(k K, obj V) bool) {
	t.M.All(f)
}
