package config

import (
	"flag"
	"os"
	"testing"
)

func TestLoadServerConfig(t *testing.T) {
	env := make(map[string]string)
	env["ADDRESS"] = "localhost:8080"
	env["STORAGE_DRIVER"] = "mem"
	env["FILE_STORAGE_PATH"] = "store.json"
	env["STORE_INTERVAL"] = "300"
	env["RESTORE_FLAG"] = "true"
	env["DATABASE_DSN"] = ""

	tests := []struct {
		name string
		env  map[string]string
		want ServerConfig
	}{
		{
			name: "try_simple_load_server_config",
			want: ServerConfig{
				Listen:        "localhost:8080",
				StorageDriver: "mem",
				StoreInterval: 300,
				StoragePath:   "store.json",
				RestoreFlag:   true,
				SyncMode:      false,
			},
			env: env,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Сбрасываем флаги перед каждым тестом
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			flag.CommandLine.SetOutput(nil)

			for key, value := range tt.env {
				os.Setenv(key, value)
			}

			// Вызываем функцию LoadServerConfig
			got := LoadServerConfig()

			// Проверяем, что настройки загружены корректно
			if got != tt.want {
				t.Errorf("LoadServerConfig() = %v, want %v", got, tt.want)
			}

			// Очищаем переменные окружения после теста
			for key := range tt.env {
				os.Unsetenv(key)
			}
		})
	}
}
