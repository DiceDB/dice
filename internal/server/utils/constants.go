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

const (
	EmptyStr = ""

	ObjectType      string = "object"
	ArrayType       string = "array"
	StringType      string = "string"
	IntegerType     string = "integer"
	NumberType      string = "number"
	BooleanType     string = "boolean"
	NullType        string = "null"
	UnknownType     string = "unknown"
	NumberZeroValue int    = 0
	JSONIngest      string = "JSON.INGEST"

	GET      string = "GET"
	SET      string = "SET"
	INCRBY   string = "INCRBY"
	OVERFLOW string = "OVERFLOW"
	WRAP     string = "WRAP"
	SAT      string = "SAT"
	FAIL     string = "FAIL"
	SIGNED   string = "SIGNED"
	UNSIGNED string = "UNSIGNED"
)
