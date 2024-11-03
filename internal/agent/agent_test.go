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

func Test_fixServerURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "simple fix",
			args: args{url: "http://google.com"},
			want: "http://google.com",
		},
		{
			name: "simple fix 2",
			args: args{url: "google.com"},
			want: "http://google.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fixServerURL(tt.args.url); got != tt.want {
				t.Errorf("fixServerURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgent_loadPSUtilsMetrics(t *testing.T) {

	tests := []struct {
		name           string
		lenIsMoreThen0 bool
	}{
		{
			name:           "load cpu utils metrics",
			lenIsMoreThen0: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Agent{data: Data{}}
			if got := a.loadPSUtilsMetrics(); len(got.Counters) != 0 {
				t.Errorf("Agent.loadPSUtilsMetrics() = %+v, want %v", got, tt.lenIsMoreThen0)
			}
			if got := a.loadPSUtilsMetrics(); len(got.Gauges) == 0 {
				t.Errorf("Agent.loadPSUtilsMetrics() = %+v, want %v", got, tt.lenIsMoreThen0)
			}
		})
	}
}
