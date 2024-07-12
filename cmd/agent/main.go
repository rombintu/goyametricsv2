package main

import (
	"time"

	"github.com/rombintu/goyametricsv2/internal/agent"
)

func main() {
	a := agent.NewAgent(
		"localhost:8080",
		2,
		10,
	)

	go a.RunPoll()
	go a.RunReport()
	for {
		time.Sleep(1 * time.Second)
	}
}
