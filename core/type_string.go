package core

import "strconv"

// Similar to
// tryObjectEncoding function in Redis
func deduceTypeEncoding(v string) (o, e uint8) {
	oType := ObjTypeString
	if _, err := strconv.ParseInt(v, 10, 64); err == nil {
		return oType, ObjEncodingInt
	}
	if len(v) <= 44 {
		return oType, ObjEncodingEmbStr
	}
	return oType, ObjEncodingRaw
}
