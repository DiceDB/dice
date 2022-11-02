package istorageengines

import (
	"sync"

	"github.com/dicedb/dice/object"
)

type IKVStorage interface {
	Put(key string, obj *object.Obj)
	Get(key string) *object.Obj
	Del(key string) bool
	GetCount() uint64
	GetStorage() *sync.Map
	// GetExpiry()
}
