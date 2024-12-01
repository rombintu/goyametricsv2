// Package common tests
package common

import (
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
