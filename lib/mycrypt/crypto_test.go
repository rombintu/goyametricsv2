// Package mycrypt
package mycrypt

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path"
	"testing"
)

func TestGenRSAKeyPair(t *testing.T) {
	tmpDir := os.TempDir()
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "gen_files",
			args:    args{filePath: path.Join(tmpDir, "master")},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := GenRSAKeyPair(tt.args.filePath); (err != nil) != tt.wantErr {
				t.Errorf("GenRSAKeyPair() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidPrivateKey tests the ValidPrivateKey function with various scenarios.
func TestValidPrivateKey(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name           string
		generateKey    func() (*rsa.PrivateKey, error)
		saveKey        func(key *rsa.PrivateKey) (string, error)
		expectedResult bool
	}{
		{
			name: "Valid Private Key",
			generateKey: func() (*rsa.PrivateKey, error) {
				return rsa.GenerateKey(rand.Reader, 4096)
			},
			saveKey: func(key *rsa.PrivateKey) (string, error) {
				tempFile, err := os.CreateTemp("", "private_key_test")
				if err != nil {
					return "", err
				}
				defer tempFile.Close()
				if err := SavePrivateKey(tempFile.Name(), key); err != nil {
					return "", err
				}
				return tempFile.Name(), nil
			},
			expectedResult: true,
		},
		{
			name: "Non-Existent File",
			generateKey: func() (*rsa.PrivateKey, error) {
				return nil, nil
			},
			saveKey: func(key *rsa.PrivateKey) (string, error) {
				return "non_existent_file.pem", nil
			},
			expectedResult: false,
		},
		{
			name: "Invalid Key Content",
			generateKey: func() (*rsa.PrivateKey, error) {
				return nil, nil
			},
			saveKey: func(key *rsa.PrivateKey) (string, error) {
				tempFile, err := os.CreateTemp("", "invalid_key_test")
				if err != nil {
					return "", err
				}
				defer tempFile.Close()
				if _, err := tempFile.Write([]byte("invalid content")); err != nil {
					return "", err
				}
				return tempFile.Name(), nil
			},
			expectedResult: false,
		},
		{
			name: "Invalid Key Length",
			generateKey: func() (*rsa.PrivateKey, error) {
				return rsa.GenerateKey(rand.Reader, 2048)
			},
			saveKey: func(key *rsa.PrivateKey) (string, error) {
				tempFile, err := os.CreateTemp("", "private_key_test")
				if err != nil {
					return "", err
				}
				defer tempFile.Close()
				if err := SavePrivateKey(tempFile.Name(), key); err != nil {
					return "", err
				}
				return tempFile.Name(), nil
			},
			expectedResult: false,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var filePath string
			if tc.generateKey != nil {
				key, err := tc.generateKey()
				if err != nil {
					t.Fatalf("Failed to generate private key: %v", err)
				}
				if tc.saveKey != nil {
					filePath, err = tc.saveKey(key)
					if err != nil {
						t.Fatalf("Failed to save private key: %v", err)
					}
					defer os.Remove(filePath)
				}
			} else {
				if tc.saveKey != nil {
					filePath, _ = tc.saveKey(nil)
				}
			}

			result := ValidPrivateKey(filePath)
			if result != tc.expectedResult {
				t.Errorf("Expected ValidPrivateKey(%s) to be %v, but got %v", filePath, tc.expectedResult, result)
			}
		})
	}
}

// TestSavePrivateKey tests the SavePrivateKey function with various scenarios.
func TestSavePrivateKey(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name          string
		generateKey   func() (*rsa.PrivateKey, error)
		saveKey       func(key *rsa.PrivateKey) (string, error)
		expectedError bool
	}{
		{
			name: "Valid Private Key",
			generateKey: func() (*rsa.PrivateKey, error) {
				return rsa.GenerateKey(rand.Reader, 4096)
			},
			saveKey: func(key *rsa.PrivateKey) (string, error) {
				tempFile, err := os.CreateTemp("", "private_key_test")
				if err != nil {
					return "", err
				}
				defer tempFile.Close()
				if err := SavePrivateKey(tempFile.Name(), key); err != nil {
					return "", err
				}
				return tempFile.Name(), nil
			},
			expectedError: false,
		},
		{
			name: "Invalid File Path",
			generateKey: func() (*rsa.PrivateKey, error) {
				return rsa.GenerateKey(rand.Reader, 4096)
			},
			saveKey: func(key *rsa.PrivateKey) (string, error) {
				invalidFilePath := "/invalid/path/private_key_test"
				if err := SavePrivateKey(invalidFilePath, key); err != nil {
					return invalidFilePath, err
				}
				return invalidFilePath, nil
			},
			expectedError: true,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var filePath string
			var err error

			// Generate the private key
			key, err := tc.generateKey()
			if err != nil {
				t.Fatalf("Failed to generate private key: %v", err)
			}

			// Save the private key to the file
			filePath, err = tc.saveKey(key)
			if err != nil {
				if !tc.expectedError {
					t.Fatalf("Unexpected error: %v", err)
				}
				return
			}
			defer os.Remove(filePath)

			// If no error was expected, read and verify the saved key
			if !tc.expectedError {
				// Read the saved file
				savedKeyBytes, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("Failed to read saved private key file: %v", err)
				}

				// Decode the PEM block
				block, _ := pem.Decode(savedKeyBytes)
				if block == nil || block.Type != "RSA PRIVATE KEY" {
					t.Fatalf("Failed to decode PEM block")
				}

				// Parse the private key
				savedPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
				if err != nil {
					t.Fatalf("Failed to parse private key: %v", err)
				}

				// Compare the original private key with the parsed private key
				if savedPrivateKey.N.Cmp(key.N) != 0 {
					t.Errorf("Saved private key does not match the original private key")
				}
			}
		})
	}
}

func TestSavePublicKey(t *testing.T) {
	testCases := []struct {
		name          string
		generateKey   func() (*rsa.PublicKey, error)
		saveKey       func(key *rsa.PublicKey) (string, error)
		expectedError bool
	}{
		{
			name: "Valid Public Key",
			generateKey: func() (*rsa.PublicKey, error) {
				privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
				if err != nil {
					return nil, err
				}
				return &privateKey.PublicKey, nil
			},
			saveKey: func(key *rsa.PublicKey) (string, error) {
				tempFile, err := os.CreateTemp("", "public_key_test")
				if err != nil {
					return "", err
				}
				defer tempFile.Close()
				if err := SavePublicKey(tempFile.Name(), key); err != nil {
					return "", err
				}
				return tempFile.Name(), nil
			},
			expectedError: false,
		},
		{
			name: "Invalid File Path",
			generateKey: func() (*rsa.PublicKey, error) {
				privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
				if err != nil {
					return nil, err
				}
				return &privateKey.PublicKey, nil
			},
			saveKey: func(key *rsa.PublicKey) (string, error) {
				invalidFilePath := "/invalid/path/public_key_test"
				if err := SavePublicKey(invalidFilePath, key); err != nil {
					return invalidFilePath, err
				}
				return invalidFilePath, nil
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var filePath string
			var err error

			key, err := tc.generateKey()
			if err != nil {
				t.Fatalf("Failed to generate public key: %v", err)
			}

			filePath, err = tc.saveKey(key)
			if err != nil {
				if !tc.expectedError {
					t.Fatalf("Unexpected error: %v", err)
				}
				return
			}
			defer os.Remove(filePath)

			if !tc.expectedError {
				savedKeyBytes, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("Failed to read saved public key file: %v", err)
				}

				block, _ := pem.Decode(savedKeyBytes)
				if block == nil || block.Type != "PUBLIC KEY" {
					t.Fatalf("Failed to decode PEM block")
				}

				savedPublicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
				if err != nil {
					t.Fatalf("Failed to parse public key: %v", err)
				}

				if savedPublicKey.N.Cmp(key.N) != 0 {
					t.Errorf("Saved public key does not match the original public key")
				}
			}
		})
	}
}

func TestLoadPrivateKey(t *testing.T) {
	testCases := []struct {
		name          string
		generateKey   func() (*rsa.PrivateKey, error)
		saveKey       func(key *rsa.PrivateKey) (string, error)
		expectedError bool
	}{
		{
			name: "Valid Private Key",
			generateKey: func() (*rsa.PrivateKey, error) {
				return rsa.GenerateKey(rand.Reader, 4096)
			},
			saveKey: func(key *rsa.PrivateKey) (string, error) {
				tempFile, err := os.CreateTemp("", "private_key_test")
				if err != nil {
					return "", err
				}
				defer tempFile.Close()
				if err := SavePrivateKey(tempFile.Name(), key); err != nil {
					return "", err
				}
				return tempFile.Name(), nil
			},
			expectedError: false,
		},
		{
			name: "Non-Existent File",
			generateKey: func() (*rsa.PrivateKey, error) {
				return nil, nil
			},
			saveKey: func(key *rsa.PrivateKey) (string, error) {
				return "non_existent_file.pem", nil
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var filePath string
			var err error

			key, err := tc.generateKey()
			if err != nil {
				t.Fatalf("Failed to generate private key: %v", err)
			}

			if tc.saveKey != nil {
				filePath, err = tc.saveKey(key)
				if err != nil {
					t.Fatalf("Failed to save private key: %v", err)
				}
				defer os.Remove(filePath)
			}

			_, err = LoadPrivateKey(filePath)
			if tc.expectedError && err == nil {
				t.Errorf("Expected error, but got none")
			} else if !tc.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestLoadPublicKey(t *testing.T) {
	testCases := []struct {
		name          string
		generateKey   func() (*rsa.PublicKey, error)
		saveKey       func(key *rsa.PublicKey) (string, error)
		expectedError bool
	}{
		{
			name: "Valid Public Key",
			generateKey: func() (*rsa.PublicKey, error) {
				privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
				if err != nil {
					return nil, err
				}
				return &privateKey.PublicKey, nil
			},
			saveKey: func(key *rsa.PublicKey) (string, error) {
				tempFile, err := os.CreateTemp("", "public_key_test")
				if err != nil {
					return "", err
				}
				defer tempFile.Close()
				if err := SavePublicKey(tempFile.Name(), key); err != nil {
					return "", err
				}
				return tempFile.Name(), nil
			},
			expectedError: false,
		},
		{
			name: "Non-Existent File",
			generateKey: func() (*rsa.PublicKey, error) {
				return nil, nil
			},
			saveKey: func(key *rsa.PublicKey) (string, error) {
				return "non_existent_file.pem", nil
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var filePath string
			var err error

			key, err := tc.generateKey()
			if err != nil {
				t.Fatalf("Failed to generate public key: %v", err)
			}

			if tc.saveKey != nil {
				filePath, err = tc.saveKey(key)
				if err != nil {
					t.Fatalf("Failed to save public key: %v", err)
				}
				defer os.Remove(filePath)
			}

			_, err = LoadPublicKey(filePath)
			if tc.expectedError && err == nil {
				t.Errorf("Expected error, but got none")
			} else if !tc.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestEncryptWithPublicKey tests the EncryptWithPublicKey function.
func TestEncryptWithPublicKey(t *testing.T) {
	// Generate a private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Derive the public key from the private key
	publicKey := &privateKey.PublicKey

	// Define test cases
	testCases := []struct {
		name          string
		plaintext     string
		expectedError bool
	}{
		{
			name:          "Valid Plaintext",
			plaintext:     "Hello, World!",
			expectedError: false,
		},
		{
			name:          "Empty Plaintext",
			plaintext:     "",
			expectedError: false,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a buffer with the plaintext
			buff := bytes.NewBufferString(tc.plaintext)

			// Encrypt the content using the public key
			err := EncryptWithPublicKey(publicKey, buff)
			if err != nil && !tc.expectedError {
				t.Fatalf("Unexpected error: %v", err)
			} else if err == nil && tc.expectedError {
				t.Fatalf("Expected error, but got none")
			}

			// Decrypt the content using the private key
			if !tc.expectedError {
				ciphertext := buff.Bytes()
				decryptedBytes, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, ciphertext)
				if err != nil {
					t.Fatalf("Failed to decrypt ciphertext: %v", err)
				}

				// Compare the decrypted content with the original plaintext
				if string(decryptedBytes) != tc.plaintext {
					t.Errorf("Decrypted content does not match the original plaintext")
				}
			}
		})
	}
}
