package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/dicedb/dice/internal/server/utils"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
)

const (
	DefaultHost           string = "0.0.0.0"
	DefaultPort           int    = 7379
	DefaultConfigName     string = "dice.toml"
	DefaultConfigFilePath string = "./"

	EvictSimpleFirst   = "simple-first"
	EvictAllKeysRandom = "allkeys-random"
	EvictAllKeysLRU    = "allkeys-lru"
	EvictAllKeysLFU    = "allkeys-lfu"
)

var (
	Host = DefaultHost
	Port = DefaultPort

	EnableHTTP = true
	HTTPPort   = 8082
	// if RequirePass is set to an empty string, no authentication is required
	RequirePass = utils.EmptyStr

	CustomConfigFilePath = utils.EmptyStr
	ConfigFileLocation   = utils.EmptyStr

	InitConfigCmd = false
)

type Config struct {
	Server struct {
		Addr                   string        `mapstructure:"addr"`
		Port                   int           `mapstructure:"port"`
		KeepAlive              int32         `mapstructure:"keepalive"`
		Timeout                int32         `mapstructure:"timeout"`
		MaxConn                int32         `mapstructure:"max-conn"`
		ShardCronFrequency     time.Duration `mapstructure:"shardcronfrequency"`
		MultiplexerPollTimeout time.Duration `mapstructure:"servermultiplexerpolltimeout"`
		MaxClients             int           `mapstructure:"maxclients"`
		MaxMemory              int64         `mapstructure:"maxmemory"`
		EvictionPolicy         string        `mapstructure:"evictionpolicy"`
		EvictionRatio          float64       `mapstructure:"evictionratio"`
		KeysLimit              int           `mapstructure:"keyslimit"`
		AOFFile                string        `mapstructure:"aoffile"`
		PersistenceEnabled     bool          `mapstructure:"persistenceenabled"`
		WriteAOFOnCleanup      bool          `mapstructure:"writeaofoncleanup"`
		LFULogFactor           int           `mapstructure:"lfulogfactor"`
		LogLevel               string        `mapstructure:"loglevel"`
		PrettyPrintLogs        bool          `mapstructure:"prettyprintlogs"`
	} `mapstructure:"server"`
	Auth struct {
		UserName string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"auth"`
	Network struct {
		IOBufferLength    int `mapstructure:"iobufferlength"`
		IOBufferLengthMAX int `mapstructure:"iobufferlengthmax"`
	} `mapstructure:"network"`
}

// Default configurations for internal use
var baseConfig = Config{
	Server: struct {
		Addr                   string        `mapstructure:"addr"`
		Port                   int           `mapstructure:"port"`
		KeepAlive              int32         `mapstructure:"keepalive"`
		Timeout                int32         `mapstructure:"timeout"`
		MaxConn                int32         `mapstructure:"max-conn"`
		ShardCronFrequency     time.Duration `mapstructure:"shardcronfrequency"`
		MultiplexerPollTimeout time.Duration `mapstructure:"servermultiplexerpolltimeout"`
		MaxClients             int           `mapstructure:"maxclients"`
		MaxMemory              int64         `mapstructure:"maxmemory"`
		EvictionPolicy         string        `mapstructure:"evictionpolicy"`
		EvictionRatio          float64       `mapstructure:"evictionratio"`
		KeysLimit              int           `mapstructure:"keyslimit"`
		AOFFile                string        `mapstructure:"aoffile"`
		PersistenceEnabled     bool          `mapstructure:"persistenceenabled"`
		WriteAOFOnCleanup      bool          `mapstructure:"writeaofoncleanup"`
		LFULogFactor           int           `mapstructure:"lfulogfactor"`
		LogLevel               string        `mapstructure:"loglevel"`
		PrettyPrintLogs        bool          `mapstructure:"prettyprintlogs"`
	}{
		Addr:                   DefaultHost,
		Port:                   DefaultPort,
		KeepAlive:              int32(300),
		Timeout:                int32(300),
		MaxConn:                int32(0),
		ShardCronFrequency:     1 * time.Second,
		MultiplexerPollTimeout: 100 * time.Millisecond,
		MaxClients:             20000,
		MaxMemory:              0,
		EvictionPolicy:         EvictAllKeysLFU,
		EvictionRatio:          0.40,
		KeysLimit:              10000,
		AOFFile:                "./dice-master.aof",
		PersistenceEnabled:     true,
		WriteAOFOnCleanup:      true,
		LFULogFactor:           10,
		LogLevel:               "info",
		PrettyPrintLogs:        false,
	},
	Auth: struct {
		UserName string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	}{
		UserName: "dice",
		Password: RequirePass,
	},
	Network: struct {
		IOBufferLength    int `mapstructure:"iobufferlength"`
		IOBufferLengthMAX int `mapstructure:"iobufferlengthmax"`
	}{
		IOBufferLength:    512,
		IOBufferLengthMAX: 50 * 1024,
	},
}

var defaultConfig Config

func init() {
	config := baseConfig
	env := os.Getenv("DICE_ENV")
	switch env {
	case "dev":
		config.Server.LogLevel = "debug"
		config.Server.PrettyPrintLogs = true
	default:
	}
	logLevel := os.Getenv("DICE_LOG_LEVEL")
	if logLevel != "" {
		config.Server.LogLevel = logLevel
	}
	defaultConfig = config
}

// DiceConfig is the global configuration object for dice
var DiceConfig *Config = &defaultConfig

func SetupConfig() {
	if InitConfigCmd {
		ConfigFileLocation = getConfigPath()
		createConfigFile(ConfigFileLocation)
		return
	}

	// Check if both -o and -c flags are set
	if areBothFlagsSet() {
		slog.Error("Both -o and -c flags are set. Please use only one flag.")
		return
	}

	// Check if -o flag is set
	if CustomConfigFilePath != utils.EmptyStr && isValidDirPath() {
		createConfigFile(filepath.Join(CustomConfigFilePath, DefaultConfigName))
		return
	}

	// Check if -c flag is set
	if ConfigFileLocation != utils.EmptyStr || isConfigFilePresent() {
		setUpViperConfig(ConfigFileLocation)
		return
	}

	// If no flags are set, use default configurations with prioritizing command line flags
	mergeFlagsWithConfig()
}

func createConfigFile(configFilePath string) {
	if _, err := os.Stat(configFilePath); err == nil {
		slog.Warn("config file already exists", slog.String("path", configFilePath))
		setUpViperConfig(configFilePath)
		return
	}

	if err := writeConfigFile(configFilePath); err != nil {
		slog.Warn("starting DiceDB with default configurations.", slog.Any("error", err))
		return
	}

	setUpViperConfig(configFilePath)
	slog.Info("config file created at %s with default configurations", slog.Any("path", configFilePath))
}

func writeConfigFile(configFilePath string) error {
	dir := filepath.Dir(configFilePath)
	if _, err := os.Stat(dir); err != nil {
		return err
	}

	slog.Info("creating default config file at %s", slog.Any("path", configFilePath))
	file, err := os.Create(configFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	err = encoder.Encode(defaultConfig)
	return err
}

func isValidDirPath() bool {
	if CustomConfigFilePath == utils.EmptyStr {
		return false
	}

	info, err := os.Stat(CustomConfigFilePath)
	if os.IsNotExist(err) || err != nil {
		return false
	}

	if !info.IsDir() {
		return false
	}
	return true
}

// This function checks if both -o and -c flags are set or not
func areBothFlagsSet() bool {
	return ConfigFileLocation != utils.EmptyStr && CustomConfigFilePath != utils.EmptyStr
}

func setUpViperConfig(configFilePath string) {
	// if configFilepath has config file then that file name will be viper.SetConfigName
	if configFilePath != filepath.Join(DefaultConfigFilePath, DefaultConfigName) {
		viper.SetConfigName(strings.Split(filepath.Base(configFilePath), ".")[0])
	} else {
		viper.SetConfigName("dice")
	}

	if configFilePath == utils.EmptyStr {
		viper.AddConfigPath(DefaultConfigFilePath)
	} else {
		viper.AddConfigPath(filepath.Dir(configFilePath))
	}

	viper.SetConfigType("toml")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slog.Warn("config file not found. Using default configurations.")
			return
		}
		slog.Error("Error reading config file", slog.Any("error", err))
	}

	if err := viper.Unmarshal(&DiceConfig); err != nil {
		slog.Error("Error unmarshalling config file", slog.Any("error", err))
		slog.Warn("starting DiceDB with default configurations.")
		return
	}

	// override default configurations with command line flags
	mergeFlagsWithConfig()

	slog.Info("configurations loaded successfully.")
}

func mergeFlagsWithConfig() {
	if RequirePass != utils.EmptyStr {
		DiceConfig.Auth.Password = RequirePass
	}

	if Host != DefaultHost {
		DiceConfig.Server.Addr = Host
	}

	if Port != DefaultPort {
		DiceConfig.Server.Port = Port
	}
}

// This function checks if the config file is present or not at ConfigFileLocation
func isConfigFilePresent() bool {
	_, err := os.Stat(ConfigFileLocation)
	return err == nil
}

// This function returns the config file path based on the OS
func getConfigPath() string {
	switch runtime.GOOS {
	case "windows":
		ConfigFileLocation = filepath.Join("C:", "ProgramData", "dice", DefaultConfigName)
	case "darwin", "linux":
		ConfigFileLocation = filepath.Join(string(filepath.Separator), "etc", "dice", DefaultConfigName)
	default:
		// Default to current directory if OS is unknown
		ConfigFileLocation = filepath.Join(".", DefaultConfigName)
	}
	return ConfigFileLocation
}

// This function is only used for testing purposes
func ResetConfig() {
	DiceConfig = &defaultConfig
}
