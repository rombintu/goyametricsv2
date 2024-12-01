// Package common provides utility functions for file operations and other common tasks.
package common

import (
	"os"
)

// FileIsExists checks if a file exists at the specified path.
// It returns true if the file exists, otherwise false.
func FileIsExists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

// ReWriteFile overwrites the file at the specified path with the provided data.
// It returns an error if the operation fails.
func ReWriteFile(filePath string, data []byte) error {
	return os.WriteFile(filePath, data, 0600)
}
