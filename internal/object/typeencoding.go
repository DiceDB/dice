// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
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
