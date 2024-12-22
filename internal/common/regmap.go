// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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
