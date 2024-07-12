package main

import (
	"time"

	"github.com/rombintu/goyametricsv2/internal/agent"
	"github.com/rombintu/goyametricsv2/internal/config"
)

func main() {
	config := config.LoadAgentConfigFromFlags()
	a := agent.NewAgent(config)

	go a.RunPoll()
	go a.RunReport()
	for {
		time.Sleep(1 * time.Second)
	}
}
