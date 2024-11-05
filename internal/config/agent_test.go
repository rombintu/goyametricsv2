// Package config AgentConfig
package config

import (
	"os"
	"testing"
)

func TestLoadAgentConfig(t *testing.T) {
	env := make(map[string]string)
	env["ADDRESS"] = "localhost:8080"
	env["REPORT_INTERVAL"] = "2"
	env["POLL_INTERVAL"] = "10"
	env["RATE_LIMIT"] = "0"
	env["KEY"] = "secret"
	tests := []struct {
		name string
		env  map[string]string
		want AgentConfig
	}{
		{
			name: "try_simple_load_agent_config",
			want: AgentConfig{
				Address:        "localhost:8080",
				PollInterval:   10,
				ReportInterval: 2,
				RateLimit:      0,
				HashKey:        "secret",
			},
			env: env,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.env {
				os.Setenv(key, value)
			}

			// Вызываем функцию LoadServerConfig
			got := LoadAgentConfig()

			// Проверяем, что настройки загружены корректно
			if got != tt.want {
				t.Errorf("LoadAgentConfig() = %v, want %v", got, tt.want)
			}

			// Очищаем переменные окружения после теста
			for key := range tt.env {
				os.Unsetenv(key)
			}
		})
	}
}
