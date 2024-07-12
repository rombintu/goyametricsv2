package config

import (
	"flag"
	"os"
	"strconv"
)

type ServerConfig struct {
	Listen        string `yaml:"Listen" env-default:"localhost:8080"`
	StorageDriver string `yaml:"StorageDriver" env-default:"mem"`
}

type AgentConfig struct {
	Address        string `yaml:"address" env-default:"http://localhost:8080"`
	PollInterval   int64  `yaml:"pollInterval" env-default:"2"`
	ReportInterval int64  `yaml:"reportInterval" env-default:"10"`
}

// Try load Server Config from flags
func LoadServerConfigFromFlags() ServerConfig {
	var config ServerConfig
	a := flag.String("a", "localhost:8080", "Server address")
	s := flag.String("storageDriver", "mem", "Storage driver")
	flag.Parse()

	config.Listen = *a
	config.StorageDriver = *s
	return config
}

// Load Agent Config from Environment, if any var empty - load from flags or set default
func LoadAgentConfig() AgentConfig {
	var config AgentConfig
	fromFlags := LoadAgentConfigFromFlags()

	address, ok := os.LookupEnv("ADDRESS")
	if !ok {
		config.Address = fromFlags.Address
	} else {
		config.Address = address
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

	config.Address = *a
	config.ReportInterval = *r
	config.PollInterval = *p

	return config
}
