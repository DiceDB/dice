package core

import (
	"strconv"
	"github.com/dicedb/dice/object"
)
// Similar to
// tryObjectEncoding function in Redis
func deduceTypeEncoding(v string) (uint8, uint8) {
	oType := object.OBJ_TYPE_STRING
	if _, err := strconv.ParseInt(v, 10, 64); err == nil {
		return oType, object.OBJ_ENCODING_INT
	}
	if len(v) <= 44 {
		return oType, object.OBJ_ENCODING_EMBSTR
	}
	return oType, object.OBJ_ENCODING_RAW
}
