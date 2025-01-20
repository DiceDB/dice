// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

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
