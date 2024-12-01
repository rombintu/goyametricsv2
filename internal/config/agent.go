// Package config AgentConfig
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

type AgentConfig struct {
	Address        string `env-default:"http://localhost:8080" json:"address"`
	PollInterval   int64  `env-default:"2" json:"poll_interval"`
	ReportInterval int64  `env-default:"10" json:"report_interval"`
	EnvMode        string `env-default:"dev"`
	HashKey        string
	RateLimit      int64
	// Путь до файла с публичным ключом.
	PublicKeyFile string `json:"crypto_key"`
	SecureMode    bool

	ConfigPathFile string
}

// Try load Server Config from flags
func loadAgentConfigFromFlags() AgentConfig {
	var config AgentConfig
	a := flag.String("a", defaultServerURL, hintServerURL)
	r := flag.Int64("r", defaultReportInterval, hintReportInterval)
	p := flag.Int64("p", defaultPollInterval, hintPollInterval)
	k := flag.String("k", defaultHashKey, hintHashKey)
	l := flag.Int64("l", defaultRateLimit, hintRateLimit)
	pubkey := flag.String("crypto-key", defaultPubkeyFile, hintPubkeyFile)

	c := flag.String("c", defaultPathConfig, hintPathConfig)
	flag.Parse()

	config.Address = *a
	config.ReportInterval = *r
	config.PollInterval = *p
	config.HashKey = *k
	config.RateLimit = *l

	config.PublicKeyFile = *pubkey
	config.ConfigPathFile = *c
	return config
}

// Load Agent Config from Environment, if any var empty - load from flags or set default
func LoadAgentConfig() AgentConfig {
	var config AgentConfig
	var fromFile AgentConfig
	fromFlags := loadAgentConfigFromFlags()

	config.ConfigPathFile = tryLoadFromEnv("CONFIG", fromFlags.ConfigPathFile, "")

	if config.ConfigPathFile != "" {
		var err error
		fromFile, err = loadAgentConfigFromFile(config.ConfigPathFile)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	// Из тз нужно сделать такое ключевое слово, иначе не проходят тесты
	// ADDRESS отвечает за адрес эндпоинта HTTP-сервера.
	config.Address = tryLoadFromEnv("ADDRESS", fromFlags.Address, fromFile.Address)
	config.ReportInterval = tryLoadFromEnv("REPORT_INTERVAL", fromFlags.ReportInterval, fromFile.ReportInterval)
	config.PollInterval = tryLoadFromEnv("POLL_INTERVAL", fromFlags.PollInterval, fromFile.PollInterval)

	config.HashKey = tryLoadFromEnv("KEY", fromFlags.HashKey, fromFile.HashKey)
	config.RateLimit = tryLoadFromEnv("RATE_LIMIT", fromFlags.RateLimit, fromFile.RateLimit)

	config.PublicKeyFile = tryLoadFromEnv("CRYPTO_KEY", fromFlags.PublicKeyFile, fromFile.PublicKeyFile)
	return config
}

func loadAgentConfigFromFile(configPathFile string) (AgentConfig, error) {
	var newConfig AgentConfig

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
