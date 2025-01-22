// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	DiceDBVersion = "0.1.0"
)

type DiceDBConfig struct {
	Host string `mapstructure:"host" default:"0.0.0.0" description:"the host address to bind to"`
	Port int    `mapstructure:"port" default:"7379" description:"the port to bind to"`

	Username string `mapstructure:"username" default:"dicedb" description:"the username to use for authentication"`
	Password string `mapstructure:"password" default:"" description:"the password to use for authentication"`

	LogLevel string `mapstructure:"log-level" default:"info" description:"the log level"`

	EnableWatch bool `mapstructure:"enable-watch" default:"false" description:"enable support for .WATCH commands and real-time reactivity"`
	MaxClients  int  `mapstructure:"max-clients" default:"20000" description:"the maximum number of clients to accept"`
	NumShards   int  `mapstructure:"num-shards" default:"-1" description:"number of shards to create. defaults to number of cores"`

	EnableWAL                         bool   `mapstructure:"enable-wal" default:"true" description:"enable write-ahead logging"`
	WALEngine                         string `mapstructure:"wal-engine" default:"aof" description:"wal engine to use, values: sqlite, aof"`
	WALDir                            string `mapstructure:"wal-dir" default:"/var/log/dicedb" description:"the directory to store WAL segments"`
	WALMode                           string `mapstructure:"wal-mode" default:"buffered" description:"wal mode to use, values: buffered, unbuffered"`
	WALWriteMode                      string `mapstructure:"wal-write-mode" default:"default" description:"wal file write mode to use, values: default, fsync"`
	WALBufferSizeMB                   int    `mapstructure:"wal-buffer-size-mb" default:"1" description:"the size of the wal write buffer in megabytes"`
	WALRotationMode                   string `mapstructure:"wal-rotation-mode" default:"segment-size" description:"wal rotation mode to use, values: segment-size, time"`
	WALMaxSegmentSizeMB               int    `mapstructure:"wal-max-segment-size-mb" default:"16" description:"the maximum size of a wal segment file in megabytes before rotation"`
	WALMaxSegmentRotationTimeSec      int    `mapstructure:"wal-max-segment-rotation-time-sec" default:"60" description:"the time interval (in seconds) after which wal a segment is rotated"`
	WALBufferSyncIntervalMillis       int    `mapstructure:"wal-buffer-sync-interval-ms" default:"200" description:"the interval (in milliseconds) at which the wal write buffer is synced to disk"`
	WALRetentionMode                  string `mapstructure:"wal-retention-mode" default:"num-segments" description:"the new horizon for wal segment post cleanup. values: num-segments, time, checkpoint"`
	WALMaxSegmentCount                int    `mapstructure:"wal-max-segment-count" default:"10" description:"the maximum number of segments to retain, if the retention mode is 'num-segments'"`
	WALMaxSegmentRetentionDurationSec int    `mapstructure:"wal-max-segment-retention-duration-sec" default:"600" description:"the maximum duration (in seconds) for wal segments retention"`
	WALRecoveryMode                   string `mapstructure:"wal-recovery-mode" default:"strict" description:"wal recovery mode in case of a corruption, values: strict, truncate, ignore"`
}

var GlobalDiceDBConfig *DiceDBConfig

func Init(flags *pflag.FlagSet) {
	viper.SetConfigName("dicedb")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/dicedb")

	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); !ok && err != nil {
		panic(err)
	}

	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Name == "help" {
			return
		}
		viper.Set(flag.Name, flag.Value.String())
	})

	if err := viper.Unmarshal(&GlobalDiceDBConfig); err != nil {
		panic(err)
	}
}
