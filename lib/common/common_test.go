// Package common tests
package common

import (
	"fmt"
	"net"
	"os"
	"testing"
)

func TestFileIsExists(t *testing.T) {
	testCases := []struct {
		name           string
		filePath       string
		createFile     bool
		expectedResult bool
	}{
		{
			name:           "File_Exists",
			filePath:       "testfile.txt",
			createFile:     true,
			expectedResult: true,
		},
		{
			name:           "File_Does_Not_Exist",
			filePath:       "nonexistentfile.txt",
			createFile:     false,
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.createFile {
				file, err := os.Create(tc.filePath)
				if err != nil {
					t.Fatalf("Failed to create file: %v", err)
				}
				file.Close()
				defer os.Remove(tc.filePath)
			}

			result := FileIsExists(tc.filePath)
			if result != tc.expectedResult {
				t.Errorf("Expected FileIsExists(%s) to be %v, but got %v", tc.filePath, tc.expectedResult, result)
			}
		})
	}
}

func TestReWriteFile(t *testing.T) {
	testCases := []struct {
		name          string
		filePath      string
		data          []byte
		expectedError bool
	}{
		{
			name:          "Valid_File_Path",
			filePath:      "testfile.txt",
			data:          []byte("test data"),
			expectedError: false,
		},
		{
			name:          "Invalid_File_Path",
			filePath:      "/invalid/path/testfile.txt",
			data:          []byte("test data"),
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ReWriteFile(tc.filePath, tc.data)
			if tc.expectedError && err == nil {
				t.Errorf("Expected error, but got none")
			} else if !tc.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tc.expectedError {
				defer os.Remove(tc.filePath)

				writtenData, err := os.ReadFile(tc.filePath)
				if err != nil {
					t.Fatalf("Failed to read written file: %v", err)
				}

				if string(writtenData) != string(tc.data) {
					t.Errorf("Written data does not match expected data")
				}
			}
		})
	}
}

// TestIsPortInUse_WhenPortIsFree проверяет, что функция возвращает false для свободного порта.
func TestIsPortInUse_WhenPortIsFree(t *testing.T) {
	port := int64(3200) // Используем порт, который, скорее всего, свободен

	// Проверяем, что порт свободен
	if IsPortInUse(port) {
		t.Errorf("Expected port %d to be free, but it's in use", port)
	}
}

// TestIsPortInUse_WhenPortIsInUse проверяет, что функция возвращает true для занятого порта.
func TestIsPortInUse_WhenPortIsInUse(t *testing.T) {
	port := int64(3201) // Используем порт для теста

	// Запускаем тестовый сервер на порту
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer listener.Close()

	// Проверяем, что порт занят
	if !IsPortInUse(port) {
		t.Errorf("Expected port %d to be in use, but it's free", port)
	}
}

// TestIsPortInUse_WhenPortIsInvalid проверяет, что функция корректно обрабатывает неверный порт.
func TestIsPortInUse_WhenPortIsInvalid(t *testing.T) {
	invalidPort := int64(-1) // Неверный порт

	// Проверяем, что функция возвращает false для неверного порта
	if IsPortInUse(invalidPort) {
		t.Errorf("Expected invalid port %d to be considered free", invalidPort)
	}
}
