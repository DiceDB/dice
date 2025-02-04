// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package resp

const (
	getUnwatchKey = "getunwatchkey"
)

type getUnwatchTestCase struct {
	key string
	val string
}

var getUnwatchTestCases = []getUnwatchTestCase{
	{getUnwatchKey, "value1"},
	{getUnwatchKey, "value2"},
	{getUnwatchKey, "value3"},
	{getUnwatchKey, "value4"},
}
