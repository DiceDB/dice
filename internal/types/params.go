package types

type Param string

const (
	CH   Param = "CH"
	INCR Param = "INCR"
	GT   Param = "GT"
	LT   Param = "LT"

	EX      Param = "EX"
	PX      Param = "PX"
	EXAT    Param = "EXAT"
	PXAT    Param = "PXAT"
	XX      Param = "XX"
	NX      Param = "NX"
	KEEPTTL Param = "KEEPTTL"
	GET     Param = "GET"

	PERSIST Param = "PERSIST"
)
