// Package myhash provides utility functions for generating and validating SHA256 HMAC hashes.
// It also includes an Echo middleware for checking the integrity of request bodies using HMAC hashes.
package myhash

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rombintu/goyametricsv2/internal/logger"
	"go.uber.org/zap"
)

// ToSHA256AndHMAC generates a SHA256 HMAC hash for the given byte slice using the provided key.
//
// Parameters:
// - src: The byte slice to be hashed.
// - key: The secret key used for generating the HMAC hash.
//
// Returns:
// - The hexadecimal representation of the generated HMAC hash.
func ToSHA256AndHMAC(src []byte, key string) string {
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write(src)
	return hex.EncodeToString(hash.Sum(nil))
}

// Constants defining the SHA256 header and messages for hash validation.
const (
	Sha256Header   = "HashSHA256"
	hashIsNotValid = "hash is not valid"
	hashIsValid    = "hash is valid"
	hashIsEmpty    = "hash is empty"
)

// HashCheckMiddleware creates an Echo middleware function that checks the integrity of the request body
// by comparing the provided HMAC hash with the calculated hash.
//
// Parameters:
// - key: The secret key used for generating and validating the HMAC hash.
//
// Returns:
// - An Echo middleware function that wraps the next handler with hash validation.
func HashCheckMiddleware(key string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip hash validation if the key is not set or the hash header is empty
			if key == "" || c.Request().Header.Get(Sha256Header) == "" {
				return next(c)
			}

			// Read the request body
			body, err := io.ReadAll(c.Request().Body)
			if err != nil {
				c.Error(err)
			}
			defer c.Request().Body.Close()

			// Set the response content type header if the request content type is JSON
			if c.Request().Header.Get(echo.HeaderContentType) == echo.MIMEApplicationJSON {
				c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			}

			// Check the hash from the request header
			hashPayload := c.Request().Header.Get(Sha256Header)
			hashOriginal := ToSHA256AndHMAC(body, key)
			if hashPayload == "" {
				return c.String(http.StatusBadRequest, hashIsEmpty)
			} else if hashPayload != hashOriginal {
				logger.Log.Debug(hashIsNotValid, zap.String("payload", hashPayload), zap.String("original", hashOriginal))
				return c.String(http.StatusBadRequest, hashIsNotValid)
			} else {
				logger.Log.Debug(hashIsValid, zap.String("hash", hashPayload))
			}

			// Replace the request body with the original body for further processing
			c.Request().Body = io.NopCloser(bytes.NewBuffer(body))
			return next(c)
		}
	}
}
