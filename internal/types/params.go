// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

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

	PERSIST Param = "PERSIST"
)
