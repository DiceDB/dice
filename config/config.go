package config

import (
	"errors"
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/dicedb/dice/internal/server/utils"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	DiceDBVersion string = "0.0.5"

	DefaultHost           string = "0.0.0.0"
	DefaultPort           int    = 7379
	DefaultConfigName     string = "dice.toml"
	DefaultConfigFilePath string = "./"

	EvictSimpleFirst   = "simple-first"
	EvictAllKeysRandom = "allkeys-random"
	EvictAllKeysLRU    = "allkeys-lru"
	EvictAllKeysLFU    = "allkeys-lfu"

	DefaultKeysLimit int = 200000000
)

var (
	Host = DefaultHost
	Port = DefaultPort

	EnableMultiThreading = false
	EnableHTTP           = true
	HTTPPort             = 8082

	EnableWebsocket     = true
	WebsocketPort       = 8379
	NumShards       int = -1

	// if RequirePass is set to an empty string, no authentication is required
	RequirePass = utils.EmptyStr

	CustomConfigFilePath = utils.EmptyStr
	FileLocation         = utils.EmptyStr

	InitConfigCmd = false

	KeysLimit = DefaultKeysLimit

	EnableProfiling = false

	EnableWatch = true
)

type Config struct {
	Version     string `mapstructure:"version"`
	InstanceID  string `mapstructure:"instance_id"`
	AsyncServer struct {
		Addr      string `mapstructure:"addr"`
		Port      int    `mapstructure:"port"`
		KeepAlive int32  `mapstructure:"keepalive"`
		Timeout   int32  `mapstructure:"timeout"`
		MaxConn   int32  `mapstructure:"max-conn"`
	} `mapstructure:"asyncserver"`

	HTTP struct {
		Enabled bool `mapstructure:"enabled"`
		Port    int  `mapstructure:"port"`
	} `mapstructure:"http"`

	WebSocket struct {
		Enabled                 bool          `mapstructure:"enabled"`
		Port                    int           `mapstructure:"port"`
		MaxWriteResponseRetries int           `mapstructure:"maxwriteresponseretries"`
		WriteResponseTimeout    time.Duration `mapstructure:"writeresponsetimeout"`
	} `mapstructure:"websocket"`

	Performance struct {
		WatchChanBufSize       int           `mapstructure:"watchchanbufsize"`
		ShardCronFrequency     time.Duration `mapstructure:"shardcronfrequency"`
		MultiplexerPollTimeout time.Duration `mapstructure:"servermultiplexerpolltimeout"`
		MaxClients             int32         `mapstructure:"maxclients"`
		EnableMultiThreading   bool          `mapstructure:"enablemultithreading"`
		StoreMapInitSize       int           `mapstructure:"storemapinitsize"`
		AdhocReqChanBufSize    int           `mapstructure:"adhocreqchanbufsize"`
		EnableProfiling        bool          `mapstructure:"profiling"`
	} `mapstructure:"performance"`

	Memory struct {
		MaxMemory      int64   `mapstructure:"maxmemory"`
		EvictionPolicy string  `mapstructure:"evictionpolicy"`
		EvictionRatio  float64 `mapstructure:"evictionratio"`
		KeysLimit      int     `mapstructure:"keyslimit"`
		LFULogFactor   int     `mapstructure:"lfulogfactor"`
	} `mapstructure:"memory"`

	Persistence struct {
		AOFFile            string `mapstructure:"aoffile"`
		PersistenceEnabled bool   `mapstructure:"persistenceenabled"`
		WriteAOFOnCleanup  bool   `mapstructure:"writeaofoncleanup"`
	} `mapstructure:"persistence"`

	Logging struct {
		LogLevel        string `mapstructure:"loglevel"`
		PrettyPrintLogs bool   `mapstructure:"prettyprintlogs"`
	} `mapstructure:"logging"`

	Auth struct {
		UserName string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"auth"`

	Network struct {
		IOBufferLength    int `mapstructure:"iobufferlength"`
		IOBufferLengthMAX int `mapstructure:"iobufferlengthmax"`
	} `mapstructure:"network"`

	NumShards int `mapstructure:"num_shards"`
}

// Default configurations for internal use
var baseConfig = Config{
	Version: DiceDBVersion,
	AsyncServer: struct {
		Addr      string `mapstructure:"addr"`
		Port      int    `mapstructure:"port"`
		KeepAlive int32  `mapstructure:"keepalive"`
		Timeout   int32  `mapstructure:"timeout"`
		MaxConn   int32  `mapstructure:"max-conn"`
	}{
		Addr:      DefaultHost,
		Port:      DefaultPort,
		KeepAlive: int32(300),
		Timeout:   int32(300),
		MaxConn:   int32(0),
	},
	HTTP: struct {
		Enabled bool `mapstructure:"enabled"`
		Port    int  `mapstructure:"port"`
	}{
		Enabled: EnableHTTP,
		Port:    HTTPPort,
	},
	WebSocket: struct {
		Enabled                 bool          `mapstructure:"enabled"`
		Port                    int           `mapstructure:"port"`
		MaxWriteResponseRetries int           `mapstructure:"maxwriteresponseretries"`
		WriteResponseTimeout    time.Duration `mapstructure:"writeresponsetimeout"`
	}{
		Enabled:                 EnableWebsocket,
		Port:                    WebsocketPort,
		MaxWriteResponseRetries: 3,
		WriteResponseTimeout:    10 * time.Second,
	},
	Performance: struct {
		WatchChanBufSize       int           `mapstructure:"watchchanbufsize"`
		ShardCronFrequency     time.Duration `mapstructure:"shardcronfrequency"`
		MultiplexerPollTimeout time.Duration `mapstructure:"servermultiplexerpolltimeout"`
		MaxClients             int32         `mapstructure:"maxclients"`
		EnableMultiThreading   bool          `mapstructure:"enablemultithreading"`
		StoreMapInitSize       int           `mapstructure:"storemapinitsize"`
		AdhocReqChanBufSize    int           `mapstructure:"adhocreqchanbufsize"`
		EnableProfiling        bool          `mapstructure:"profiling"`
	}{
		WatchChanBufSize:       20000,
		ShardCronFrequency:     1 * time.Second,
		MultiplexerPollTimeout: 100 * time.Millisecond,
		MaxClients:             int32(20000),
		EnableMultiThreading:   false,
		StoreMapInitSize:       1024000,
		AdhocReqChanBufSize:    20, // assuming we wouldn't have more than 20 adhoc requests being sent at a time.
	},
	Memory: struct {
		MaxMemory      int64   `mapstructure:"maxmemory"`
		EvictionPolicy string  `mapstructure:"evictionpolicy"`
		EvictionRatio  float64 `mapstructure:"evictionratio"`
		KeysLimit      int     `mapstructure:"keyslimit"`
		LFULogFactor   int     `mapstructure:"lfulogfactor"`
	}{
		MaxMemory:      0,
		EvictionPolicy: EvictAllKeysLFU,
		EvictionRatio:  0.9,
		KeysLimit:      DefaultKeysLimit,
		LFULogFactor:   10,
	},
	Persistence: struct {
		AOFFile            string `mapstructure:"aoffile"`
		PersistenceEnabled bool   `mapstructure:"persistenceenabled"`
		WriteAOFOnCleanup  bool   `mapstructure:"writeaofoncleanup"`
	}{
		PersistenceEnabled: true,
		AOFFile:            "./dice-master.aof",
		WriteAOFOnCleanup:  false,
	},
	Logging: struct {
		LogLevel        string `mapstructure:"loglevel"`
		PrettyPrintLogs bool   `mapstructure:"prettyprintlogs"`
	}{
		LogLevel:        "info",
		PrettyPrintLogs: true,
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

func setupFlags() {
	flag.StringVar(&Host, "host", "0.0.0.0", "host for the dicedb server")
	flag.IntVar(&Port, "port", 7379, "port for the dicedb server")
	flag.BoolVar(&EnableHTTP, "enable-http", true, "run server in HTTP mode as well")
	flag.BoolVar(&EnableMultiThreading, "enable-multithreading", false, "run server in multithreading mode")
	flag.IntVar(&HTTPPort, "http-port", 8082, "HTTP port for the dicedb server")
	flag.IntVar(&WebsocketPort, "websocket-port", 8379, "Websocket port for the dicedb server")
	flag.StringVar(&RequirePass, "requirepass", RequirePass, "enable authentication for the default user")
	flag.StringVar(&CustomConfigFilePath, "o", CustomConfigFilePath, "dir path to create the config file")
	flag.StringVar(&FileLocation, "c", FileLocation, "file path of the config file")
	flag.BoolVar(&InitConfigCmd, "init-config", false, "initialize a new config file")
	flag.IntVar(&KeysLimit, "keys-limit", KeysLimit, "keys limit for the dicedb server. "+
		"This flag controls the number of keys each shard holds at startup. You can multiply this number with the "+
		"total number of shard threads to estimate how much memory will be required at system start up.")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
}

func init() {
	config := baseConfig
	config.Logging.PrettyPrintLogs = false
	config.Logging.LogLevel = "info"
	defaultConfig = config

	setupFlags()
}

// DiceConfig is the global configuration object for dice
var DiceConfig = &defaultConfig

func SetupConfig() {
	if InitConfigCmd {
		FileLocation = getConfigPath()
		createConfigFile(FileLocation)
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

	setUpViperConfig(FileLocation)
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
	slog.Info("config file created at", slog.Any("path", configFilePath))
}

func writeConfigFile(configFilePath string) error {
	dir := filepath.Dir(configFilePath)
	if _, err := os.Stat(dir); err != nil {
		return err
	}

	slog.Info("creating default config file at", slog.Any("path", configFilePath))
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
	return FileLocation != utils.EmptyStr && CustomConfigFilePath != utils.EmptyStr
}

func setUpViperConfig(configFilePath string) {
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
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			slog.Warn("config file not found. Using default configurations.")
			return
		}
		slog.Error("Error reading config file", slog.Any("error", err))
	}

	// override default configurations with command line flags
	viper.BindPFlag("Performance.EnableMultiThreading", pflag.Lookup("enable-multithreading"))
	viper.BindPFlag("Auth.Password", pflag.Lookup("requirepass"))
	viper.BindPFlag("AsyncServer.Addr", pflag.Lookup("host"))
	viper.BindPFlag("AsyncServer.Port", pflag.Lookup("port"))
	viper.BindPFlag("Memory.KeysLimit", pflag.Lookup("keys-limit"))

	if err := viper.Unmarshal(&DiceConfig); err != nil {
		slog.Error("Error unmarshalling config file", slog.Any("error", err))
		slog.Warn("starting DiceDB with default configurations.")
		return
	}

	slog.Info("configurations loaded successfully.")
}

// This function checks if the config file is present or not at default location or at -c flag location
func isConfigFilePresent() bool {
	// If -c flag is not set then look for config file in current directory use it
	if _, err := os.Stat(filepath.Join(".", DefaultConfigName)); FileLocation == utils.EmptyStr && err == nil {
		FileLocation = filepath.Join(".", DefaultConfigName)
		return true
	}

	// will be executed if -c flag is used
	_, err := os.Stat(FileLocation)

	return err == nil
}

// This function returns the config file path based on the OS
func getConfigPath() string {
	switch runtime.GOOS {
	case "windows":
		FileLocation = filepath.Join("C:", "ProgramData", "dice", DefaultConfigName)
	case "darwin", "linux":
		FileLocation = filepath.Join(string(filepath.Separator), "etc", "dice", DefaultConfigName)
	default:
		// Default to current directory if OS is unknown
		FileLocation = filepath.Join(".", DefaultConfigName)
	}
	return FileLocation
}

// ResetConfig resets the DiceConfig to default configurations. This function is only used for testing purposes
func ResetConfig() {
	DiceConfig = &defaultConfig
}
