package core

import "strconv"

// Similar to
// tryObjectEncoding function in Redis
func deduceTypeEncoding(v string) (o, e uint8) {
	if _, err := strconv.ParseInt(v, 10, 64); err == nil {
		return ObjTypeInt, ObjEncodingInt
	}
	if len(v) <= 44 {
		return ObjTypeString, ObjEncodingEmbStr
	}
	return ObjTypeString, ObjEncodingRaw
}
