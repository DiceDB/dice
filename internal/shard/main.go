// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package shard

import "github.com/dicedb/dice/internal/shardthread"

type Shard struct {
	ID     int
	Thread *shardthread.ShardThread
}
