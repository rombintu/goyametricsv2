// Package config ServerConfig
package config

import (
	"flag"

	"github.com/rombintu/goyametricsv2/internal/storage"
)

type ServerConfig struct {
	Listen        string `yaml:"Listen" env-default:"localhost:8080"`
	StorageDriver string `yaml:"StorageDriver" env-default:"mem"`
	EnvMode       string `yaml:"EnvMode" env-default:"dev"`
	StoreInterval int64  `yaml:"StoreInterval" env-default:"300"`
	StoragePath   string `yaml:"StoragePath" env-default:"store.json"`
	StorageURL    string `yaml:"StorageURL"`
	RestoreFlag   bool   `yaml:"RestoreFlag" env-default:"true"`
	SyncMode      bool   `yaml:"SyncMode" env-default:"false"`

	// Ключ для подписи
	HashKey string
	// Путь до файла с приватным ключом
	PrivateKeyFile string
	SecureMode     bool
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
	// New
	k := flag.String("k", defaultHashKey, hintHashKey)
	privateKeyFile := flag.String("crypto-key", defaultPrivateKeyFile, hintPrivateKeyFile)

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
	return config
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

	// increment 14
	config.HashKey = tryLoadFromEnv("KEY", fromFlags.HashKey)

	if config.StorageURL != "" {
		config.StorageDriver = storage.PgxDriver
	} else if (config.StoragePath != "") && (config.StoragePath != defaultStoragePath) {
		config.StorageDriver = storage.FileDriver
	}

	// increment 21
	config.PrivateKeyFile = tryLoadFromEnv("CRYPTO_KEY", fromFlags.PrivateKeyFile)

	return config
}
