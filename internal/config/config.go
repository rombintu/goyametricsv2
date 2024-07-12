package config

import (
	"flag"
	"os"
	"strconv"
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

// Load Agent Config from Environment, if any var empty - load from flags or set default
func LoadAgentConfig() AgentConfig {
	var config AgentConfig
	fromFlags := LoadAgentConfigFromFlags()

	// Try load ADDRESS (ServerUrl)
	address, ok := os.LookupEnv("ADDRESS")
	if !ok {
		config.ServerURL = fromFlags.ServerURL
	} else {
		config.ServerURL = address
	}

	// Try load REPORT_INTERVAL
	ri, ok := os.LookupEnv("REPORT_INTERVAL")
	if !ok {
		config.ReportInterval = fromFlags.ReportInterval
	} else {
		parseRI, err := strconv.ParseInt(ri, 10, 64)
		if err != nil {
			config.ReportInterval = fromFlags.ReportInterval
		} else {
			config.ReportInterval = parseRI
		}
	}

	// Try load POLL_INTERVAL
	pi, ok := os.LookupEnv("POLL_INTERVAL")
	if !ok {
		config.PollInterval = fromFlags.PollInterval
	} else {
		parsePI, err := strconv.ParseInt(pi, 10, 64)
		if err != nil {
			config.PollInterval = fromFlags.PollInterval
		} else {
			config.PollInterval = parsePI
		}
	}
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
