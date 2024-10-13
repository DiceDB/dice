package store

import (
	"testing"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/object"
)

func TestLFUEviction(t *testing.T) {
	originalEvictionPolicy := config.DiceConfig.Memory.EvictionPolicy

	store := NewStore(nil, nil)
	config.DiceConfig.Memory.EvictionPolicy = config.EvictAllKeysLFU

	// Define test cases
	tests := []struct {
		name        string
		obj         []*object.Obj
		keys        []string
		setup       func(string, *object.Obj)
		perform     func()
		expectedKey string // false if we expect the key to be deleted
		sleep       bool
	}{
		{
			name: "Test LFU - eviction pool should have least recent key if all access keys are same",
			keys: []string{"k1", "k2", "k3", "k4"},
			obj:  []*object.Obj{{}, {}, {}, {}},
			setup: func(k string, obj *object.Obj) {
				store.Put(k, obj)
			},
			perform: func() {
				PopulateEvictionPool(store)
			},
			expectedKey: "k1",
			sleep:       true,
		},
		{
			name: "Test LFU - eviction pool should have least frequently used key",
			keys: []string{"k1", "k2", "k3", "k4"},
			obj:  []*object.Obj{{}, {}, {}, {}},
			setup: func(k string, obj *object.Obj) {
				store.Put(k, obj)
			},
			perform: func() {
				// ensuring approximate counter is incremented at least one time
				for i := 0; i < 100; i++ {
					store.Get("k1")
					store.Get("k2")
					store.Get("k3")
				}

				EPool = NewEvictionPool(0)
				PopulateEvictionPool(store)
			},
			expectedKey: "k4",
			sleep:       false,
		},
	}

	// Run test cases
	for _, tc := range tests {
		store.ResetStore()
		EPool = NewEvictionPool(0)

		t.Run(tc.name, func(t *testing.T) {

			for i := 0; i < len(tc.keys); i++ {
				tc.setup(tc.keys[i], tc.obj[i])
				if tc.sleep {
					time.Sleep(1000 * time.Millisecond)
				}
			}

			tc.perform()

			if EPool.pool[0].keyPtr != tc.expectedKey {
				t.Errorf("Expected: %s but got: %s\n", tc.expectedKey, EPool.pool[0].keyPtr)
			}
		})
	}

	config.DiceConfig.Memory.EvictionPolicy = originalEvictionPolicy
}
