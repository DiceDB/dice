// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package iomultiplexer

const (
	// OpRead represents the read operation
	OpRead Operations = 1 << iota
	// OpWrite represents the write operation
	OpWrite
)
