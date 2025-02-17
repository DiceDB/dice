// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strconv"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/logger"
	"github.com/dicedb/dice/server"
	"github.com/spf13/cobra"
)

func init() {
	flags := rootCmd.PersistentFlags()

	c := config.DiceDBConfig{}
	_type := reflect.TypeOf(c)
	for i := 0; i < _type.NumField(); i++ {
		field := _type.Field(i)
		yamlTag := field.Tag.Get("mapstructure")
		descriptionTag := field.Tag.Get("description")
		defaultTag := field.Tag.Get("default")

		switch field.Type.Kind() {
		case reflect.String:
			flags.String(yamlTag, defaultTag, descriptionTag)
		case reflect.Int:
			val, _ := strconv.Atoi(defaultTag)
			flags.Int(yamlTag, val, descriptionTag)
		case reflect.Bool:
			val, _ := strconv.ParseBool(defaultTag)
			flags.Bool(yamlTag, val, descriptionTag)
		}
	}
}

var rootCmd = &cobra.Command{
	Use:   "dicedb",
	Short: "an in-memory database;",
	Run: func(cmd *cobra.Command, args []string) {
		config.Load(cmd.Flags())
		slog.SetDefault(logger.New())
		server.Start()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
