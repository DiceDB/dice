package common

type IStoreMap[K comparable, V any] interface {
	Put(key K, value V)
	Get(key K) (V, bool)
	Delete(key K) bool
	Keys() []K
	Len() int
	All(func(k K, obj V) bool)
}
