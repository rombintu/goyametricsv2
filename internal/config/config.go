package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/rombintu/goyametricsv2/internal/storage"
)

const (
	// Server
	defaultListen        = "localhost:8080"
	defaultStorageDriver = "mem"
	defaultEnvMode       = "dev"
	defaultStoreInterval = 300
	defaultStoragePath   = "store.json"
	defaultRestoreFlag   = true
	// Agent
	defaultServerURL      = defaultListen
	defaultReportInterval = 2
	defaultPollInterval   = 10

	// Server
	hintListen        = "Server address"
	hintStorageDriver = "Storage driver"
	hintEnvMode       = "Enviriment server mode"
	hintStoreInterval = "Interval between saves"
	hintStoragePath   = "Path to store data"
	hintStorageURL    = "URL or Plain creds to database"
	hintRestoreFlag   = "Restore data from store?"

	// Agent
	hintServerURL      = hintListen
	hintReportInterval = "Report interval"
	hintPollInterval   = "Poll interval"
)

type DatabaseConfig struct {
	User string `yaml:"db_user"`
	Pass string `yaml:"db_pass"`
	Host string `yaml:"db_host" env-default:"localhost"`
	Port string `yaml:"db_port" env-default:"5432"`
	Name string `yaml:"db_name" env-default:"metrics"`
}

func (db *DatabaseConfig) ToPlainText() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s sslmode=disable",
		db.Host, db.User, db.Pass, db.Name,
	)
}

type ServerConfig struct {
	Listen        string `yaml:"Listen" env-default:"localhost:8080"`
	StorageDriver string `yaml:"StorageDriver" env-default:"mem"`
	EnvMode       string `yaml:"EnvMode" env-default:"dev"`

	// New
	StoreInterval int64  `yaml:"StoreInterval" env-default:"300"`
	StoragePath   string `yaml:"StoragePath" env-default:"store.json"`

	StorageURL string `yaml:"StorageURL"`

	RestoreFlag bool `yaml:"RestoreFlag" env-default:"true"`
	SyncMode    bool `yaml:"SyncMode" env-default:"false"`
}

type AgentConfig struct {
	Address        string `yaml:"address" env-default:"http://localhost:8080"`
	PollInterval   int64  `yaml:"pollInterval" env-default:"2"`
	ReportInterval int64  `yaml:"reportInterval" env-default:"10"`
	EnvMode        string `yaml:"EnvMode" env-default:"dev"`
}

func tryLoadFromEnv(key, fromFlags string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fromFlags
	} else {
		return value
	}
}

func tryLoadFromEnvInt64(key string, fromFlags int64) int64 {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fromFlags
	} else {
		parse64, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fromFlags
		} else {
			return parse64
		}
	}
}

func tryLoadFromEnvBool(key string, fromFlags bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fromFlags
	} else {
		parseBool, err := strconv.ParseBool(value)
		if err != nil {
			return fromFlags
		} else {
			return parseBool
		}
	}
}

func LoadServerConfig() ServerConfig {
	var config ServerConfig
	fromFlags := loadServerConfigFromFlags()
	config.Listen = tryLoadFromEnv("ADDRESS", fromFlags.Listen)
	// New args
	config.StoreInterval = tryLoadFromEnvInt64("STORE_INTERVAL", fromFlags.StoreInterval)
	config.StoragePath = tryLoadFromEnv("FILE_STORAGE_PATH", fromFlags.StoragePath)
	config.RestoreFlag = tryLoadFromEnvBool("RESTORE_FLAG", fromFlags.RestoreFlag)

	// increment 10
	config.StorageDriver = tryLoadFromEnv("STORAGE_DRIVER", fromFlags.StorageDriver)
	config.StorageURL = tryLoadFromEnv("DATABASE_DSN", fromFlags.StorageURL)
	// Change to sync mode
	if config.StoreInterval == 0 {
		config.SyncMode = true
	}

	if config.StorageURL != "" {
		config.StorageDriver = storage.PgxDriver
	} else if config.StoragePath != "" {
		config.StorageDriver = storage.FileDriver
	}

	return config
}

// Try load Server Config from flags
func loadServerConfigFromFlags() ServerConfig {
	var config ServerConfig
	a := flag.String("a", defaultListen, hintListen)
	s := flag.String("driver", defaultStorageDriver, hintStorageDriver)
	e := flag.String("env", defaultEnvMode, hintEnvMode)

	// New flags
	i := flag.Int64("i", defaultStoreInterval, hintStoreInterval)
	f := flag.String("f", defaultStoragePath, hintStoragePath)
	r := flag.Bool("r", defaultRestoreFlag, hintRestoreFlag)
	d := flag.String("d", "", hintStorageURL)
	flag.Parse()

	config.Listen = *a
	config.StorageDriver = *s
	config.EnvMode = *e

	// Parse new flags
	config.StoreInterval = *i
	config.StoragePath = *f
	config.RestoreFlag = *r

	// increment 10
	config.StorageURL = *d

	return config
}

func (c *ServerConfig) StoragePathAuto() bool {
	if c.StorageDriver != defaultStorageDriver && c.StorageURL == "" {
		dbConfig := loadDatabaseConfig()
		c.StorageURL = dbConfig.ToPlainText()
		return true
	}
	return false
}

// Load Agent Config from Environment, if any var empty - load from flags or set default
func LoadAgentConfig() AgentConfig {
	var config AgentConfig
	fromFlags := loadAgentConfigFromFlags()
	// Из тз нужно сделать такое ключевое слово, иначе не проходят тесты
	// ADDRESS отвечает за адрес эндпоинта HTTP-сервера.
	config.Address = tryLoadFromEnv("ADDRESS", fromFlags.Address)
	config.ReportInterval = tryLoadFromEnvInt64("REPORT_INTERVAL", fromFlags.ReportInterval)
	config.PollInterval = tryLoadFromEnvInt64("POLL_INTERVAL", fromFlags.PollInterval)

	return config
}

// Try load Server Config from flags
func loadAgentConfigFromFlags() AgentConfig {
	var config AgentConfig
	a := flag.String("a", defaultServerURL, hintServerURL)
	r := flag.Int64("r", defaultReportInterval, hintReportInterval)
	p := flag.Int64("p", defaultPollInterval, hintPollInterval)
	flag.Parse()

	config.Address = *a
	config.ReportInterval = *r
	config.PollInterval = *p

	return config
}

func loadDatabaseConfig() DatabaseConfig {
	var dbConfig DatabaseConfig
	dbConfig.Host = os.Getenv("DB_HOST")
	dbConfig.Port = os.Getenv("DB_PORT")
	dbConfig.User = os.Getenv("DB_USER")
	dbConfig.Pass = os.Getenv("DB_PASS")
	dbConfig.Name = os.Getenv("DB_NAME")
	return dbConfig
}
