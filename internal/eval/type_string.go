package eval

import (
	"strconv"

	dstore "github.com/dicedb/dice/internal/object"
)

// Similar to
// tryObjectEncoding function in Redis
func deduceTypeEncoding(v string) (o, e uint8) {
	if _, err := strconv.ParseInt(v, 10, 64); err == nil {
		return dstore.ObjTypeInt, dstore.ObjEncodingInt
	}
	if len(v) <= 44 {
		return dstore.ObjTypeString, dstore.ObjEncodingEmbStr
	}
	return dstore.ObjTypeString, dstore.ObjEncodingRaw
}
