// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package iomultiplexer

import "errors"

var (
	// ErrInvalidMaxClients is returned when the maxClients is less than 0
	ErrInvalidMaxClients = errors.New("invalid max clients")
)
