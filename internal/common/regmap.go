package common

type RegMap[K comparable, V any] struct {
	M map[K]V
}

func (t *RegMap[K, V]) Put(key K, value V) {
	t.M[key] = value
}

func (t *RegMap[K, V]) Get(key K) (V, bool) {
	value, ok := t.M[key]
	return value, ok
}

func (t *RegMap[K, V]) Delete(key K) {
	delete(t.M, key)
}

func (t *RegMap[K, V]) Len() int {
	return len(t.M)
}

func (t *RegMap[K, V]) All(f func(k K, obj V) bool) {
	for k, v := range t.M {
		if !f(k, v) {
			break
		}
	}
}
