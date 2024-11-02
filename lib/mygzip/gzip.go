// Package mygzip gzipMiddleware
package mygzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// Constants defining the gzip header value.
const (
	GzipHeader = "gzip"
)

// gzipResponseWriter is a custom response writer that wraps the standard http.ResponseWriter
// and adds a gzip writer to compress the response content.
type gzipResponseWriter struct {
	http.ResponseWriter           // Embed the standard http.ResponseWriter
	Writer              io.Writer // The gzip writer used to compress the response content
}

// Write overrides the Write method of the standard http.ResponseWriter.
// It sets the content type header if not already set and writes the data to the gzip writer.
//
// Parameters:
// - b: The byte slice to be written.
//
// Returns:
// - The number of bytes written and any error encountered.
func (grw *gzipResponseWriter) Write(b []byte) (int, error) {
	// Set the content type header if not already set
	if grw.Header().Get(echo.HeaderContentType) == "" {
		grw.Header().Set(echo.HeaderContentType, http.DetectContentType(b))
	}
	// Write the data to the gzip writer
	return grw.Writer.Write(b)
}

// GzipMiddleware is an Echo middleware that compresses the response content using gzip
// if the client supports gzip encoding. It also handles decompressing the request body
// if it is gzip encoded.
//
// Parameters:
// - next: The next handler function in the middleware chain.
//
// Returns:
// - An Echo handler function that wraps the next handler with gzip compression.
func GzipMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Check if the client accepts gzip encoding
		acceptEncoding := c.Request().Header.Get(echo.HeaderAcceptEncoding)
		if !strings.Contains(acceptEncoding, GzipHeader) {
			return next(c)
		}

		// Check if the request body is gzip encoded
		contentEncoding := c.Request().Header.Get(echo.HeaderContentEncoding)
		if strings.Contains(contentEncoding, GzipHeader) {
			// Create a gzip reader for the request body
			gzipReader, err := gzip.NewReader(c.Request().Body)
			if err != nil {
				return err
			}
			defer gzipReader.Close()

			// Replace the request body with the gzip reader
			c.Request().Body = io.NopCloser(gzipReader)
		}

		// Set the response content encoding header to gzip
		c.Response().Header().Set(echo.HeaderContentEncoding, GzipHeader)

		// Create a gzip writer for the response
		gzipWriter := gzip.NewWriter(c.Response().Writer)
		defer gzipWriter.Close()

		// Replace the response writer with a gzip response writer
		grw := &gzipResponseWriter{
			Writer:         gzipWriter,
			ResponseWriter: c.Response().Writer,
		}
		c.Response().Writer = grw

		// Call the next handler in the middleware chain
		if err := next(c); err != nil {
			c.Error(err)
		}

		// Flush the gzip writer to ensure all data is written
		if err := gzipWriter.Flush(); err != nil {
			return err
		}

		return nil
	}
}
