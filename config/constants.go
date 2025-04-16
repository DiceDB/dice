// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package config

import "time"

const (
	EvictionRatio      float64       = 0.9
	DefaultKeysLimit   int           = 200000000
	ShardCronFrequency time.Duration = 1 * time.Second
	EnableProfile      bool          = false

	KeepAlive int32 = 300
	Timeout   int32 = 300

	DefaultConnBacklogSize = 128

	MaxRequestSize = 32 * 1024 * 1024 // 32 MB
	IoBufferSize   = 16 * 1024        // 16 KB
)
