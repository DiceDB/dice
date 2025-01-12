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

package object

import (
	"github.com/bytedance/sonic"
)

type DeepCopyable interface {
	DeepCopy() interface{}
}

func (obj *Obj) DeepCopy() *Obj {
	newObj := &Obj{
		Type:           obj.Type,
		LastAccessedAt: obj.LastAccessedAt,
	}

	// Use the DeepCopyable interface to deep copy the value
	if copier, ok := obj.Value.(DeepCopyable); ok {
		newObj.Value = copier.DeepCopy()
	} else {
		// Handle types that are not DeepCopyable
		sourceType := obj.Type
		switch sourceType {
		case ObjTypeString:
			sourceValue := obj.Value.(string)
			newObj.Value = sourceValue

		case ObjTypeJSON:
			sourceValue := obj.Value
			jsonStr, err := sonic.MarshalString(sourceValue)
			if err != nil {
				return nil
			}
			var value interface{}
			err = sonic.UnmarshalString(jsonStr, &value)
			if err != nil {
				return nil
			}
			newObj.Value = value

		default:
			return nil
		}
	}

	return newObj
}
