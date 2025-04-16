// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package common

import (
	"sync"
	"sync/atomic"
)

type RegMap[K comparable, V any] struct {
	DefaultV V
	M        sync.Map
	count    atomic.Int64
}

func (t *RegMap[K, V]) Put(key K, value V) {
	t.M.Store(key, value)
	t.count.Add(1)
}

func (t *RegMap[K, V]) Get(key K) (V, bool) {
	value, ok := t.M.Load(key)
	if !ok {
		return t.DefaultV, false
	}
	return value.(V), true
}

func (t *RegMap[K, V]) Delete(key K) {
	t.M.Delete(key)
	t.count.Add(-1)
}

func (t *RegMap[K, V]) Len() int {
	return int(t.count.Load())
}

func (t *RegMap[K, V]) All(f func(k K, obj V) bool) {
	t.M.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}
