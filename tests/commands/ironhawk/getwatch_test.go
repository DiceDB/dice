// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

const (
	getWatchKey = "getwatchkey"
)

type getWatchTestCase struct {
	key         string
	fingerprint string
	val         string
}

var getWatchTestCases = []getWatchTestCase{
	{getWatchKey, "2714318480", "value1"},
	{getWatchKey, "2714318480", "value2"},
	{getWatchKey, "2714318480", "value3"},
	{getWatchKey, "2714318480", "value4"},
}
