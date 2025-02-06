// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package svc

import (
	"fmt"
	"log"

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

func SendMessage(username, message string) {
	resp := client.Fire(&wire.Command{
		Cmd:  "SET",
		Args: []string{"last_message", fmt.Sprintf("%s:%s", username, message)},
	})
	if resp.Err != "" {
		fmt.Println("error sending message:", resp.Err)
	}
}

func Subscribe() {
	resp := client.Fire(&wire.Command{
		Cmd:  "GET.WATCH",
		Args: []string{"last_message"},
	})
	if resp.Err != "" {
		fmt.Println("error subscribing:", resp.Err)
	}
}

func ListenForMessages(onMessage func(message string)) {
	ch, err := client.WatchCh()
	if err != nil {
		panic(err)
	}
	for resp := range ch {
		log.Println("resp", resp.GetVStr())
		onMessage(resp.GetVStr())
	}
}
