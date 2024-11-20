package cli

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dicedb/dice/config"
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
	fmt.Print(`
	██████╗ ██╗ ██████╗███████╗██████╗ ██████╗ 
	██╔══██╗██║██╔════╝██╔════╝██╔══██╗██╔══██╗
	██║  ██║██║██║     █████╗  ██║  ██║██████╔╝
	██║  ██║██║██║     ██╔══╝  ██║  ██║██╔══██╗
	██████╔╝██║╚██████╗███████╗██████╔╝██████╔╝
	╚═════╝ ╚═╝ ╚═════╝╚══════╝╚═════╝ ╚═════╝
			
	`)
	configuration()

	// Find the longest key to align the values properly
	maxKeyLength := 0
	maxValueLength := 20 // Default value length for alignment
	for _, entry := range configTable {
		if len(entry.Key) > maxKeyLength {
			maxKeyLength = len(entry.Key)
		}
		if len(fmt.Sprintf("%v", entry.Value)) > maxValueLength {
			maxValueLength = len(fmt.Sprintf("%v", entry.Value))
		}
	}

	// Create the table header and separator line
	fmt.Println()
	totalWidth := maxKeyLength + maxValueLength + 7 // 7 is for spacing and pipes
	fmt.Println(strings.Repeat("-", totalWidth))
	fmt.Printf("| %-*s | %-*s |\n", maxKeyLength, "Configuration", maxValueLength, "Value")
	fmt.Println(strings.Repeat("-", totalWidth))

	// Print each configuration key-value pair without row lines
	for _, entry := range configTable {
		fmt.Printf("| %-*s | %-20v |\n", maxKeyLength, entry.Key, entry.Value)
	}

	// Final bottom line
	fmt.Println(strings.Repeat("-", totalWidth))
	fmt.Println()
}

func Execute() {
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

	case "-h", "--help":
		printUsage()
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
		fmt.Printf("Unknown option: %s\n", os.Args[1])
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`Usage: ./dicedb [/path/to/dice.conf] [options] [-]
	   ./dicedb - (read config from stdin) e.g. echo "version=1.0" | ./dicedb -
	   ./dicedb -v or --version
	   ./dicedb -h or --help`)
}
