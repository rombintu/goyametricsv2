// Package config DefaultConfig
package config

import (
	"os"
	"reflect"
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

	// Inter 22
	defaultPathConfig = ""
	hintPathConfig    = "Path to config file"
)

// tryLoadFromEnv attempts to load a configuration value from an environment variable.
// If the environment variable is not set, it returns the value from the flags.
//
// Parameters:
// - key: The name of the environment variable.
// - fromFlags: The default value to use if the environment variable is not set.
//
// Returns:
// - The value from the environment variable if set, otherwise the value from the flags.
// func tryLoadFromEnv(key, fromFlags string) string {
// 	value, ok := os.LookupEnv(key)
// 	if !ok {
// 		return fromFlags
// 	} else {
// 		return value
// 	}
// }

// Костыль который еще никто не видел на этом свете
func tryLoadFromEnv[T any](key string, fromFlags, fromFile T) T {
	// Пробуем получить значение из переменной окружения
	value, ok := os.LookupEnv(key)
	if ok && value != "" {
		// Используем рефлексию для определения типа fromFlags
		fromFlagsValue := reflect.ValueOf(fromFlags)
		fromFlagsType := fromFlagsValue.Type()

		// Преобразуем значение из переменной окружения в нужный тип
		switch fromFlagsType.Kind() {
		case reflect.String:
			return T(reflect.ValueOf(value).Convert(fromFlagsType).Interface().(T))
		case reflect.Int64:
			intValue, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fromFlags
			}
			return T(reflect.ValueOf(intValue).Convert(fromFlagsType).Interface().(T))
		case reflect.Bool:
			boolValue, err := strconv.ParseBool(value)
			if err != nil {
				return fromFlags
			}
			return T(reflect.ValueOf(boolValue).Convert(fromFlagsType).Interface().(T))
		default:
			return fromFlags
		}
	}

	// Если значение из переменной окружения пустое, пробуем значение из флагов
	if !reflect.ValueOf(fromFlags).IsZero() {
		return fromFlags
	}

	// Если значение из флагов пустое, возвращаем значение из файла
	return fromFile
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
