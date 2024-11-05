// Package config AgentConfig
package config

import (
	"os"
	"reflect"
	"testing"
)

func TestLoadAgentConfig(t *testing.T) {
	tests := []struct {
		name string
		want AgentConfig
	}{
		{
			name: "try_simple_load_agent_config",
			want: AgentConfig{
				Address:        "localhost:8080",
				PollInterval:   10,
				ReportInterval: 2,
				RateLimit:      0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем временный файл с нужными настройками
			tmpfile, err := os.CreateTemp("", "test_agent_config_*.env")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name()) // Удаляем файл после выполнения теста

			// Записываем нужные настройки в файл
			configContent := ``
			if _, err := tmpfile.Write([]byte(configContent)); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}

			// Закрываем файл, чтобы он был доступен для чтения
			if err := tmpfile.Close(); err != nil {
				t.Fatalf("Failed to close temp file: %v", err)
			}

			// Сохраняем текущее значение переменной окружения
			originalEnv := os.Getenv("CONFIG_PATH")
			defer os.Setenv("CONFIG_PATH", originalEnv) // Восстанавливаем оригинальное значение после теста

			// Устанавливаем путь к временному файлу в переменную окружения
			os.Setenv("CONFIG_PATH", tmpfile.Name())

			// Вызываем функцию LoadAgentConfig
			got := LoadAgentConfig()

			// Проверяем, что настройки загружены корректно
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadAgentConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
