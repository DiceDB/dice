package core

import (
	"github.com/bytedance/sonic"
)

type DeepCopyable interface {
	DeepCopy() interface{}
}

func (obj *Obj) DeepCopy() *Obj {
	newObj := &Obj{
		TypeEncoding:   obj.TypeEncoding,
		LastAccessedAt: obj.LastAccessedAt,
	}

	// Use the DeepCopyable interface to deep copy the value
	if copier, ok := obj.Value.(DeepCopyable); ok {
		newObj.Value = copier.DeepCopy()
	} else {
		// Handle types that are not DeepCopyable
		sourceType, _ := ExtractTypeEncoding(obj)
		switch sourceType {
		case OBJ_TYPE_STRING:
			sourceValue := obj.Value.(string)
			newObj.Value = string([]byte(sourceValue))

		case OBJ_TYPE_JSON:
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
