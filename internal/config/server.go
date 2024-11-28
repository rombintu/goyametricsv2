// Package config ServerConfig
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/rombintu/goyametricsv2/internal/storage"
)

type ServerConfig struct {
	Listen        string `env-default:"localhost:8080" json:"address"`
	StorageDriver string `env-default:"mem" json:"-"`
	EnvMode       string `env-default:"dev" json:"-"`
	StoreInterval int64  `env-default:"300" json:"store_interval"`
	StoragePath   string `env-default:"store.json" json:"store_file"`
	StorageURL    string `json:"database_dsn"`
	RestoreFlag   bool   `env-default:"true" json:"restore"`
	SyncMode      bool   `env-default:"false"`

	// Ключ для подписи
	HashKey string `json:"-"`
	// Путь до файла с приватным ключом
	PrivateKeyFile string `json:"crypto_key"`
	SecureMode     bool   `json:"-"`

	// Config parse from json
	ConfigPathFile string `json:"-"`
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
		var err error
		fromFile, err = loadServerConfigFromFile(config.ConfigPathFile)
		if err != nil {
			fmt.Println(err.Error())
		}
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

	return config
}

func loadServerConfigFromFile(configPathFile string) (ServerConfig, error) {
	var newConfig ServerConfig

	// Чтение JSON-файла
	data, err := os.ReadFile(configPathFile)
	if err != nil {
		return newConfig, err
	}

	// // Декодирование JSON в структуру
	// err = json.Unmarshal(data, &newConfig)
	err = newConfig.UnmarshalJSON(data)
	if err != nil {
		return newConfig, err
	}

	return newConfig, nil
}

func (c *ServerConfig) UnmarshalJSON(data []byte) error {
	type Alias ServerConfig // Создаем алиас для структуры, чтобы избежать рекурсии
	aux := &struct {
		*Alias
		StoreInterval string `json:"store_interval"`
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Проверяем, что поле store_interval не пустое
	if aux.StoreInterval == "" {
		return nil
	}

	// Регулярное выражение для разбора строки
	re := regexp.MustCompile(`^(\d+)([smh]?)$`)
	matches := re.FindStringSubmatch(aux.StoreInterval)
	if len(matches) != 3 {
		return fmt.Errorf("invalid store_interval format: %s", aux.StoreInterval)
	}

	// Преобразуем числовое значение в int64
	value, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid store_interval value: %s", matches[1])
	}

	// Определяем множитель в зависимости от единицы измерения
	unit := matches[2]
	switch unit {
	case "s":
		// секунды, ничего не делаем
	case "m":
		value *= 60 // минуты, умножаем на 60
	case "h":
		value *= 3600 // часы, умножаем на 3600
	default:
		return fmt.Errorf("unknown unit in store_interval: %s", unit)
	}

	// Присваиваем преобразованное значение полю структуры
	c.StoreInterval = value

	return nil
}
