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

type configEntry struct {
	Key   string
	Value interface{}
}

var configTable = []configEntry{}

// configuration function used to add configuration values to the print table at the startup.
// add entry to this function to add a new row in the startup configuration table.
func configuration() {
	// Add the version of the DiceDB to the configuration table
	addEntry("Version", config.DiceDBVersion)

	// Add the port number on which DiceDB is running to the configuration table
	addEntry("Port", config.DiceConfig.AsyncServer.Port)

	// Add whether multi-threading is enabled to the configuration table
	addEntry("Multi Threading Enabled", config.DiceConfig.Performance.EnableMultiThreading)

	// Add the number of CPU cores available on the machine to the configuration table
	addEntry("Cores", runtime.NumCPU())

	// Conditionally add the number of shards to be used for DiceDB to the configuration table
	if config.DiceConfig.Performance.EnableMultiThreading {
		if config.DiceConfig.Performance.NumShards > 0 {
			configTable = append(configTable, configEntry{"Shards", config.DiceConfig.Performance.NumShards})
		} else {
			configTable = append(configTable, configEntry{"Shards", runtime.NumCPU()})
		}
	} else {
		configTable = append(configTable, configEntry{"Shards", 1})
	}

	// Add whether the watch feature is enabled to the configuration table
	addEntry("Watch Enabled", config.DiceConfig.Performance.EnableWatch)

	// Add whether the watch feature is enabled to the configuration table
	addEntry("HTTP Enabled", config.DiceConfig.HTTP.Enabled)

	// Add whether the watch feature is enabled to the configuration table
	addEntry("Websocket Enabled", config.DiceConfig.WebSocket.Enabled)
}

func addEntry(k string, v interface{}) {
	configTable = append(configTable, configEntry{k, v})
}

// printConfigTable prints key-value pairs in a vertical table format.
func render() {
	color.Set(color.FgHiRed)
	fmt.Print(`
	██████╗ ██╗ ██████╗███████╗██████╗ ██████╗ 
	██╔══██╗██║██╔════╝██╔════╝██╔══██╗██╔══██╗
	██║  ██║██║██║     █████╗  ██║  ██║██████╔╝
	██║  ██║██║██║     ██╔══╝  ██║  ██║██╔══██╗
	██████╔╝██║╚██████╗███████╗██████╔╝██████╔╝
	╚═════╝ ╚═╝ ╚═════╝╚══════╝╚═════╝ ╚═════╝
			
	`)
	color.Unset()
	renderConfigTable()
}

func renderConfigTable() {
	configuration()
	color.Set(color.FgGreen)
	// Find the longest key to align the values properly
	// Default value length for alignment
	// Create the table header and separator line
	// 7 is for spacing and pipes
	// Print each configuration key-value pair without row lines
	// Final bottom line
	maxKeyLength := 0
	maxValueLength := 20
	for _, entry := range configTable {
		if len(entry.Key) > maxKeyLength {
			maxKeyLength = len(entry.Key)
		}
		if len(fmt.Sprintf("%v", entry.Value)) > maxValueLength {
			maxValueLength = len(fmt.Sprintf("%v", entry.Value))
		}
	}

	color.Set(color.FgGreen)
	fmt.Println()
	totalWidth := maxKeyLength + maxValueLength + 7
	fmt.Println(strings.Repeat("-", totalWidth))
	fmt.Printf("| %-*s | %-*s |\n", maxKeyLength, "Configuration", maxValueLength, "Value")
	fmt.Println(strings.Repeat("-", totalWidth))

	for _, entry := range configTable {
		fmt.Printf("| %-*s | %-20v |\n", maxKeyLength, entry.Key, entry.Value)
	}

	fmt.Println(strings.Repeat("-", totalWidth))
	fmt.Println()
	color.Unset()
}

func Execute() {
	flagsConfig := config.Config{}
	flag.StringVar(&flagsConfig.AsyncServer.Addr, "host", "0.0.0.0", "host for the DiceDB server")

	flag.IntVar(&flagsConfig.AsyncServer.Port, "port", 7379, "port for the DiceDB server")

	flag.IntVar(&flagsConfig.HTTP.Port, "http-port", 7380, "port for accepting requets over HTTP")
	flag.BoolVar(&flagsConfig.HTTP.Enabled, "enable-http", false, "enable DiceDB to listen, accept, and process HTTP")

	flag.IntVar(&flagsConfig.WebSocket.Port, "websocket-port", 7381, "port for accepting requets over WebSocket")
	flag.BoolVar(&flagsConfig.WebSocket.Enabled, "enable-websocket", false, "enable DiceDB to listen, accept, and process WebSocket")

	flag.BoolVar(&flagsConfig.Performance.EnableMultiThreading, "enable-multithreading", false, "enable multithreading execution and leverage multiple CPU cores")
	flag.IntVar(&flagsConfig.Performance.NumShards, "num-shards", -1, "number shards to create. defaults to number of cores")

	flag.BoolVar(&flagsConfig.Performance.EnableWatch, "enable-watch", false, "enable support for .WATCH commands and real-time reactivity")
	flag.BoolVar(&flagsConfig.Performance.EnableProfiling, "enable-profiling", false, "enable profiling and capture critical metrics and traces in .prof files")

	flag.StringVar(&flagsConfig.Logging.LogLevel, "log-level", "info", "log level, values: info, debug")
	flag.StringVar(&config.LogDir, "log-dir", "/tmp/dicedb", "log directory path")

	flag.BoolVar(&flagsConfig.Persistence.EnableWAL, "enable-wal", false, "enable write-ahead logging")
	flag.BoolVar(&flagsConfig.Persistence.RestoreFromWAL, "restore-wal", false, "restore the database from the WAL files")
	flag.StringVar(&flagsConfig.Persistence.WALEngine, "wal-engine", "null", "wal engine to use, values: sqlite, aof")

	flag.StringVar(&flagsConfig.Auth.Password, "requirepass", utils.EmptyStr, "enable authentication for the default user")
	flag.StringVar(&config.CustomConfigFilePath, "o", config.CustomConfigFilePath, "dir path to create the flagsConfig file")
	flag.StringVar(&config.FileLocation, "c", config.FileLocation, "file path of the config file")

	flag.IntVar(&flagsConfig.Memory.KeysLimit, "keys-limit", config.DefaultKeysLimit, "keys limit for the DiceDB server. "+
		"This flag controls the number of keys each shard holds at startup. You can multiply this number with the "+
		"total number of shard threads to estimate how much memory will be required at system start up.")
	flag.Float64Var(&flagsConfig.Memory.EvictionRatio, "eviction-ratio", 0.1, "ratio of keys to evict when the "+
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
		fmt.Println("  -http-port             Port for accepting requests over HTTP (default: 7380)")
		fmt.Println("  -enable-http           Enable DiceDB to listen, accept, and process HTTP (default: false)")
		fmt.Println("  -websocket-port        Port for accepting requests over WebSocket (default: 7381)")
		fmt.Println("  -enable-websocket      Enable DiceDB to listen, accept, and process WebSocket (default: false)")
		fmt.Println("  -enable-multithreading Enable multithreading execution and leverage multiple CPU cores (default: false)")
		fmt.Println("  -num-shards            Number of shards to create. Defaults to number of cores (default: -1)")
		fmt.Println("  -enable-watch          Enable support for .WATCH commands and real-time reactivity (default: false)")
		fmt.Println("  -enable-profiling      Enable profiling and capture critical metrics and traces in .prof files (default: false)")
		fmt.Println("  -log-level             Log level, values: info, debug (default: \"info\")")
		fmt.Println("  -log-dir               Log directory path (default: \"/tmp/dicedb\")")
		fmt.Println("  -enable-wal            Enable write-ahead logging (default: false)")
		fmt.Println("  -restore-wal           Restore the database from the WAL files (default: false)")
		fmt.Println("  -wal-engine            WAL engine to use, values: sqlite, aof (default: \"null\")")
		fmt.Println("  -requirepass           Enable authentication for the default user (default: \"\")")
		fmt.Println("  -o                     Directory path to create the config file (default: \"\")")
		fmt.Println("  -c                     File path of the config file (default: \"\")")
		fmt.Println("  -keys-limit            Keys limit for the DiceDB server (default: 0)")
		fmt.Println("  -eviction-ratio        Ratio of keys to evict when the keys limit is reached (default: 0.1)")
		color.Unset()
		os.Exit(0)
	}

	flag.Parse()
	config.MergeFlags(&flagsConfig)
	if len(os.Args) < 2 {
		if err := config.CreateConfigFile(filepath.Join(config.DefaultConfigDir, "dicedb.conf")); err != nil {
			log.Fatal(err)
		}
		render()
		return
	}

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

			filePath := filepath.Join(dirPath, "dicedb.conf")
			if _, err := os.Stat(filePath); err == nil {
				slog.Warn("Config file already exists at the specified path", slog.String("path", filePath), slog.String("action", "skipping file creation"))
				return
			}
			if err := config.CreateConfigFile(filePath); err != nil {
				log.Fatal(err)
			}
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
			render()
		} else {
			log.Fatal("Config file path not provided")
		}
	default:
		render()
	}
}
