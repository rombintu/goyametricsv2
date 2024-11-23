// Package myhash provides utility functions for generating and validating SHA256 HMAC hashes.
// It also includes an Echo middleware for checking the integrity of request bodies using HMAC hashes.
package myhash

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

const (
	testKey     = "secret"
	testPayload = `{"name":"John","age":30}`
	testHash    = "5e884898da28047151d0e56f8dc6292773603d0d6aabbddc420072026e112a6c"
)

func TestToSHA256AndHMAC(t *testing.T) {
	type args struct {
		src []byte
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "simple_bytes_2_hmac",
			args: args{src: []byte("hello"), key: "secret-key"},
			want: "98e7ffb964bb5a3f902db1fc101a5baa98b6f2cd56858210c9d70f26ac762fc7",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSHA256AndHMAC(tt.args.src, tt.args.key); got != tt.want {
				t.Errorf("ToSHA256AndHMAC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHashCheckMiddleware(t *testing.T) {
	e := echo.New()

	// Helper function to create a test handler
	createTestHandler := func(t *testing.T, expectedStatus int, expectedBody string) echo.HandlerFunc {
		return func(c echo.Context) error {
			assert.Equal(t, expectedStatus, c.Response().Status)
			assert.Equal(t, expectedBody, c.Response().Header().Get(echo.HeaderContentType))
			return nil
		}
	}

	// Helper function to create a test request
	createTestRequest := func(payload, hash string) *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		if hash != "" {
			req.Header.Set(Sha256Header, hash)
		}
		return req
	}

	// Test cases
	tests := []struct {
		name           string
		key            string
		payload        string
		hash           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "No Key Provided",
			key:            "",
			payload:        testPayload,
			hash:           testHash,
			expectedStatus: 0,
			expectedBody:   "",
		},
		{
			name:           "No Hash Provided",
			key:            testKey,
			payload:        testPayload,
			hash:           "",
			expectedStatus: 0,
			expectedBody:   "",
		},
		{
			name:           "Invalid Hash",
			key:            testKey,
			payload:        testPayload,
			hash:           "invalidhash",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   hashIsNotValid,
		},
		{
			name:           "Valid Hash",
			key:            testKey,
			payload:        testPayload,
			hash:           testHash,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(tt.payload, tt.hash)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			middleware := HashCheckMiddleware(tt.key)
			handler := middleware(createTestHandler(t, tt.expectedStatus, tt.expectedBody))

			err := handler(c)
			assert.NoError(t, err)

			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rec.Body.String())
			}
		})
	}
}
