package agent

import (
	"testing"

	"github.com/rombintu/goyametricsv2/internal/config"
)

func TestAgentLoadMetrics(t *testing.T) {
	config := config.LoadAgentConfig()
	agent := NewAgent(config)
	agent.loadMetrics()

	if len(agent.data.Counters) == 0 && agent.pollCount == 0 {
		t.Error("Expected counters metrics to be loaded")
	}
	if len(agent.data.Gauges) == 0 {
		t.Error("Expected gauges metrics to be loaded")
	}
}
