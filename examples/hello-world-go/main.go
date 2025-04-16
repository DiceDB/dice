// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package main

import (
	"fmt"

	"github.com/dicedb/dicedb-go"
	"github.com/dicedb/dicedb-go/wire"
)

func main() {
	// create a new DiceDB client and connect to the server
	client, err := dicedb.NewClient("localhost", 7379)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// define a key and value
	key := "k1"
	value := "v1"

	// set the key and value
	resp := client.Fire(&wire.Command{
		Cmd:  "SET",
		Args: []string{key, value},
	})
	if resp.Status == wire.Status_ERR {
		fmt.Println("error setting key:", resp.Message)
		return
	}
	fmt.Printf("successfully set key %s=%s\n", key, value)

	// get the key and value
	resp = client.Fire(&wire.Command{
		Cmd:  "GET",
		Args: []string{key},
	})
	if resp.Status == wire.Status_ERR {
		fmt.Println("error getting key:", resp.Message)
		return
	}

	fmt.Printf("successfully got key %s=%s\n", key, resp.GetGETRes().Value)
}
