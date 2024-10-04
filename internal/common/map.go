package common

type ITable[K comparable, V any] interface {
	Put(key K, value V)
	Get(key K) (V, bool)
	Delete(key K)
	Len() int
	All(func(k K, obj V) bool)
}
