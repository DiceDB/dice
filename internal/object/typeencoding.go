package object

import (
	"errors"

	diceerrors "github.com/dicedb/dice/internal/errors"
)

func GetType(te uint8) uint8 {
	return (te >> 4) << 4
}

func GetEncoding(te uint8) uint8 {
	return te & 0b00001111
}

func AssertType(te, t uint8) error {
	if GetType(te) != t {
		return errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
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
		return diceerrors.NewErrWithMessage(diceerrors.WrongKeyTypeErr)
	}
	if err := AssertEncoding(typeEncoding, expectedEncoding); err != nil {
		return diceerrors.NewErrWithMessage(diceerrors.WrongKeyTypeErr)
	}
	return nil
}
