// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package requestparser

import (
	"github.com/dicedb/dice/internal/cmd"
)

type Parser interface {
	Parse(data []byte) ([]*cmd.DiceDBCmd, error)
}
