// Package mycrypt provides utility functions for RSA key generation, validation, and encryption/decryption.
package mycrypt

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/rombintu/goyametricsv2/internal/logger"
	"github.com/rombintu/goyametricsv2/lib/common"
	"go.uber.org/zap"
)

// ValidPrivateKey checks if the private key file at the specified path is valid.
// It returns true if the file exists, is a valid RSA private key, and has a length of 4096 bits.
func ValidPrivateKey(filePath string) bool {
	if !common.FileIsExists(filePath) {
		logger.Log.Debug("file not exists", zap.String("file", filePath))
		return false
	}

	// Read the file content
	keyBytes, err := os.ReadFile(filePath)
	if err != nil {
		logger.Log.Debug("error read file", zap.String("file", filePath))
		return false
	}

	// Decode the PEM block
	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		logger.Log.Debug("error decode file", zap.String("file", filePath))
		return false
	}

	// Parse the private key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		logger.Log.Debug("error parse private key", zap.String("file", filePath))
		return false
	}

	// Check the key length
	if privateKey.N.BitLen() != 4096 {
		logger.Log.Debug("error private key len != 4096", zap.String("file", filePath))
		return false
	}

	// Validate the key
	if err := privateKey.Validate(); err != nil {
		logger.Log.Debug("invalid RSA private key", zap.String("file", filePath))
		return false
	}

	return true
}

// GenRSAKeyPair generates a new RSA key pair with a 4096-bit key length and saves it to the specified files.
// It returns the generated private key or an error if the operation fails.
func GenRSAKeyPair(filename string) (*rsa.PrivateKey, error) {
	// Generate the private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	// Derive the public key from the private key
	publicKey := &privateKey.PublicKey

	if err := SavePrivateKey(filename, privateKey); err != nil {
		return nil, err
	}
	if err := SavePublicKey(filename+".pub", publicKey); err != nil {
		return nil, err
	}

	return privateKey, nil
}

// SavePrivateKey saves the given RSA private key to the specified file in PEM format.
func SavePrivateKey(filename string, key *rsa.PrivateKey) error {
	// Serialize the private key to PKCS#1, ASN.1 DER format
	keyBytes := x509.MarshalPKCS1PrivateKey(key)

	// Create a PEM block
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}

	// Save to file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, block)
}

// SavePublicKey saves the given RSA public key to the specified file in PEM format.
func SavePublicKey(filename string, key *rsa.PublicKey) error {
	keyBytes := x509.MarshalPKCS1PublicKey(key)

	// Create a PEM block
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: keyBytes,
	}

	// Save to file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, block)
}

// LoadPrivateKey loads an RSA private key from the specified file.
func LoadPrivateKey(filename string) (*rsa.PrivateKey, error) {
	privateKeyBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(privateKeyBytes)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

// LoadPublicKey loads an RSA public key from the specified file.
func LoadPublicKey(filename string) (*rsa.PublicKey, error) {
	publicKeyBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(publicKeyBytes)
	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return publicKey, nil
}

// EncryptMiddleware is an Echo middleware that decrypts the request body using the specified private key file.
// It only decrypts POST and PUT requests.
func EncryptMiddleware(privateKeyFile string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Decrypt the request body for POST and PUT methods
			if c.Request().Method == http.MethodPost || c.Request().Method == http.MethodPut {
				body, err := io.ReadAll(c.Request().Body)
				if err != nil {
					return err
				}

				// Load the private key from the file
				privateKey, err := LoadPrivateKey(privateKeyFile)
				if err != nil {
					return err
				}
				decryptedBytes, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, body)
				if err != nil {
					return err
				}
				c.Request().Body = io.NopCloser(bytes.NewReader(decryptedBytes))
			}

			// Pass control to the next handler
			if err := next(c); err != nil {
				c.Error(err)
			}

			return nil
		}
	}
}

// EncryptWithPublicKey encrypts the content of the given buffer using the specified public key.
// The encrypted content is then written back to the buffer.
func EncryptWithPublicKey(pubKey *rsa.PublicKey, buff *bytes.Buffer) error {
	// Read the buffer content
	plaintext := buff.Bytes()

	// Encrypt the content using the public key
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, plaintext)
	if err != nil {
		return err
	}

	// Clear the buffer and write the encrypted message to the buffer
	buff.Reset()
	_, err = buff.Write(ciphertext)
	if err != nil {
		return err
	}

	return nil
}
