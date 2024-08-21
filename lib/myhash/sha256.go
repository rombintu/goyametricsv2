package myhash

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/labstack/echo"
	"github.com/rombintu/goyametricsv2/internal/logger"
	"go.uber.org/zap"
)

func ToSHA256AndHMAC(src []byte, key string) string {
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write(src)
	return hex.EncodeToString(hash.Sum(nil))

	// data := append(src, []byte(key)...)
	// h := sha256.New()
	// h.Write(data)
	// return string(h.Sum(nil))
}

const (
	Sha256Header   = "HashSHA256"
	hashIsNotValid = "hash is not valid"
	hashIsValid    = "hash is valid"
	hashIsEmpty    = "hash is empty"
)

func HashCheckMiddleware(key string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip if KEY not set
			if key == "" {
				return next(c)
			}

			body, err := io.ReadAll(c.Request().Body)
			if err != nil {
				c.Error(err)
			}
			defer c.Request().Body.Close()
			// Check hash from request
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

			c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

			return next(c)
		}
	}
}
