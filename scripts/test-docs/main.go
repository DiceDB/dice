// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package main

import (
	"fmt"
	"strings"

	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dicedb-go"
	"github.com/dicedb/dicedb-go/wire"
)

func main() {
	for _, c := range cmd.CommandRegistry.CommandMetas {
		if strings.Contains(c.Examples, "WATCH") {
			continue
		}
		for _, linesStr := range strings.Split(c.Examples, "localhost:7379>") {
			lines := strings.Split(linesStr, "\n")
			command := strings.TrimSpace(lines[0])
			if command == "" {
				continue
			}
			output := strings.TrimSpace(strings.Join(lines[1:], "\n"))
			fmt.Println("cmd", command)
			fmt.Println("output", output)

			client, err := dicedb.NewClient("localhost", 7379)
			if err != nil {
				fmt.Println("error", err)
				continue
			}

			res := client.Fire(&wire.Command{
				Cmd:  strings.Split(command, " ")[0],
				Args: strings.Split(command, " ")[1:],
			})

			// TODO: Write a function to compare the output
			// to the expected output
			// The idea is to test the documentation examples and making sure
			// what we ship in the docs is working as expected.
			fmt.Println("res", res)
		}
	}
}
