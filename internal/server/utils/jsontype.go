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

func GetJSONFieldType(v interface{}) string {
	switch v.(type) {
	case map[string]interface{}:
		return ObjectType
	case []interface{}:
		return ArrayType
	case string:
		return StringType
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return IntegerType
	case float32, float64:
		return NumberType
	case bool:
		return BooleanType
	case nil:
		return NullType
	default:
		return UnknownType
	}
}
