// Package config ServerConfig
package config

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/rombintu/goyametricsv2/internal/storage"
)

type ServerConfig struct {
	Listen        string `env-default:"localhost:8080" json:"address"`
	StorageDriver string `env-default:"mem"`
	EnvMode       string `env-default:"dev" json:"env_mode"`
	StoreInterval int64  `env-default:"300" json:"store_interval"`
	StoragePath   string `env-default:"store.json" json:"store_file"`
	StorageURL    string `json:"database_dsn"`
	RestoreFlag   bool   `env-default:"true" json:"restore"`
	SyncMode      bool   `env-default:"false"`

	// Ключ для подписи
	HashKey string
	// Путь до файла с приватным ключом
	PrivateKeyFile string `json:"crypto_key"`
	SecureMode     bool

	// Config parse from json
	ConfigPathFile string
}

// Try load Server Config from flags
func loadServerConfigFromFlags() ServerConfig {
	var config ServerConfig
	a := flag.String("a", defaultListen, hintListen)
	s := flag.String("driver", defaultStorageDriver, hintStorageDriver)
	e := flag.String("env", defaultEnvMode, hintEnvMode)
	i := flag.Int64("i", defaultStoreInterval, hintStoreInterval)
	f := flag.String("f", defaultStoragePath, hintStoragePath)
	r := flag.Bool("r", defaultRestoreFlag, hintRestoreFlag)
	d := flag.String("d", "", hintStorageURL)

	k := flag.String("k", defaultHashKey, hintHashKey)
	privateKeyFile := flag.String("crypto-key", defaultPrivateKeyFile, hintPrivateKeyFile)

	configFile := flag.String("c", defaultPathConfig, hintPathConfig)

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
	// increment 14
	config.HashKey = *k

	// increment 21
	config.PrivateKeyFile = *privateKeyFile

	// increment 22
	config.ConfigPathFile = *configFile

	return config
}

func LoadServerConfig() ServerConfig {
	var fromFile ServerConfig
	var config ServerConfig
	fromFlags := loadServerConfigFromFlags()

	config.ConfigPathFile = tryLoadFromEnv("CONFIG", fromFlags.ConfigPathFile, "")

	if config.ConfigPathFile != "" {
		fromFile, _ = loadServerConfigFromFile(config.ConfigPathFile)
	}

	config.Listen = tryLoadFromEnv("ADDRESS", fromFlags.Listen, fromFile.Listen)
	// New args
	config.StoreInterval = tryLoadFromEnv("STORE_INTERVAL", fromFlags.StoreInterval, fromFile.StoreInterval)
	config.StoragePath = tryLoadFromEnv("FILE_STORAGE_PATH", fromFlags.StoragePath, fromFile.StoragePath)
	config.RestoreFlag = tryLoadFromEnv("RESTORE_FLAG", fromFlags.RestoreFlag, fromFile.RestoreFlag)

	// increment 10
	config.StorageDriver = tryLoadFromEnv("STORAGE_DRIVER", fromFlags.StorageDriver, fromFile.StorageDriver)
	config.StorageURL = tryLoadFromEnv("DATABASE_DSN", fromFlags.StorageURL, fromFile.StorageURL)
	// Change to sync mode
	if config.StoreInterval == 0 {
		config.SyncMode = true
	}

	// increment 14
	config.HashKey = tryLoadFromEnv("KEY", fromFlags.HashKey, fromFile.HashKey)

	if config.StorageURL != "" {
		config.StorageDriver = storage.PgxDriver
	} else if (config.StoragePath != "") && (config.StoragePath != defaultStoragePath) {
		config.StorageDriver = storage.FileDriver
	}

	// increment 21
	config.PrivateKeyFile = tryLoadFromEnv("CRYPTO_KEY", fromFlags.PrivateKeyFile, fromFile.PrivateKeyFile)

	// increment 22

	return config
}

func loadServerConfigFromFile(configPathFile string) (ServerConfig, error) {
	var newConfig ServerConfig

	// Чтение JSON-файла
	data, err := os.ReadFile(configPathFile)
	if err != nil {
		return newConfig, err
	}

	// Декодирование JSON в структуру
	err = json.Unmarshal(data, &newConfig)
	if err != nil {
		return newConfig, err
	}

	return newConfig, nil
}
