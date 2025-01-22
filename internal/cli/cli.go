// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cli

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/fatih/color"
)

// configuration function used to add configuration values to the print table at the startup.
// add entry to this function to add a new row in the startup configuration table.
func printConfiguration() {
	// Add the version of the DiceDB
	slog.Info("starting DiceDB", slog.String("version", config.DiceDBVersion))

	// Add the port number on which DiceDB is running
	slog.Info("running with", slog.Int("port", config.DiceConfig.RespServer.Port))

	//	 HTTP and WebSocket server configuration
	if config.DiceConfig.HTTP.Enabled {
		slog.Info("running with", slog.Int("http-port", config.DiceConfig.HTTP.Port))
	}

	if config.DiceConfig.WebSocket.Enabled {
		slog.Info("running with", slog.Int("websocket-port", config.DiceConfig.WebSocket.Port))
	}

	// Add the number of CPU cores available on the machine
	slog.Info("running with", slog.Int("cores", runtime.NumCPU()))

	// Conditionally add the number of shards to be used for DiceDB
	numShards := runtime.NumCPU()
	if config.DiceConfig.Performance.NumShards > 0 {
		numShards = config.DiceConfig.Performance.NumShards
	}
	slog.Info("running with", slog.Int("shards", numShards))

	// Add whether the watch feature is enabled
	slog.Info("running with", slog.Bool("watch", config.DiceConfig.Performance.EnableWatch))

	// Add whether the watch feature is enabled
	slog.Info("running with", slog.Bool("profiling", config.DiceConfig.Performance.EnableProfiling))

	// Add whether the persistence feature is enabled
	slog.Info("running with", slog.Bool("persistence", config.DiceConfig.Persistence.Enabled))
}

// printConfigTable prints key-value pairs in a vertical table format.
func render() {
	fmt.Print(`
	██████╗ ██╗ ██████╗███████╗██████╗ ██████╗ 
	██╔══██╗██║██╔════╝██╔════╝██╔══██╗██╔══██╗
	██║  ██║██║██║     █████╗  ██║  ██║██████╔╝
	██║  ██║██║██║     ██╔══╝  ██║  ██║██╔══██╗
	██████╔╝██║╚██████╗███████╗██████╔╝██████╔╝
	╚═════╝ ╚═╝ ╚═════╝╚══════╝╚═════╝ ╚═════╝
			
`)
	printConfiguration()
}

func Execute() {
	flagsConfig := config.Config{}
	flag.StringVar(&flagsConfig.RespServer.Addr, "host", "0.0.0.0", "host for the DiceDB server")

	flag.IntVar(&flagsConfig.RespServer.Port, "port", 7379, "port for the DiceDB server")

	flag.IntVar(&flagsConfig.HTTP.Port, "http-port", 8082, "port for accepting requets over HTTP")
	flag.BoolVar(&flagsConfig.HTTP.Enabled, "enable-http", false, "enable DiceDB to listen, accept, and process HTTP")

	flag.IntVar(&flagsConfig.WebSocket.Port, "websocket-port", 8379, "port for accepting requets over WebSocket")
	flag.BoolVar(&flagsConfig.WebSocket.Enabled, "enable-websocket", false, "enable DiceDB to listen, accept, and process WebSocket")

	flag.IntVar(&flagsConfig.Performance.NumShards, "num-shards", -1, "number shards to create. defaults to number of cores")

	flag.BoolVar(&flagsConfig.Performance.EnableWatch, "enable-watch", false, "enable support for .WATCH commands and real-time reactivity")
	flag.BoolVar(&flagsConfig.Performance.EnableProfiling, "enable-profiling", false, "enable profiling and capture critical metrics and traces in .prof files")

	flag.StringVar(&flagsConfig.Logging.LogLevel, "log-level", "info", "log level, values: info, debug")
	flag.StringVar(&config.DiceConfig.Logging.LogDir, "log-dir", "/tmp/dicedb", "log directory path")

	flag.BoolVar(&flagsConfig.Persistence.Enabled, "enable-persistence", false, "enable write-ahead logging")
	flag.BoolVar(&flagsConfig.Persistence.RestoreFromWAL, "restore-wal", false, "restore the database from the WAL files")
	flag.StringVar(&flagsConfig.Persistence.WALEngine, "wal-engine", "null", "wal engine to use, values: sqlite, aof")

	flag.StringVar(&flagsConfig.Auth.Password, "requirepass", utils.EmptyStr, "enable authentication for the default user")
	flag.StringVar(&config.CustomConfigFilePath, "o", config.CustomConfigFilePath, "dir path to create the flagsConfig file")
	flag.StringVar(&config.CustomConfigDirPath, "c", config.CustomConfigDirPath, "file path of the config file")

	flag.IntVar(&flagsConfig.Memory.KeysLimit, "keys-limit", config.DefaultKeysLimit, "keys limit for the DiceDB server. "+
		"This flag controls the number of keys each shard holds at startup. You can multiply this number with the "+
		"total number of shard threads to estimate how much memory will be required at system start up.")
	flag.Float64Var(&flagsConfig.Memory.EvictionRatio, "eviction-ratio", 0.9, "ratio of keys to evict when the "+
		"keys limit is reached")

	flag.Usage = func() {
		color.Set(color.FgYellow)
		fmt.Println("Usage: ./dicedb [options] [config-file]")
		color.Unset()

		color.Set(color.FgGreen)
		fmt.Println("Options:")
		color.Unset()

		color.Set(color.FgCyan)
		fmt.Println("  -v, --version          Show the version of DiceDB")
		fmt.Println("  -h, --help             Show this help message")
		fmt.Println("  -host                  Host for the DiceDB server (default: \"0.0.0.0\")")
		fmt.Println("  -port                  Port for the DiceDB server (default: 7379)")
		fmt.Println("  -http-port             Port for accepting requests over HTTP (default: 8082)")
		fmt.Println("  -enable-http           Enable DiceDB to listen, accept, and process HTTP (default: false)")
		fmt.Println("  -websocket-port        Port for accepting requests over WebSocket (default: 8379)")
		fmt.Println("  -enable-websocket      Enable DiceDB to listen, accept, and process WebSocket (default: false)")
		fmt.Println("  -num-shards            Number of shards to create. Defaults to number of cores (default: -1)")
		fmt.Println("  -enable-watch          Enable support for .WATCH commands and real-time reactivity (default: false)")
		fmt.Println("  -enable-profiling      Enable profiling and capture critical metrics and traces in .prof files (default: false)")
		fmt.Println("  -log-level             Log level, values: info, debug (default: \"info\")")
		fmt.Println("  -log-dir               Log directory path (default: \"/tmp/dicedb\")")
		fmt.Println("  -enable-persistence    Enable write-ahead logging (default: false)")
		fmt.Println("  -restore-wal           Restore the database from the WAL files (default: false)")
		fmt.Println("  -wal-engine            WAL engine to use, values: sqlite, aof (default: \"null\")")
		fmt.Println("  -requirepass           Enable authentication for the default user (default: \"\")")
		fmt.Println("  -o                     Directory path to create the config file (default: \"\")")
		fmt.Println("  -c                     File path of the config file (default: \"\")")
		fmt.Println("  -keys-limit            Keys limit for the DiceDB server (default: 200000000)")
		fmt.Println("  -eviction-ratio        Ratio of keys to evict when the keys limit is reached (default: 0.9)")
		color.Unset()
		os.Exit(0)
	}

	flag.Parse()

	if len(os.Args) > 2 {
		switch os.Args[1] {
		case "-v", "--version":
			fmt.Println("dicedb version", config.DiceDBVersion)
			os.Exit(0)

		case "-":
			parser := config.NewConfigParser()
			if err := parser.ParseFromStdin(); err != nil {
				log.Fatal(err)
			}
			if err := parser.Loadconfig(config.DiceConfig); err != nil {
				log.Fatal(err)
			}
			fmt.Println(config.DiceConfig.Version)
		case "-o", "--output":
			if len(os.Args) < 3 {
				log.Fatal("Output file path not provided")
			} else {
				dirPath := os.Args[2]
				if dirPath == "" {
					log.Fatal("Output file path not provided")
				}

				info, err := os.Stat(dirPath)
				switch {
				case os.IsNotExist(err):
					log.Fatal("Output file path does not exist")
				case err != nil:
					log.Fatalf("Error checking output file path: %v", err)
				case !info.IsDir():
					log.Fatal("Output file path is not a directory")
				}

				filePath := filepath.Join(dirPath, config.DefaultConfigName)
				if _, err := os.Stat(filePath); err == nil {
					slog.Warn("Config file already exists at the specified path", slog.String("path", filePath), slog.String("action", "skipping file creation"))
					return
				}
				if err := config.CreateConfigFile(filePath); err != nil {
					log.Fatal(err)
				}

				config.MergeFlags(&flagsConfig)
				render()
			}
		case "-c", "--config":
			if len(os.Args) >= 3 {
				filePath := os.Args[2]
				if filePath == "" {
					log.Fatal("Error: Config file path not provided")
				}

				info, err := os.Stat(filePath)
				switch {
				case os.IsNotExist(err):
					log.Fatalf("Config file does not exist: %s", filePath)
				case err != nil:
					log.Fatalf("Unable to check config file: %v", err)
				}

				if info.IsDir() {
					log.Fatalf("Config file path points to a directory: %s", filePath)
				}

				if !strings.HasSuffix(filePath, ".conf") {
					log.Fatalf("Config file must have a .conf extension: %s", filePath)
				}

				parser := config.NewConfigParser()
				if err := parser.ParseFromFile(filePath); err != nil {
					log.Fatal(err)
				}
				if err := parser.Loadconfig(config.DiceConfig); err != nil {
					log.Fatal(err)
				}

				config.MergeFlags(&flagsConfig)
				render()
			} else {
				log.Fatal("Config file path not provided")
			}
		default:
			defaultConfig(&flagsConfig)
		}
	}

	defaultConfig(&flagsConfig)
}

func defaultConfig(flags *config.Config) {
	if err := config.CreateConfigFile(filepath.Join(config.DefaultConfigDir, config.DefaultConfigName)); err != nil {
		log.Fatal(err)
	}

	config.MergeFlags(flags)
	render()
}
