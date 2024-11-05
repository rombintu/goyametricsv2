package config

import (
	"os"
	"reflect"
	"testing"
)

func TestLoadServerConfig(t *testing.T) {
	tests := []struct {
		name string
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем временный файл с нужными настройками
			tmpfile, err := os.CreateTemp("", "test_server_config_*.env")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name()) // Удаляем файл после выполнения теста

			// Записываем нужные настройки в файл
			configContent := `
ADDRESS=localhost:8080
STORAGE_DRIVER=mem
STORE_INTERVAL=300
FILE_STORAGE_PATH=store.json
RESTORE_FLAG=true
`
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

			// Вызываем функцию LoadServerConfig
			got := LoadServerConfig()

			// Проверяем, что настройки загружены корректно
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadServerConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
