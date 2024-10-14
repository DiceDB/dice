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
