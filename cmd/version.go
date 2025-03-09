// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"

	"github.com/dicedb/dice/config"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print the version of DiceDB",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(config.DiceDBVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
