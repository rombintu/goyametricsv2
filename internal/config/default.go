// Package config DefaultConfig
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

	// Inter 21
	defaultPrivateKeyFile = ""
	defaultPubkeyFile     = ""
	hintPrivateKeyFile    = "Path to private key of server"
	hintPubkeyFile        = "Path to public key of server"
)

type DatabaseConfig struct {
	User string `yaml:"db_user"`
	Pass string `yaml:"db_pass"`
	Host string `yaml:"db_host" env-default:"localhost"`
	Port string `yaml:"db_port" env-default:"5432"`
	Name string `yaml:"db_name" env-default:"metrics"`
}

// tryLoadFromEnv attempts to load a configuration value from an environment variable.
// If the environment variable is not set, it returns the value from the flags.
//
// Parameters:
// - key: The name of the environment variable.
// - fromFlags: The default value to use if the environment variable is not set.
//
// Returns:
// - The value from the environment variable if set, otherwise the value from the flags.
func tryLoadFromEnv(key, fromFlags string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fromFlags
	} else {
		return value
	}
}

// tryLoadFromEnvInt64 attempts to load an integer configuration value from an environment variable.
// If the environment variable is not set or cannot be parsed, it returns the value from the flags.
//
// Parameters:
// - key: The name of the environment variable.
// - fromFlags: The default value to use if the environment variable is not set or cannot be parsed.
//
// Returns:
// - The parsed value from the environment variable if set and valid, otherwise the value from the flags.
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

// tryLoadFromEnvBool attempts to load a boolean configuration value from an environment variable.
// If the environment variable is not set or cannot be parsed, it returns the value from the flags.
//
// Parameters:
// - key: The name of the environment variable.
// - fromFlags: The default value to use if the environment variable is not set or cannot be parsed.
//
// Returns:
// - The parsed value from the environment variable if set and valid, otherwise the value from the flags.
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
