package common

import (
	"os"
)

func FileIsExists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func ReWriteFile(filePath string, data []byte) error {
	return os.WriteFile(filePath, data, 0600)
}
