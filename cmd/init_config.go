// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/config"
	"github.com/spf13/cobra"
)

var initConfigCmd = &cobra.Command{
	Use:   "config-init",
	Short: "creates a config file with default values",
	Run: func(cmd *cobra.Command, args []string) {
		config.InitConfig(cmd.Flags())
	},
}

func init() {
	initConfigCmd.Flags().BoolP("overwrite", "", false, "overwrite the existing config")
	rootCmd.AddCommand(initConfigCmd)
}
