package hash

// KV represents a key-value pair, where K is the type of the key and V is the type of the value.
type KV[K comparable, V any] struct {
	Key   K
	Value V
}

const (
	NOT_FOUND  int64  = -2
	PAST       int64  = -1
	EXPIRY_SET int64  = 1
)
type IHashMap[K comparable, V any] interface {
	Set(key K, value V) (*V, bool)
	SetAll(data map[K]V) int64
	Get(key K) (*V, bool)
	Delete(key K)
	Keys() []K
	Values() []V
	Items() [][]interface{}
	Find(pattern K) []K
	Clear()
	GetExpiry(key K) (*V,int64)
	SetExpiry(key K, expiryMs int64) int64
	SetExpiryUnixMilli(key K, expiryMs uint64) int64
	HasExpired(key K) bool
	Len() int64
	ALen() int64
	// All(func(k K, obj V) bool)
	CreateOrModify(key K, modify func(old V)(V, error)) (*V, error)
	// Scan(pattern K, count int, offset int) []KV[K, V]
}
