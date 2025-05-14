// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package svc

import (
	"fmt"
	"strconv"

	"github.com/dicedb/dicedb-go"
	"github.com/dicedb/dicedb-go/wire"
)

var (
	client *dicedb.Client
)

func init() {
	var err error
	client, err = dicedb.NewClient("localhost", 7379)
	if err != nil {
		panic(err)
	}
}

func UpdateScore(player string, score int) {
	resp := client.Fire(&wire.Command{
		Cmd:  "ZADD",
		Args: []string{"game:scores", strconv.Itoa(score), player},
	})
	if resp.Status == wire.Status_ERR {
		fmt.Println("error updating score:", resp.Message)
	}
}

func Subscribe() {
	resp := client.Fire(&wire.Command{
		Cmd:  "ZRANGE.WATCH",
		Args: []string{"game:scores", "1", "5", "BYRANK"},
	})
	if resp.Status == wire.Status_ERR {
		fmt.Println("error subscribing:", resp.Message)
	}
}

func ListenForMessages(onMessage func(result *wire.Result)) {
	ch, err := client.WatchCh()
	if err != nil {
		panic(err)
	}
	for resp := range ch {
		if resp.Status == wire.Status_ERR {
			fmt.Println("error listening for messages:", resp.Message)
		} else {
			onMessage(resp)
		}
	}
}
