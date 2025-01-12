// This file is part of DiceDB.
// Copyright (C) 2025DiceDB (dicedb.io).
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

import "github.com/cockroachdb/swiss"

type SwissTable[K comparable, V any] struct {
	M *swiss.Map[K, V]
}

func (t *SwissTable[K, V]) Put(key K, value V) {
	t.M.Put(key, value)
}

func (t *SwissTable[K, V]) Get(key K) (V, bool) {
	return t.M.Get(key)
}

func (t *SwissTable[K, V]) Delete(key K) {
	t.M.Delete(key)
}

func (t *SwissTable[K, V]) Len() int {
	return t.M.Len()
}

func (t *SwissTable[K, V]) All(f func(k K, obj V) bool) {
	t.M.All(f)
}
