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

func TestLoadAgentConfigFromFile(t *testing.T) {
	testCases := []struct {
		name           string
		configPathFile string
		createFile     bool
		configData     string
		expectedConfig AgentConfig
		expectedError  bool
	}{
		{
			name:           "Valid_Config_File",
			configPathFile: "testconfig.json",
			createFile:     true,
			configData:     `{"address": "localhost:8080", "report_interval": 10}`,
			expectedConfig: AgentConfig{
				Address:        "localhost:8080",
				ReportInterval: 10,
			},
			expectedError: false,
		},
		{
			name:           "Non-Existent_File",
			configPathFile: "nonexistentfile.json",
			createFile:     false,
			expectedError:  true,
		},
		{
			name:           "Invalid_JSON",
			configPathFile: "invalidconfig.json",
			createFile:     false,
			configData:     `{"host": "localhost", "port": "invalid"}`,
			expectedError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.createFile {
				err := os.WriteFile(tc.configPathFile, []byte(tc.configData), 0644)
				if err != nil {
					t.Fatalf("Failed to create config file: %v", err)
				}
				defer os.Remove(tc.configPathFile)
			}

			config, err := loadAgentConfigFromFile(tc.configPathFile)
			if tc.expectedError && err == nil {
				t.Errorf("Expected error, but got none")
			} else if !tc.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tc.expectedError {
				if config != tc.expectedConfig {
					t.Errorf("Expected config %+v, but got %+v", tc.expectedConfig, config)
				}
			}
		})
	}
}
