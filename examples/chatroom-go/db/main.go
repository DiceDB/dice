// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package db

import "github.com/dicedb/dicedb-go"

var (
	Client *dicedb.Client
)

func init() {
	var err error
	Client, err = dicedb.NewClient("localhost", 7379)
	if err != nil {
		panic(err)
	}
}
