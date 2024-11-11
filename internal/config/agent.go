// Package config AgentConfig
package config

type AgentConfig struct {
	Address        string `yaml:"address" env-default:"http://localhost:8080"`
	PollInterval   int64  `yaml:"pollInterval" env-default:"2"`
	ReportInterval int64  `yaml:"reportInterval" env-default:"10"`
	EnvMode        string `yaml:"EnvMode" env-default:"dev"`
	HashKey        string
	RateLimit      int64
}

func LoadConfigCollector() *ConfigCollector {
	priorities := make(PriorityMap)
	priorities[0] = envPriority  // 1 priority
	priorities[1] = flagPriority // 2 priority
	units := []ConfigUnit{
		{
			Code:        "a",
			KeyFlag:     "address",
			KeyEnv:      "ADDRESS",
			Description: hintServerURL,
			WantType:    stringType,
			Default:     defaultServerURL,
		},
		{
			Code:        "pi",
			KeyFlag:     "pollInterval",
			KeyEnv:      "POLL_INTERVAL",
			Description: hintPollInterval,
			WantType:    int64Type,
			Default:     defaultPollInterval,
		},
	}
	return NewConfigCollector(priorities, units...)
}

func LoadAgentConfig() AgentConfig {
	var config AgentConfig
	c := LoadConfigCollector()
	config.Address = Get[string](c.Search("a"))
	config.PollInterval = Get[int64](c.Search("pi"))
	return config
}

// // Try load Server Config from flags
// func loadAgentConfigFromFlags() AgentConfig {
// 	var config AgentConfig
// 	a := flag.String("a", defaultServerURL, hintServerURL)
// 	r := flag.Int64("r", defaultReportInterval, hintReportInterval)
// 	p := flag.Int64("p", defaultPollInterval, hintPollInterval)
// 	k := flag.String("k", defaultHashKey, hintHashKey)
// 	l := flag.Int64("l", defaultRateLimit, hintRateLimit)
// 	flag.Parse()

// 	config.Address = *a
// 	config.ReportInterval = *r
// 	config.PollInterval = *p
// 	config.HashKey = *k
// 	config.RateLimit = *l

// 	return config
// }

// // Load Agent Config from Environment, if any var empty - load from flags or set default
// func LoadAgentConfig() AgentConfig {
// 	var config AgentConfig
// 	fromFlags := loadAgentConfigFromFlags()
// 	// Из тз нужно сделать такое ключевое слово, иначе не проходят тесты
// 	// ADDRESS отвечает за адрес эндпоинта HTTP-сервера.
// 	config.Address = tryLoadFromEnv("ADDRESS", fromFlags.Address)
// 	config.ReportInterval = tryLoadFromEnvInt64("REPORT_INTERVAL", fromFlags.ReportInterval)
// 	config.PollInterval = tryLoadFromEnvInt64("POLL_INTERVAL", fromFlags.PollInterval)

// 	config.HashKey = tryLoadFromEnv("KEY", fromFlags.HashKey)
// 	config.RateLimit = tryLoadFromEnvInt64("RATE_LIMIT", fromFlags.RateLimit)
// 	return config
// }
