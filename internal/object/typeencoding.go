package object

import (
	diceerrors "github.com/dicedb/dice/internal/errors"
)

func AssertTypeWithError(te, t ObjectType) error {
	if te != t {
		return diceerrors.ErrWrongTypeOperation
	}
	return nil
}

func AssertType(_type, expectedType ObjectType) []byte {
	if err := AssertTypeWithError(_type, expectedType); err != nil {
		return diceerrors.NewErrWithMessage(diceerrors.WrongKeyTypeErr)
	}
	return nil
}
