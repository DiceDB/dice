// This file is part of DiceDB.
// Copyright (C) 2025  DiceDB (dicedb.io).
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

package utils

import (
	"strconv"
	"strings"

	diceerrors "github.com/dicedb/dice/internal/errors"
)

type BitFieldOp struct {
	Kind   string
	EType  string
	EVal   int64
	Offset int64
	Value  int64
}

// parseEncodingAndOffet function parses offset and encoding type for bitfield commands
// as this part is common to all subcommands
func parseBitfieldEncodingAndOffset(args []string) (eType, eVal, offset interface{}, err error) {
	encodingRaw := args[0]
	offsetRaw := args[1]
	switch encodingRaw[0] {
	case 'i':
		eType = SIGNED
		eVal, err = strconv.ParseInt(encodingRaw[1:], 10, 64)
		if err != nil {
			err = diceerrors.ErrGeneral("Invalid bitfield type. Use something like i16 u8. Note that u64 is not supported but i64 is")
			return eType, eVal, offset, err
		}
		if eVal.(int64) <= 0 || eVal.(int64) > 64 {
			err = diceerrors.ErrGeneral("Invalid bitfield type. Use something like i16 u8. Note that u64 is not supported but i64 is")
			return eType, eVal, offset, err
		}
	case 'u':
		eType = UNSIGNED
		eVal, err = strconv.ParseInt(encodingRaw[1:], 10, 64)
		if err != nil {
			err = diceerrors.ErrGeneral("Invalid bitfield type. Use something like i16 u8. Note that u64 is not supported but i64 is")
			return eType, eVal, offset, err
		}
		if eVal.(int64) <= 0 || eVal.(int64) >= 64 {
			err = diceerrors.ErrGeneral("Invalid bitfield type. Use something like i16 u8. Note that u64 is not supported but i64 is")
			return eType, eVal, offset, err
		}
	default:
		err = diceerrors.ErrGeneral("Invalid bitfield type. Use something like i16 u8. Note that u64 is not supported but i64 is")
		return eType, eVal, offset, err
	}

	switch offsetRaw[0] {
	case '#':
		offset, err = strconv.ParseInt(offsetRaw[1:], 10, 64)
		if err != nil {
			err = diceerrors.ErrGeneral("bit offset is not an integer or out of range")
			return eType, eVal, offset, err
		}
		offset = offset.(int64) * eVal.(int64)
	default:
		offset, err = strconv.ParseInt(offsetRaw, 10, 64)
		if err != nil {
			err = diceerrors.ErrGeneral("bit offset is not an integer or out of range")
			return eType, eVal, offset, err
		}
	}
	return eType, eVal, offset, err
}

func ParseBitfieldOps(args []string, readOnly bool) (ops []BitFieldOp, err error) {
	var overflowType string

	for i := 1; i < len(args); {
		isReadOnlyCommand := false
		switch strings.ToUpper(args[i]) {
		case GET:
			if len(args) <= i+2 {
				return nil, diceerrors.ErrSyntax
			}
			eType, eVal, offset, err := parseBitfieldEncodingAndOffset(args[i+1 : i+3])
			if err != nil {
				return nil, err
			}
			ops = append(ops, BitFieldOp{
				Kind:   GET,
				EType:  eType.(string),
				EVal:   eVal.(int64),
				Offset: offset.(int64),
				Value:  int64(-1),
			})
			i += 3
			isReadOnlyCommand = true
		case SET:
			if len(args) <= i+3 {
				return nil, diceerrors.ErrSyntax
			}
			eType, eVal, offset, err := parseBitfieldEncodingAndOffset(args[i+1 : i+3])
			if err != nil {
				return nil, err
			}
			value, err1 := strconv.ParseInt(args[i+3], 10, 64)
			if err1 != nil {
				return nil, diceerrors.ErrIntegerOutOfRange
			}
			ops = append(ops, BitFieldOp{
				Kind:   SET,
				EType:  eType.(string),
				EVal:   eVal.(int64),
				Offset: offset.(int64),
				Value:  value,
			})
			i += 4
		case INCRBY:
			if len(args) <= i+3 {
				return nil, diceerrors.ErrSyntax
			}
			eType, eVal, offset, err := parseBitfieldEncodingAndOffset(args[i+1 : i+3])
			if err != nil {
				return nil, err
			}
			value, err1 := strconv.ParseInt(args[i+3], 10, 64)
			if err1 != nil {
				return nil, diceerrors.ErrIntegerOutOfRange
			}
			ops = append(ops, BitFieldOp{
				Kind:   INCRBY,
				EType:  eType.(string),
				EVal:   eVal.(int64),
				Offset: offset.(int64),
				Value:  value,
			})
			i += 4
		case OVERFLOW:
			if len(args) <= i+1 {
				return nil, diceerrors.ErrSyntax
			}
			switch strings.ToUpper(args[i+1]) {
			case WRAP, FAIL, SAT:
				overflowType = strings.ToUpper(args[i+1])
			default:
				return nil, diceerrors.ErrGeneral("Invalid OVERFLOW type specified")
			}
			ops = append(ops, BitFieldOp{
				Kind:   OVERFLOW,
				EType:  overflowType,
				EVal:   int64(-1),
				Offset: int64(-1),
				Value:  int64(-1),
			})
			i += 2
		default:
			return nil, diceerrors.ErrSyntax
		}

		if readOnly && !isReadOnlyCommand {
			return nil, diceerrors.ErrGeneral("BITFIELD_RO only supports the GET subcommand")
		}
	}

	return ops, nil
}
