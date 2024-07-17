package agent

import (
	"testing"

	"github.com/rombintu/goyametricsv2/internal/config"
)

func TestAgentLoadMetrics(t *testing.T) {
	config := config.LoadAgentConfigFromFlags()
	agent := NewAgent(config)
	agent.loadMetrics()
	if len(agent.metrics) == 0 {
		t.Error("Expected metrics to be loaded")
	}
}
