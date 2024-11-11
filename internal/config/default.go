// Package config DefaultConfig
package config

import (
	"flag"
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
	defaultServerURL            = defaultListen
	defaultReportInterval int64 = 2
	defaultPollInterval   int64 = 10
	defaultRateLimit      int64 = 0

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

type PriorityMap map[int]string // Неизменяемый список приоритетов

const (
	filePriority = "file"
	flagPriority = "flag"
	envPriority  = "env"
)

const (
	stringType = "string"
	int64Type  = "int64"
	boolType   = "bool"
)

type ConfigUnit struct {
	Code        string
	KeyFlag     string
	KeyFile     string
	KeyEnv      string
	Description string
	WantType    string
	Value       any
	Default     any
	Defined     bool
}

func (cu *ConfigUnit) Set(v any) {
	if v != nil {
		cu.Value = v
		cu.Defined = true
		return
	}
	cu.Value = cu.Default
}

// Дженерик
func Get[T any](cu ConfigUnit) T {
	return cu.Value.(T)
}

type ConfigCollector struct {
	cursor      int
	priorities  PriorityMap
	ConfigUnits []ConfigUnit
}

func (cl *ConfigCollector) Search(code string) ConfigUnit {
	for {
		unit, ok := cl.Next()
		if unit.Code == code {
			return unit
		}
		if !ok {
			break
		}
	}
	return ConfigUnit{}
}

func NewConfigCollector(priorities PriorityMap, cUnits ...ConfigUnit) *ConfigCollector {
	cl := new(ConfigCollector)
	cl.priorities = priorities
	// ... Start collect
	for _, unit := range cUnits {
		if !unit.Defined {
			for _, pr := range priorities {
				switch pr {
				// case filePriority:
				case flagPriority:
					switch unit.WantType {
					case int64Type:
						val := flag.Int64(unit.KeyFlag, unit.Default.(int64), unit.Description)
						unit.Set(*val)
					case stringType:
						val := flag.String(unit.KeyFlag, unit.Default.(string), unit.Description)
						unit.Set(*val)
					case boolType:
						val := flag.Bool(unit.KeyFlag, unit.Default.(bool), unit.Description)
						unit.Set(*val)

					}
				case envPriority:
					envValue, ok := os.LookupEnv(unit.KeyEnv)
					if ok {
						unit.Set(envValue)
					}
				}
			}
		}
	}
	flag.Parse()

	// ... End collect
	cl.ConfigUnits = append(cl.ConfigUnits, cUnits...)
	return cl
}

func (cl *ConfigCollector) Next() (ConfigUnit, bool) {
	if cl.cursor <= len(cl.ConfigUnits) {
		cl.cursor++
	}
	return cl.ConfigUnits[cl.cursor-1], cl.cursor < len(cl.ConfigUnits)
}
