// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

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
