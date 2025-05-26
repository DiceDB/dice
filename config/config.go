// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	DiceDBVersion = "-"
)

// init initializes the DiceDBVersion variable by reading the
// VERSION file from the project root.
// This function runs automatically when the package is imported.
func init() {
	// Get the absolute path of the current file (config.go)
	// using runtime reflection
	_, currentFile, _, _ := runtime.Caller(0) //nolint:dogsled

	// Navigate up two directories from config.go to reach the project root
	// (config.go is in the config/ directory, so we need to go up twice)
	projectRoot := filepath.Dir(filepath.Dir(currentFile))

	// Read the VERSION file from the project root
	// This approach works regardless of where the program is executed from
	version, err := os.ReadFile(filepath.Join(projectRoot, "VERSION"))
	if err != nil {
		slog.Error("could not read the version file", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Store the version string in the package-level DiceDBVersion variable
	DiceDBVersion = strings.TrimSpace(string(version))
}

var Config *DiceDBConfig

type DiceDBConfig struct {
	Host string `mapstructure:"host" default:"0.0.0.0" description:"the host address to bind to"`
	Port int    `mapstructure:"port" default:"7379" description:"the port to bind to"`

	Username string `mapstructure:"username" default:"dicedb" description:"the username to use for authentication"`
	Password string `mapstructure:"password" default:"" description:"the password to use for authentication"`

	LogLevel string `mapstructure:"log-level" default:"info" description:"the log level"`

	EnableWatch bool `mapstructure:"enable-watch" default:"false" description:"enable support for .WATCH commands and real-time reactivity"`
	MaxClients  int  `mapstructure:"max-clients" default:"20000" description:"the maximum number of clients to accept"`
	NumShards   int  `mapstructure:"num-shards" default:"-1" description:"number of shards to create. defaults to number of cores"`

	Engine string `mapstructure:"engine" default:"ironhawk" description:"the engine to use, values: ironhawk"`

	EnableWAL                   bool   `mapstructure:"enable-wal" default:"false" description:"enable write-ahead logging"`
	WALVariant                  string `mapstructure:"wal-variant" default:"forge" description:"wal variant to use, values: forge"`
	WALDir                      string `mapstructure:"wal-dir" default:"logs" description:"the directory to store WAL segments"`
	WALBufferSizeMB             int    `mapstructure:"wal-buffer-size-mb" default:"1" description:"the size of the wal write buffer in megabytes"`
	WALRotationMode             string `mapstructure:"wal-rotation-mode" default:"time" description:"wal rotation mode to use, values: segment-size, time"`
	WALMaxSegmentSizeMB         int    `mapstructure:"wal-max-segment-size-mb" default:"16" description:"the maximum size of a wal segment file in megabytes before rotation"`
	WALSegmentRotationTimeSec   int    `mapstructure:"wal-max-segment-rotation-time-sec" default:"60" description:"the time interval (in seconds) after which wal a segment is rotated"`
	WALBufferSyncIntervalMillis int    `mapstructure:"wal-buffer-sync-interval-ms" default:"200" description:"the interval (in milliseconds) at which the wal write buffer is synced to disk"`
}

func Load(flags *pflag.FlagSet) {
	configureMetadataDir()
	viper.SetConfigName("dicedb")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(MetadataDir)

	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); !ok && err != nil {
		if err.Error() != "While parsing config: yaml: control characters are not allowed" {
			panic(err)
		}
	}

	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Name == "help" {
			return
		}

		// Only updated parsed configs if the user sets value or viper doesn't have default values for config flags set
		if flag.Changed || !viper.IsSet(flag.Name) {
			viper.Set(flag.Name, flag.Value.String())
		}
	})

	if err := viper.Unmarshal(&Config); err != nil {
		panic(err)
	}
}

// InitConfig initializes the config file.
// If the config file does not exist, it creates a new one.
// If the config file exists, it overwrites the existing config with the new key-values.
// and overwrite should replace the existing config with the new
// key-values and default values.
// If the metadata direcoty is inaccessible, then it uses the current working directory
// as the metadata directory.
func InitConfig(flags *pflag.FlagSet) {
	Load(flags)
	configPath := filepath.Join(MetadataDir, "dicedb.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		err := viper.WriteConfigAs(configPath)
		if err != nil {
			slog.Error("could not write the config file",
				slog.String("path", configPath),
				slog.String("error", err.Error()))
			os.Exit(1)
		}
		slog.Info("config created", slog.String("path", configPath))
	} else {
		if overwrite, _ := flags.GetBool("overwrite"); overwrite {
			err := viper.WriteConfigAs(configPath)
			if err != nil {
				slog.Error("could not write the config file",
					slog.String("path", configPath),
					slog.String("error", err.Error()))
				os.Exit(1)
			}

			// Current behavior - If key changed, then only overwrides.
			// TODO: Ideally, we should have a config-update function that updates the
			// existing config with the new key-values.
			// and overwrite should replace the existing config with the new
			// key-values and default values.
			slog.Info("config overwritten", slog.String("path", configPath))
		} else {
			slog.Info("config already exists. skipping.", slog.String("path", configPath))
			slog.Info("run with --overwrite to overwrite the existing config")
		}
	}
}

// configureMetadataDir creates the default metadata directory to be used
// for DiceDB metadataother persistent data
func configureMetadataDir() {
	// Creating dir with owner only permission
	// The reason we are lost logging the warning is because
	// this is not a hard dependency.
	// DiceDB can also run without metadata directory.and in that case
	// current directory will be used as metadata directory.
	if err := os.MkdirAll(MetadataDir, 0o700); err != nil {
		fmt.Printf("could not create metadata directory at %s. error: %s\n", MetadataDir, err)
		fmt.Println("using current directory as metadata directory")
		MetadataDir = "."
	}
}

func initDefaultConfig() *DiceDBConfig {
	defaultConfig := &DiceDBConfig{}
	configType := reflect.TypeOf(*defaultConfig)
	configValue := reflect.ValueOf(defaultConfig).Elem()

	for i := 0; i < configType.NumField(); i++ {
		field := configType.Field(i)
		value := configValue.Field(i)

		tag := field.Tag.Get("default")
		if tag != "" {
			switch value.Kind() {
			case reflect.String:
				value.SetString(tag)
			case reflect.Int:
				intVal := 0
				_, err := fmt.Sscanf(tag, "%d", &intVal)
				if err == nil {
					value.SetInt(int64(intVal))
				}
			case reflect.Bool:
				boolVal := false
				_, err := fmt.Sscanf(tag, "%t", &boolVal)
				if err == nil {
					value.SetBool(boolVal)
				}
			}
		}
	}

	return defaultConfig
}

func ForceInit(config *DiceDBConfig) {
	defaultConfig := initDefaultConfig()

	configType := reflect.TypeOf(*config)
	configValue := reflect.ValueOf(config).Elem()

	defaultConfigValue := reflect.ValueOf(defaultConfig).Elem()

	for i := 0; i < configType.NumField(); i++ {
		value := configValue.Field(i)
		defaultValue := defaultConfigValue.Field(i)
		if value.Interface() == reflect.Zero(value.Type()).Interface() {
			value.Set(defaultValue)
		}
	}

	Config = config
}
