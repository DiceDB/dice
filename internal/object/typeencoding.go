package object

import (
	"errors"

	diceerrors "github.com/dicedb/dice/internal/errors"
)

func GetType(te uint8) uint8 {
	return (te >> 4) << 4
}

func AssertTypeWithError(te, t uint8) error {
	if GetType(te) != t {
		return errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	return nil
}

func AssertType(Type, expectedType uint8) []byte {
	if err := AssertTypeWithError(Type, expectedType); err != nil {
		return diceerrors.NewErrWithMessage(diceerrors.WrongKeyTypeErr)
	}
	return nil
}
