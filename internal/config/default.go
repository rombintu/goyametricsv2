package config

import (
	"os"
	"strconv"
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
	defaultRateLimit      = 0

	// Server
	hintListen        = "Server address"
	hintStorageDriver = "Storage driver"
	hintEnvMode       = "Enviriment server mode"
	hintStoreInterval = "Interval between saves"
	hintStoragePath   = "Path to store data"
	hintStorageURL    = "URL or Plain creds to database"
	hintRestoreFlag   = "Restore data from store?"

	// Iter 14
	hintHashKey    = "Key for hash"
	defaultHashKey = ""

	// Agent
	hintServerURL      = hintListen
	hintReportInterval = "Report interval"
	hintPollInterval   = "Poll interval"
	hintRateLimit      = "Rate limit. 0 - unlimited"
)

type DatabaseConfig struct {
	User string `yaml:"db_user"`
	Pass string `yaml:"db_pass"`
	Host string `yaml:"db_host" env-default:"localhost"`
	Port string `yaml:"db_port" env-default:"5432"`
	Name string `yaml:"db_name" env-default:"metrics"`
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
