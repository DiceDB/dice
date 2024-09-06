package store

import (
	"errors"

	"github.com/dicedb/dice/internal/diceerrors"
)

func GetType(te uint8) uint8 {
	return (te >> 4) << 4
}

func GetEncoding(te uint8) uint8 {
	return te & 0b00001111
}

func AssertType(te, t uint8) error {
	if GetType(te) != t {
		return errors.New("the operation is not permitted on this type")
	}
	return nil
}

func AssertEncoding(te, e uint8) error {
	if GetEncoding(te) != e {
		return errors.New("the operation is not permitted on this encoding")
	}
	return nil
}

func AssertTypeAndEncoding(typeEncoding, expectedType, expectedEncoding uint8) []byte {
	if err := AssertType(typeEncoding, expectedType); err != nil {
		return diceerrors.NewErrWithMessage("Existing key has wrong Dice type")
	}
	if err := AssertEncoding(typeEncoding, expectedEncoding); err != nil {
		return diceerrors.NewErrWithMessage("Existing key has wrong Dice type")
	}
	return nil
}
