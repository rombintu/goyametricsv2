package config

import (
	"flag"
)

type ServerConfig struct {
	Address       string `yaml:"address" env-default:"localhost:8080"`
	Environment   string `yaml:"Environment" env-default:"local"`
	StorageDriver string `yaml:"StorageDriver" env-default:"mem"`
}

type AgentConfig struct {
	ServerURL      string `yaml:"serverURL" env-default:"http://localhost:8080"`
	PollInterval   int64  `yaml:"pollInterval" env-default:"2"`
	ReportInterval int64  `yaml:"reportInterval" env-default:"10"`
	Mode           string `yaml:"mode" env-default:"debug"`
}

// Try load Server Config from flags
func LoadServerConfigFromFlags() ServerConfig {
	var config ServerConfig
	e := flag.String("environment", "local", "Environment")
	s := flag.String("storageDriver", "mem", "Storage driver")
	a := flag.String("a", "localhost:8080", "Server address")
	flag.Parse()

	config.Address = *a
	config.StorageDriver = *s
	config.Environment = *e
	return config
}

// Try load Server Config from flags
func LoadAgentConfigFromFlags() AgentConfig {
	var config AgentConfig
	a := flag.String("a", "localhost:8080", "Server address")
	r := flag.Int64("r", 10, "Report interval")
	p := flag.Int64("p", 2, "Poll interval")
	flag.Parse()

	config.ServerURL = *a
	config.ReportInterval = *r
	config.PollInterval = *p

	return config
}
