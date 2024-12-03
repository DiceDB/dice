package object

import (
	"errors"

	diceerrors "github.com/dicedb/dice/internal/errors"
)

func AssertTypeWithError(te, t ObjectType) error {
	if te != t {
		return errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	return nil
}

func AssertType(_type, expectedType ObjectType) []byte {
	if err := AssertTypeWithError(_type, expectedType); err != nil {
		return diceerrors.NewErrWithMessage(diceerrors.WrongKeyTypeErr)
	}
	return nil
}
