package common

type RegMap[K comparable, V any] struct {
	M map[K]V
}

func (rm *RegMap[K, V]) Put(key K, value V) {
	rm.M[key] = value
}

func (rm *RegMap[K, V]) Get(key K) (V, bool) {
	value, ok := rm.M[key]
	return value, ok
}

func (rm *RegMap[K, V]) Delete(key K) bool {
	if _, ok := rm.M[key]; ok {
		delete(rm.M, key)
		return true
	}
	return false
}

func (rm *RegMap[K, V]) Keys() []K {
	keys := make([]K, 0, len(rm.M))
	for k := range rm.M {
		keys = append(keys, k)
	}
	return keys
}

func (rm *RegMap[K, V]) Len() int {
	return len(rm.M)
}

func (rm *RegMap[K, V]) All(f func(k K, obj V) bool) {
	for k, v := range rm.M {
		if !f(k, v) {
			break
		}
	}
}
