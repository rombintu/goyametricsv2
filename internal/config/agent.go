// Package config AgentConfig
package config

import (
	"flag"
)

type AgentConfig struct {
	Address        string `yaml:"address" env-default:"http://localhost:8080"`
	PollInterval   int64  `yaml:"pollInterval" env-default:"2"`
	ReportInterval int64  `yaml:"reportInterval" env-default:"10"`
	EnvMode        string `yaml:"EnvMode" env-default:"dev"`
	HashKey        string
	RateLimit      int64
	// Путь до файла с публичным ключом.
	PublicKeyFile string
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
	flag.Parse()

	config.Address = *a
	config.ReportInterval = *r
	config.PollInterval = *p
	config.HashKey = *k
	config.RateLimit = *l

	config.PublicKeyFile = *pubkey

	return config
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

	config.HashKey = tryLoadFromEnv("KEY", fromFlags.HashKey)
	config.RateLimit = tryLoadFromEnvInt64("RATE_LIMIT", fromFlags.RateLimit)

	config.PublicKeyFile = tryLoadFromEnv("CRYPTO_KEY", fromFlags.PublicKeyFile)
	return config
}
