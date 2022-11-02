package pool

import (
	"sync"

	"github.com/dicedb/dice/handlers"
	"github.com/dicedb/dice/object"
)

type diceWorker struct {
	wg  *sync.WaitGroup
	job func(i interface{})
}

func NewDiceWorker(f func(i interface{})) *diceWorker {
	return &diceWorker{
		wg:  new(sync.WaitGroup),
		job: f,
	}
}

func (dw *diceWorker) Work(dh *handlers.DiceKVstoreHandler) (err error) {

	// Get an instance of default dice pool obj
	pool, err := NewDefaultDicePool(func(i interface{}) {
		dw.job(i)
		dw.wg.Done()
	})
	defer pool.Release()
	if err != nil {
		return
	}

	// Iterate over the KV and pass each key and value
	dh.Store.Range(func(key, value interface{}) bool {
		var buffer = object.DiceWorkerBuffer{
			Key:   key.(string),
			Value: value.(*object.Obj),
		}
		dw.wg.Add(1)
		err = pool.Invoke(buffer)
		if err != nil {
			dw.wg.Done()
			return false
		}
		return true
	})
	dw.wg.Wait()
	return
}
