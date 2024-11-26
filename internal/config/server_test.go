package config

import (
	"flag"
	"os"
	"testing"
)

func teardown(envlist map[string]string) {
	for key, value := range envlist {
		os.Setenv(key, value)
	}
}

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
			defer teardown(tt.env)
			// Сбрасываем флаги перед каждым тестом
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			flag.CommandLine.SetOutput(nil)

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

func TestLoadServerConfigFromFile(t *testing.T) {
	testCases := []struct {
		name           string
		configPathFile string
		createFile     bool
		configData     string
		expectedConfig ServerConfig
		expectedError  bool
	}{
		{
			name:           "Valid_Config_File",
			configPathFile: "testconfig.json",
			createFile:     true,
			configData:     `{"address": "localhost:8080", "store_interval": 0}`,
			expectedConfig: ServerConfig{
				Listen:        "localhost:8080",
				StoreInterval: 0,
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

			config, err := loadServerConfigFromFile(tc.configPathFile)
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
