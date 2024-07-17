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
	// Agent
	defaultServerURL      = defaultListen
	defaultReportInterval = 2
	defaultPollInterval   = 10

	// Server
	hintListen        = "Server address"
	hintStorageDriver = "Storage driver"
	hintEnvMode       = "Enviriment server mode"
	// Agent
	hintServerURL      = hintListen
	hintReportInterval = "Report interval"
	hintPollInterval   = "Poll interval"
)

type ServerConfig struct {
	Listen        string `yaml:"Listen" env-default:"localhost:8080"`
	StorageDriver string `yaml:"StorageDriver" env-default:"mem"`
	EnvMode       string `yaml:"EnvMode" env-default:"dev"`
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

func LoadServerConfig() ServerConfig {
	var config ServerConfig
	fromFlags := loadServerConfigFromFlags()
	config.Listen = tryLoadFromEnv("ADDRESS", fromFlags.Listen)
	return config
}

// Try load Server Config from flags
func loadServerConfigFromFlags() ServerConfig {
	var config ServerConfig
	a := flag.String("a", defaultListen, hintListen)
	s := flag.String("storageDriver", defaultStorageDriver, hintStorageDriver)
	e := flag.String("env", defaultEnvMode, hintEnvMode)
	flag.Parse()

	config.Listen = *a
	config.StorageDriver = *s
	config.EnvMode = *e
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
