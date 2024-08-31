package core

import (
	"errors"
	"github.com/dicedb/dice/core/diceerrors"
)

func getType(te uint8) uint8 {
	return (te >> 4) << 4
}

func getEncoding(te uint8) uint8 {
	return te & 0b00001111
}

func assertType(te, t uint8) error {
	if getType(te) != t {
		return errors.New("the operation is not permitted on this type")
	}
	return nil
}

func assertEncoding(te, e uint8) error {
	if getEncoding(te) != e {
		return errors.New("the operation is not permitted on this encoding")
	}
	return nil
}

func assertTypeAndEncoding(typeEncoding, expectedType, expectedEncoding uint8) []byte {
	if err := assertType(typeEncoding, expectedType); err != nil {
		return diceerrors.NewErrWithMessage("Existing key has wrong Dice type")
	}
	if err := assertEncoding(typeEncoding, expectedEncoding); err != nil {
		return diceerrors.NewErrWithMessage("Existing key has wrong Dice type")
	}
	return nil
}
