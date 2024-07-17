package main

import (
	"time"

	"github.com/labstack/gommon/log"
	"github.com/rombintu/goyametricsv2/internal/agent"
	"github.com/rombintu/goyametricsv2/internal/config"
)

func main() {
	// Еще попытка. Гитхаб лагает
	config := config.LoadAgentConfig()
	a := agent.NewAgent(config)
	log.Infof("Agent starting on: %s", config.Address)
	go a.RunPoll()
	go a.RunReport()
	for {
		time.Sleep(1 * time.Second)
	}
}
