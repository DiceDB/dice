// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/dicedb/dice/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var initConfigCmd = &cobra.Command{
	Use:   "init-config",
	Short: "creates a config file at dicedb.yaml with default values",
	Run: func(cmd *cobra.Command, args []string) {
		config.Init(cmd.Flags())
		_ = viper.WriteConfigAs(filepath.Join(config.DicedbDataDir, "dicedb.yaml"))
		fmt.Println("config created at ", config.DicedbDataDir)
	},
}

func init() {
	rootCmd.AddCommand(initConfigCmd)
}
