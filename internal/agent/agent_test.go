package agent

import "testing"

func TestAgentLoadMetrics(t *testing.T) {
	agent := NewAgent("localhost:8080", 2, 10)
	agent.loadMetrics()
	if len(agent.metrics) == 0 {
		t.Error("Expected metrics to be loaded")
	}
}
