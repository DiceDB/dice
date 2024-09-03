package constants

const (
	EmptyStr = ""

	Ex      string = "EX"
	Px      string = "PX"
	Pxat    string = "PXAT"
	Exat    string = "EXAT"
	XX      string = "XX"
	NX      string = "NX"
	Xx      string = "xx"
	Nx      string = "nx"
	GT      string = "GT"
	LT      string = "LT"
	KEEPTTL string = "KEEPTTL"
	Keepttl string = "keepttl"
	Sync    string = "SYNC"
	Async   string = "ASYNC"

	AND string = "AND"
	OR  string = "OR"
	XOR string = "XOR"
	NOT string = "NOT"

	Asc string = "asc"

	String string = "string"
	Int    string = "int"
	Float  string = "float"
	Bool   string = "bool"
	Nil    string = "nil"

	OperatorEquals              string = "="
	OperatorNotEquals           string = "!="
	OperatorNotEqualsTo         string = "<>"
	OperatorLessThanEqualsTo    string = "<="
	OperatorGreaterThanEqualsTo string = ">="

	ObjectType      string = "object"
	ArrayType       string = "array"
	StringType      string = "string"
	IntegerType     string = "integer"
	NumberType      string = "number"
	BooleanType     string = "boolean"
	NullType        string = "null"
	UnknownType     string = "unknown"
	NumberZeroValue int    = 0

	Qwatch string = "qwatch"
)

// Temporary set for ignoring these commands while tcl tests.
// Once these commands are implemented we can remove them from the set one by one.
var IgnoreCommands = map[string]string{
	"SELECT":    "ignore for tcl test",
	"FUNCTION":  "ignore for tcl test",
	"FLUSHALL":  "ignore for tcl test",
	"RPUSH":     "ignore for tcl test",
	"HGET":      "ignore for tcl test",
	"LRANGE":    "ignore for tcl test",
	"ACL":       "ignore for tcl test",
	"FLUSHDB":   "ignore for tcl test",
	"SCAN":      "ignore for tcl test",
	"SCARD":     "ignore for tcl test",
	"SLAVEOF":   "ignore for tcl test",
	"BLPOP":     "ignore for tcl test",
	"ZADD":      "ignore for tcl test",
	"BZPOPMIN":  "ignore for tcl test",
	"BZPOPMAX":  "ignore for tcl test",
	"DEBUG":     "ignore for tcl test",
	"REPLICAOF": "ignore for tcl test",
	"SAVE":      "ignore for tcl test",
	"CONFIG":    "ignore for tcl test",
}
